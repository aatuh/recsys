package recsysvc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aatuh/api-toolkit/authorization"
	"github.com/aatuh/recsys-suite/api/recsys-algo/algorithm"
	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"
	"github.com/aatuh/recsys-suite/api/recsys-algo/rules"
	"github.com/google/uuid"
)

// AlgoEngineConfig wires the recsys-algo engine into the service layer.
type AlgoEngineConfig struct {
	Version          string
	DefaultNamespace string
	AlgorithmConfig  algorithm.Config
}

// AlgoEngine adapts recsys-algo to the service Engine port.
type AlgoEngine struct {
	engine           *algorithm.Engine
	similar          *algorithm.SimilarItemsEngine
	store            recmodel.EngineStore
	version          string
	defaultNamespace string
}

// NewAlgoEngine constructs a new adapter backed by recsys-algo.
func NewAlgoEngine(cfg AlgoEngineConfig, store recmodel.EngineStore, rulesManager *rules.Manager) *AlgoEngine {
	if store == nil {
		store = noopAlgoStore{}
	}
	algo := algorithm.NewEngine(cfg.AlgorithmConfig, store, rulesManager)
	similar := algorithm.NewSimilarItemsEngine(store, cfg.AlgorithmConfig.CoVisWindowDays)
	version := strings.TrimSpace(cfg.Version)
	if version == "" {
		version = defaultAlgoVersion
	}
	ns := strings.TrimSpace(cfg.DefaultNamespace)
	if ns == "" {
		ns = "default"
	}
	return &AlgoEngine{
		engine:           algo,
		similar:          similar,
		store:            store,
		version:          version,
		defaultNamespace: ns,
	}
}

// Version returns the configured algorithm version label.
func (e *AlgoEngine) Version() string {
	if e == nil || strings.TrimSpace(e.version) == "" {
		return defaultAlgoVersion
	}
	return e.version
}

// Recommend runs recsys-algo for the normalized request.
func (e *AlgoEngine) Recommend(ctx context.Context, req RecommendRequest) ([]Item, []Warning, error) {
	if e == nil || e.engine == nil {
		return nil, nil, nil
	}
	algoReq := e.mapRecommendRequest(ctx, req)
	resp, trace, err := e.engine.Recommend(ctx, algoReq)
	if err != nil {
		return nil, nil, err
	}
	items := e.mapItems(resp.Items, req.Options.Explain, req.Options.IncludeReasons)
	warnings := e.warningsFromTrace(trace)
	adjusted, extra := e.applyPinnedOverrides(req, items, trace)
	warnings = append(warnings, extra...)
	filtered, extra := e.applyPostConstraints(ctx, algoReq, req, adjusted)
	warnings = append(warnings, extra...)
	filtered, extra = e.applyCandidateAllowList(req, filtered)
	warnings = append(warnings, extra...)
	return filtered, warnings, nil
}

// Similar runs the similar-items engine.
func (e *AlgoEngine) Similar(ctx context.Context, req SimilarRequest) ([]Item, []Warning, error) {
	if e == nil || e.similar == nil {
		return nil, nil, nil
	}
	algoReq := algorithm.SimilarItemsRequest{
		OrgID:     e.orgIDFromContext(ctx),
		ItemID:    strings.TrimSpace(req.ItemID),
		Namespace: e.namespaceFor(req.Surface),
		K:         req.K,
	}
	resp, err := e.similar.FindSimilar(ctx, algoReq)
	if err != nil {
		if errors.Is(err, recmodel.ErrFeatureUnavailable) {
			return nil, []Warning{{Code: "SIGNAL_UNAVAILABLE", Detail: "similarity signals unavailable"}}, nil
		}
		return nil, nil, err
	}
	items := e.mapItems(resp.Items, req.Options.Explain, req.Options.IncludeReasons)
	return items, nil, nil
}

func (e *AlgoEngine) mapRecommendRequest(ctx context.Context, req RecommendRequest) algorithm.Request {
	anchors := collectAnchorIDs(req)
	algoReq := algorithm.Request{
		OrgID:          e.orgIDFromContext(ctx),
		UserID:         chooseUserID(req.User),
		Namespace:      e.namespaceFor(req.Surface),
		Surface:        req.Surface,
		SegmentID:      req.Segment,
		K:              req.K,
		Constraints:    buildPopConstraints(req),
		Blend:          mapWeights(req.Weights),
		IncludeReasons: req.Options.IncludeReasons,
		ExplainLevel:   mapExplainLevel(req.Options.Explain),
		InjectAnchors:  len(anchors) > 0,
		AnchorItemIDs:  anchors,
	}
	return algoReq
}

func (e *AlgoEngine) mapItems(items []algorithm.ScoredItem, explain string, includeReasons bool) []Item {
	if len(items) == 0 {
		return nil
	}
	out := make([]Item, len(items))
	for i, it := range items {
		out[i] = Item{
			ItemID: it.ItemID,
			Score:  it.Score,
		}
		if includeReasons && len(it.Reasons) > 0 {
			out[i].Reasons = append([]string(nil), it.Reasons...)
		}
		if strings.EqualFold(explain, "none") {
			continue
		}
		explainBlock := mapExplainBlock(it.Explain, it.Reasons)
		if explainBlock != nil {
			out[i].Explain = explainBlock
		}
	}
	return out
}

func mapExplainLevel(raw string) algorithm.ExplainLevel {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "full":
		return algorithm.ExplainLevelFull
	case "summary":
		return algorithm.ExplainLevelNumeric
	default:
		return algorithm.ExplainLevelTags
	}
}

func mapExplainBlock(block *algorithm.ExplainBlock, reasons []string) *ItemExplain {
	if block == nil {
		return nil
	}
	signals := make(map[string]float64)
	if block.Blend != nil {
		signals["pop"] = block.Blend.Contributions.Pop
		signals["cooc"] = block.Blend.Contributions.Cooc
		signals["emb"] = block.Blend.Contributions.Similarity
		if signals["pop"] == 0 && signals["cooc"] == 0 && signals["emb"] == 0 {
			signals["pop"] = block.Blend.PopNorm
			signals["cooc"] = block.Blend.CoocNorm
			signals["emb"] = block.Blend.SimilarityNorm
		}
	}
	rules := extractRuleReasons(reasons)
	if len(signals) == 0 && len(rules) == 0 {
		return nil
	}
	return &ItemExplain{Signals: signals, Rules: rules}
}

func extractRuleReasons(reasons []string) []string {
	if len(reasons) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(reasons))
	out := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		if !strings.HasPrefix(reason, "rule.") {
			continue
		}
		if _, ok := seen[reason]; ok {
			continue
		}
		seen[reason] = struct{}{}
		out = append(out, reason)
	}
	return out
}

func (e *AlgoEngine) warningsFromTrace(trace *algorithm.TraceData) []Warning {
	if trace == nil || len(trace.SignalStatus) == 0 {
		return nil
	}
	out := make([]Warning, 0, len(trace.SignalStatus))
	for signal, status := range trace.SignalStatus {
		if status.Available && !status.Partial && strings.TrimSpace(status.Error) == "" {
			continue
		}
		code := "SIGNAL_UNAVAILABLE"
		if status.Partial {
			code = "SIGNAL_PARTIAL"
		}
		detail := fmt.Sprintf("%s signal unavailable", signal)
		if strings.TrimSpace(status.Error) != "" {
			detail = fmt.Sprintf("%s unavailable: %s", signal, status.Error)
		}
		out = append(out, Warning{Code: code, Detail: detail})
	}
	return out
}

func (e *AlgoEngine) applyPostConstraints(
	ctx context.Context,
	algoReq algorithm.Request,
	req RecommendRequest,
	items []Item,
) ([]Item, []Warning) {
	if len(items) == 0 || req.Constraints == nil {
		return items, nil
	}
	tagStore, ok := e.store.(recmodel.TagStore)
	if !ok {
		return items, nil
	}
	forbidden := normalizeTags(req.Constraints.ForbiddenTags)
	maxPerTag := normalizeTagLimits(req.Constraints.MaxPerTag)
	if len(forbidden) == 0 && len(maxPerTag) == 0 {
		return items, nil
	}
	itemIDs := make([]string, 0, len(items))
	for _, item := range items {
		if item.ItemID != "" {
			itemIDs = append(itemIDs, item.ItemID)
		}
	}
	tags, err := tagStore.ListItemsTags(ctx, algoReq.OrgID, algoReq.Namespace, itemIDs)
	if err != nil {
		return items, []Warning{{Code: "TAG_LOOKUP_FAILED", Detail: "failed to fetch item tags for constraints"}}
	}
	counts := make(map[string]int)
	filtered := make([]Item, 0, len(items))
	filteredCount := 0
	for _, item := range items {
		if item.ItemID == "" {
			continue
		}
		info := tags[item.ItemID]
		itemTags := recmodel.NormalizeTags(info.Tags)
		if violatesForbidden(itemTags, forbidden) {
			filteredCount++
			continue
		}
		if violatesMaxPerTag(itemTags, counts, maxPerTag) {
			filteredCount++
			continue
		}
		filtered = append(filtered, item)
	}
	if filteredCount == 0 {
		return items, nil
	}
	return filtered, []Warning{{Code: "CONSTRAINTS_FILTERED", Detail: fmt.Sprintf("%d items removed by tag constraints", filteredCount)}}
}

func (e *AlgoEngine) applyPinnedOverrides(req RecommendRequest, items []Item, trace *algorithm.TraceData) ([]Item, []Warning) {
	if trace == nil || len(trace.RulePinned) == 0 {
		return items, nil
	}
	itemByID := make(map[string]Item, len(items))
	for _, item := range items {
		if item.ItemID == "" {
			continue
		}
		itemByID[item.ItemID] = item
	}
	pinned := make([]Item, 0, len(trace.RulePinned))
	pinnedSet := make(map[string]struct{}, len(trace.RulePinned))
	injected := 0
	for _, pin := range trace.RulePinned {
		id := strings.TrimSpace(pin.ItemID)
		if id == "" {
			continue
		}
		if _, seen := pinnedSet[id]; seen {
			continue
		}
		pinnedSet[id] = struct{}{}
		if item, ok := itemByID[id]; ok {
			pinned = append(pinned, item)
			continue
		}
		injected++
		pinned = append(pinned, Item{
			ItemID: id,
			Score:  pin.Score,
		})
	}
	if len(pinned) == 0 {
		return items, nil
	}
	rest := make([]Item, 0, len(items))
	for _, item := range items {
		if _, ok := pinnedSet[item.ItemID]; ok {
			continue
		}
		rest = append(rest, item)
	}
	combined := append(pinned, rest...)
	if req.K > 0 && len(combined) > req.K {
		combined = combined[:req.K]
	}
	if injected == 0 {
		return combined, nil
	}
	return combined, []Warning{{Code: "RULE_PIN_INJECTED", Detail: fmt.Sprintf("%d pinned items were injected into results", injected)}}
}

func (e *AlgoEngine) applyCandidateAllowList(req RecommendRequest, items []Item) ([]Item, []Warning) {
	if req.Candidates == nil || len(req.Candidates.IncludeIDs) == 0 {
		return items, nil
	}
	allowed := normalizeList(req.Candidates.IncludeIDs, false)
	if len(allowed) == 0 {
		return items, nil
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, id := range allowed {
		allowedSet[id] = struct{}{}
	}
	filtered := make([]Item, 0, len(items))
	removed := 0
	for _, item := range items {
		if _, ok := allowedSet[item.ItemID]; ok {
			filtered = append(filtered, item)
			continue
		}
		removed++
	}
	if removed == 0 {
		return items, nil
	}
	if len(filtered) == 0 {
		return filtered, []Warning{{Code: "CANDIDATES_INCLUDE_EMPTY", Detail: "no items matched candidates.include_ids"}}
	}
	return filtered, []Warning{{Code: "CANDIDATES_INCLUDE_FILTERED", Detail: fmt.Sprintf("%d items removed by candidates.include_ids", removed)}}
}

func buildPopConstraints(req RecommendRequest) *recmodel.PopConstraints {
	if req.Constraints == nil && (req.Candidates == nil || len(req.Candidates.ExcludeIDs) == 0) {
		return nil
	}
	constraints := recmodel.PopConstraints{}
	if req.Constraints != nil {
		constraints.IncludeTagsAny = normalizeTags(req.Constraints.RequiredTags)
	}
	if req.Candidates != nil && len(req.Candidates.ExcludeIDs) > 0 {
		constraints.ExcludeItemIDs = normalizeList(req.Candidates.ExcludeIDs, false)
	}
	return &constraints
}

func mapWeights(weights *Weights) *algorithm.BlendWeights {
	if weights == nil {
		return nil
	}
	return &algorithm.BlendWeights{
		Pop:        weights.Pop,
		Cooc:       weights.Cooc,
		Similarity: weights.Emb,
	}
}

func collectAnchorIDs(req RecommendRequest) []string {
	var anchors []string
	if req.Anchors != nil {
		anchors = append(anchors, req.Anchors.ItemIDs...)
		if req.Anchors.MaxAnchors > 0 && len(anchors) > req.Anchors.MaxAnchors {
			anchors = anchors[:req.Anchors.MaxAnchors]
		}
	}
	if req.Candidates != nil && len(req.Candidates.IncludeIDs) > 0 {
		anchors = append(anchors, req.Candidates.IncludeIDs...)
	}
	return normalizeList(anchors, false)
}

func chooseUserID(user UserRef) string {
	if v := strings.TrimSpace(user.UserID); v != "" {
		return v
	}
	if v := strings.TrimSpace(user.SessionID); v != "" {
		return v
	}
	return strings.TrimSpace(user.AnonymousID)
}

func (e *AlgoEngine) namespaceFor(surface string) string {
	if ns := strings.TrimSpace(surface); ns != "" {
		return ns
	}
	return e.defaultNamespace
}

func (e *AlgoEngine) orgIDFromContext(ctx context.Context) uuid.UUID {
	tenant := ""
	if ctx != nil {
		if v, ok := authorization.TenantIDFromContext(ctx); ok {
			tenant = v
		}
	}
	return tenantToUUID(tenant)
}

func tenantToUUID(tenant string) uuid.UUID {
	tenant = strings.TrimSpace(tenant)
	if tenant == "" {
		return uuid.Nil
	}
	if id, err := uuid.Parse(tenant); err == nil {
		return id
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(tenant))
}

func normalizeTags(tags []string) []string {
	return recmodel.NormalizeTags(tags)
}

func normalizeList(values []string, lower bool) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		item := strings.TrimSpace(v)
		if lower {
			item = strings.ToLower(item)
		}
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func normalizeTagLimits(limits map[string]int) map[string]int {
	if len(limits) == 0 {
		return nil
	}
	out := make(map[string]int, len(limits))
	for tag, limit := range limits {
		tag = recmodel.NormalizeTag(tag)
		if tag == "" || limit <= 0 {
			continue
		}
		out[tag] = limit
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func violatesForbidden(tags []string, forbidden []string) bool {
	if len(tags) == 0 || len(forbidden) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		seen[tag] = struct{}{}
	}
	for _, tag := range forbidden {
		if _, ok := seen[tag]; ok {
			return true
		}
	}
	return false
}

func violatesMaxPerTag(tags []string, counts map[string]int, limits map[string]int) bool {
	if len(tags) == 0 || len(limits) == 0 {
		return false
	}
	for _, tag := range tags {
		limit, ok := limits[tag]
		if !ok {
			continue
		}
		if counts[tag]+1 > limit {
			return true
		}
	}
	for _, tag := range tags {
		if _, ok := limits[tag]; ok {
			counts[tag]++
		}
	}
	return false
}

// noopAlgoStore is used when no backing store is provided.
type noopAlgoStore struct{}

func (noopAlgoStore) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int, c *recmodel.PopConstraints) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (noopAlgoStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]recmodel.ItemTags, error) {
	return map[string]recmodel.ItemTags{}, nil
}

var _ recmodel.EngineStore = noopAlgoStore{}
