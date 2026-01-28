package recommendation

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/aatuh/recsys-algo/rules"
	"recsys/internal/types"
	spectypes "recsys/specs/types"

	"github.com/aatuh/recsys-algo/algorithm"

	recmodel "github.com/aatuh/recsys-algo/model"

	"github.com/google/uuid"
)

// ValidationError represents client-side validation failures.
type ValidationError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// Store defines the persistence contract needed by the service.
type Store interface {
	recmodel.EngineStore
	recmodel.AvailabilityStore
	recmodel.HistoryStore
}

// SegmentSelection describes the resolved segment context.
type SegmentSelection struct {
	Profile     *types.SegmentProfile
	SegmentID   string
	ProfileID   string
	RuleID      int64
	UserTraits  map[string]any
	UserCreated time.Time
}

// SegmentSelector resolves segment information prior to ranking.
type SegmentSelector func(ctx context.Context, req algorithm.Request, httpReq spectypes.RecommendRequest) (SegmentSelection, error)

// Result encapsulates recommendation outputs for the HTTP layer.
type Result struct {
	Response     spectypes.RecommendResponse
	AlgoRequest  algorithm.Request
	AlgoConfig   algorithm.Config
	AlgoResponse *algorithm.Response
	TraceData    *algorithm.TraceData
	SourceStats  map[string]algorithm.SourceMetric
}

// Service orchestrates recommendation requests.
type Service struct {
	store                 Store
	rules                 *rules.Manager
	blendResolver         BlendConfigResolver
	segmentBlendOverrides map[string]ResolvedBlendConfig
	clock                 algorithm.Clock
	signalObserver        algorithm.SignalObserver
	newUserBlendAlpha     *float64
	newUserBlendBeta      *float64
	newUserBlendGamma     *float64
	newUserMMRLambda      *float64
	newUserPopFanout      *int
	starterPresets        map[string]map[string]float64
	starterDecayEvents    int
}

// New constructs a recommendation service.
func New(store Store, rulesManager *rules.Manager) *Service {
	return &Service{store: store, rules: rulesManager, starterPresets: defaultStarterPresets(), starterDecayEvents: 5}
}

// WithBlendResolver configures a resolver for runtime blend overrides.
func (s *Service) WithBlendResolver(resolver BlendConfigResolver) *Service {
	s.blendResolver = resolver
	return s
}

// WithClock injects a clock used for deterministic algorithm behavior.
func (s *Service) WithClock(clock algorithm.Clock) *Service {
	s.clock = clock
	return s
}

// WithSignalObserver registers an observer for signal telemetry.
func (s *Service) WithSignalObserver(observer algorithm.SignalObserver) *Service {
	s.signalObserver = observer
	return s
}

// WithSegmentBlendOverrides configures static blend weights per segment name.
func (s *Service) WithSegmentBlendOverrides(overrides map[string]ResolvedBlendConfig) *Service {
	if len(overrides) == 0 {
		s.segmentBlendOverrides = nil
		return s
	}
	if s.segmentBlendOverrides == nil {
		s.segmentBlendOverrides = make(map[string]ResolvedBlendConfig, len(overrides))
	}
	for segment, cfg := range overrides {
		key := strings.ToLower(strings.TrimSpace(segment))
		if key == "" {
			continue
		}
		s.segmentBlendOverrides[key] = cfg
	}
	return s
}

// WithNewUserOverrides configures blend/MMR overrides for new or sparse-history users.
func (s *Service) WithNewUserOverrides(alpha, beta, gamma, mmr *float64, popFanout *int) *Service {
	s.newUserBlendAlpha = alpha
	s.newUserBlendBeta = beta
	s.newUserBlendGamma = gamma
	s.newUserMMRLambda = mmr
	s.newUserPopFanout = popFanout
	return s
}

// WithStarterProfiles configures starter-profile presets and decay behavior.
func (s *Service) WithStarterProfiles(presets map[string]map[string]float64, decayEvents int) *Service {
	if len(presets) == 0 {
		presets = defaultStarterPresets()
	} else {
		// defensive copy
		copied := make(map[string]map[string]float64, len(presets))
		for segment, weights := range presets {
			if len(weights) == 0 {
				continue
			}
			inner := make(map[string]float64, len(weights))
			for key, weight := range weights {
				inner[strings.ToLower(strings.TrimSpace(key))] = weight
			}
			copied[strings.ToLower(strings.TrimSpace(segment))] = inner
		}
		if len(copied) == 0 {
			copied = defaultStarterPresets()
		}
		presets = copied
	}
	if decayEvents <= 0 {
		decayEvents = 5
	}
	s.starterPresets = presets
	s.starterDecayEvents = decayEvents
	return s
}

// Recommend executes the ranking pipeline using the provided context.
func (s *Service) Recommend(
	ctx context.Context,
	orgID uuid.UUID,
	req spectypes.RecommendRequest,
	baseCfg algorithm.Config,
	selector SegmentSelector,
) (*Result, error) {
	algoReq, cfg, selection, err := s.prepareRecommendInputs(ctx, orgID, req, baseCfg, selector)
	if err != nil {
		return nil, err
	}

	engine := algorithm.NewEngine(
		cfg,
		s.store,
		s.rules,
		algorithm.WithClock(s.clock),
		algorithm.WithSignalObserver(s.signalObserver),
	)
	algoResp, traceData, err := engine.Recommend(ctx, algoReq)
	if err != nil {
		return nil, err
	}

	if selection.SegmentID != "" {
		algoResp.SegmentID = selection.SegmentID
	}
	if selection.ProfileID != "" {
		algoResp.ProfileID = selection.ProfileID
	}

	httpResp := convertToHTTPResponse(algoResp)

	return &Result{
		Response:     httpResp,
		AlgoRequest:  algoReq,
		AlgoConfig:   cfg,
		AlgoResponse: algoResp,
		TraceData:    traceData,
		SourceStats:  traceData.SourceMetrics,
	}, nil
}

// Rerank applies the blended scoring pipeline to a provided candidate set.
func (s *Service) Rerank(
	ctx context.Context,
	orgID uuid.UUID,
	req spectypes.RerankRequest,
	baseCfg algorithm.Config,
	selector SegmentSelector,
) (*Result, error) {
	if len(req.Items) == 0 {
		return nil, ValidationError{Code: "empty_items", Message: "items must not be empty"}
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if req.K <= 0 || req.K > len(req.Items) {
		req.K = len(req.Items)
	}

	converted := rerankAsRecommend(req)
	algoReq, cfg, selection, err := s.prepareRecommendInputs(ctx, orgID, converted, baseCfg, selector)
	if err != nil {
		return nil, err
	}

	candidates, err := buildPrefetchedCandidates(req.Items)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, ValidationError{Code: "invalid_items", Message: "no valid candidate IDs provided"}
	}
	algoReq.PrefetchedCandidates = candidates
	if algoReq.K <= 0 || algoReq.K > len(candidates) {
		algoReq.K = len(candidates)
	}

	engine := algorithm.NewEngine(
		cfg,
		s.store,
		s.rules,
		algorithm.WithClock(s.clock),
		algorithm.WithSignalObserver(s.signalObserver),
	)
	algoResp, traceData, err := engine.Recommend(ctx, algoReq)
	if err != nil {
		return nil, err
	}

	if selection.SegmentID != "" {
		algoResp.SegmentID = selection.SegmentID
	}
	if selection.ProfileID != "" {
		algoResp.ProfileID = selection.ProfileID
	}

	httpResp := convertToHTTPResponse(algoResp)

	return &Result{
		Response:     httpResp,
		AlgoRequest:  algoReq,
		AlgoConfig:   cfg,
		AlgoResponse: algoResp,
		TraceData:    traceData,
		SourceStats:  traceData.SourceMetrics,
	}, nil
}

func (s *Service) prepareRecommendInputs(
	ctx context.Context,
	orgID uuid.UUID,
	req spectypes.RecommendRequest,
	baseCfg algorithm.Config,
	selector SegmentSelector,
) (algorithm.Request, algorithm.Config, SegmentSelection, error) {
	algoReq, err := buildAlgorithmRequest(orgID, req)
	if err != nil {
		return algorithm.Request{}, algorithm.Config{}, SegmentSelection{}, err
	}
	cfg := baseCfg

	var selection SegmentSelection
	if selector != nil {
		sel, err := selector(ctx, algoReq, req)
		if err != nil {
			return algorithm.Request{}, algorithm.Config{}, SegmentSelection{}, err
		}
		selection = sel
		if selection.Profile != nil {
			applySegmentProfile(&cfg, *selection.Profile)
			if selection.ProfileID == "" {
				selection.ProfileID = selection.Profile.ProfileID
			}
		}
		algoReq.SegmentID = selection.SegmentID
	}

	segmentHint := strings.TrimSpace(selection.SegmentID)
	if segmentHint == "" && selection.UserTraits != nil {
		if seg, ok := selection.UserTraits["segment"].(string); ok {
			segmentHint = seg
		}
	}
	isNewSegment := strings.EqualFold(segmentHint, "new_users")
	starterSelection := selection
	if starterSelection.SegmentID == "" && segmentHint != "" {
		starterSelection.SegmentID = segmentHint
	}

	blendOverrides(&cfg, s.resolveBlend(ctx, algoReq.Namespace))
	s.applySegmentBlendOverrides(&cfg, segmentHint)
	applyOverrides(&cfg, req.Overrides)

	recentItems, err := s.recentInteractionItems(ctx, cfg, algoReq)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return algorithm.Request{}, algorithm.Config{}, SegmentSelection{}, err
		}
		recentItems = nil
	}

	recentEventCount := len(recentItems)
	recentCountKnown := recentItems != nil
	minEvents := cfg.ProfileMinEventsForBoost
	if minEvents < 0 {
		minEvents = 0
	}
	isSparseHistory := false
	if !recentCountKnown {
		isSparseHistory = true
	} else if minEvents == 0 {
		isSparseHistory = recentEventCount == 0
	} else {
		isSparseHistory = recentEventCount < minEvents
	}

	if isNewSegment {
		cfg.RuleExcludeEvents = false
		if fanout := s.newUserPopFanout; fanout != nil && *fanout > 0 {
			cfg.PopularityFanout = *fanout
		} else if cfg.PopularityFanout < 1000 {
			cfg.PopularityFanout = 1000
		}
	}

	if isNewSegment || isSparseHistory {
		if fanout := s.newUserPopFanout; fanout != nil && *fanout > 0 {
			cfg.PopularityFanout = *fanout
		} else if cfg.PopularityFanout < 1000 {
			cfg.PopularityFanout = 1000
		}
	}

	if isNewSegment || isSparseHistory {
		if weight := s.newUserBlendAlpha; weight != nil {
			cfg.BlendAlpha = *weight
		}
		if weight := s.newUserBlendBeta; weight != nil {
			cfg.BlendBeta = *weight
		}
		if weight := s.newUserBlendGamma; weight != nil {
			cfg.BlendGamma = *weight
		}
		if lambda := s.newUserMMRLambda; lambda != nil {
			cfg.MMRLambda = *lambda
		}
	}

	effectiveEventCount := recentEventCount
	if !recentCountKnown {
		effectiveEventCount = -1
	}
	if (isNewSegment || isSparseHistory) && cfg.ProfileMinEventsForBoost > 0 && effectiveEventCount >= cfg.ProfileMinEventsForBoost {
		effectiveEventCount = cfg.ProfileMinEventsForBoost - 1
		if effectiveEventCount < 0 {
			effectiveEventCount = 0
		}
	}
	algoReq.RecentEventCount = effectiveEventCount
	algoReq.InjectAnchors = isNewSegment || isSparseHistory
	if len(recentItems) > 0 {
		var anchors []string
		if s.store != nil {
			if avail, err := s.store.ListItemsAvailability(ctx, algoReq.OrgID, algoReq.Namespace, recentItems); err == nil {
				anchors = filterAnchorsByAvailability(recentItems, avail)
			} else {
				anchors = nil
			}
		} else {
			anchors = filterAnchorsByAvailability(recentItems, nil)
		}
		if len(anchors) > 0 {
			algoReq.AnchorItemIDs = anchors
		}
	}

	starterProfile, starterWeight := s.buildStarterProfile(
		ctx,
		cfg,
		algoReq,
		starterSelection,
		recentEventCount,
		recentCountKnown,
		recentItems,
	)
	if len(starterProfile) > 0 {
		algoReq.StarterProfile = starterProfile
		algoReq.StarterBlendWeight = starterWeight
	}
	if isNewSegment && len(algoReq.AnchorItemIDs) == 0 && len(starterProfile) > 0 {
		anchorLimit := algoReq.K
		if anchorLimit <= 0 {
			anchorLimit = 10
		}
		if anchorLimit > 5 {
			anchorLimit = 5
		}
		if anchors := s.starterAnchors(ctx, cfg, algoReq, starterProfile, anchorLimit); len(anchors) > 0 {
			algoReq.AnchorItemIDs = anchors
		}
	}

	if err := validateAlgorithmInputs(cfg, algoReq); err != nil {
		return algorithm.Request{}, algorithm.Config{}, SegmentSelection{}, err
	}

	return algoReq, cfg, selection, nil
}

func rerankAsRecommend(req spectypes.RerankRequest) spectypes.RecommendRequest {
	return spectypes.RecommendRequest{
		UserID:         req.UserID,
		Namespace:      req.Namespace,
		K:              req.K,
		Context:        req.Context,
		IncludeReasons: req.IncludeReasons,
		ExplainLevel:   req.ExplainLevel,
		Blend:          req.Blend,
		Overrides:      req.Overrides,
		Constraints:    req.Constraints,
	}
}

func buildPrefetchedCandidates(items []spectypes.RerankCandidate) ([]recmodel.ScoredItem, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]recmodel.ScoredItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		id := strings.TrimSpace(item.ItemID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		score := 0.0
		if item.Score != nil {
			score = *item.Score
		}
		out = append(out, recmodel.ScoredItem{ItemID: id, Score: score})
	}
	return out, nil
}

func validateAlgorithmInputs(cfg algorithm.Config, req algorithm.Request) error {
	var cfgErrs algorithm.ValidationErrors
	if err := cfg.Validate(); err != nil {
		if v, ok := err.(algorithm.ValidationErrors); ok {
			cfgErrs = v
		} else {
			return err
		}
	}

	var reqErrs algorithm.ValidationErrors
	if err := req.Validate(); err != nil {
		if v, ok := err.(algorithm.ValidationErrors); ok {
			reqErrs = v
		} else {
			return err
		}
	}

	if len(cfgErrs) == 0 && len(reqErrs) == 0 {
		return nil
	}

	details := make(map[string]any)
	if len(cfgErrs) > 0 {
		details["config"] = cfgErrs.Fields()
	}
	if len(reqErrs) > 0 {
		details["request"] = reqErrs.Fields()
	}

	code := "invalid_request"
	message := "invalid recommendation request"
	switch {
	case len(cfgErrs) > 0 && len(reqErrs) > 0:
		code = "invalid_input"
		message = "invalid recommendation input"
	case len(cfgErrs) > 0:
		code = "invalid_config"
		message = "invalid recommendation config"
	}
	return ValidationError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func buildAlgorithmRequest(orgID uuid.UUID, req spectypes.RecommendRequest) (algorithm.Request, error) {
	ns := req.Namespace
	if ns == "" {
		ns = "default"
	}

	constraints, err := parseConstraints(req.Constraints)
	if err != nil {
		return algorithm.Request{}, err
	}

	return algorithm.Request{
		OrgID:          orgID,
		UserID:         req.UserID,
		Namespace:      ns,
		K:              req.K,
		Blend:          parseBlend(req.Blend),
		Constraints:    constraints,
		IncludeReasons: req.IncludeReasons,
		ExplainLevel:   algorithm.NormalizeExplainLevel(req.ExplainLevel),
		Surface:        extractSurface(req.Context),
	}, nil
}

func extractSurface(ctx map[string]any) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx["surface"]; ok {
		if s, ok := val.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func filterAnchorsByAvailability(candidates []string, availability map[string]bool) []string {
	if len(candidates) == 0 {
		return nil
	}
	filtered := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, id := range candidates {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		if availability != nil {
			if ok, present := availability[trimmed]; !present || !ok {
				continue
			}
		}
		seen[trimmed] = struct{}{}
		filtered = append(filtered, trimmed)
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func (s *Service) starterAnchors(
	ctx context.Context,
	cfg algorithm.Config,
	req algorithm.Request,
	starterProfile map[string]float64,
	limit int,
) []string {
	if limit <= 0 || len(starterProfile) == 0 || s.store == nil {
		return nil
	}

	type tagWeight struct {
		tag    string
		weight float64
	}
	weights := make([]tagWeight, 0, len(starterProfile))
	for tag, weight := range starterProfile {
		if weight <= 0 {
			continue
		}
		weights = append(weights, tagWeight{tag: tag, weight: weight})
	}
	if len(weights) == 0 {
		return nil
	}
	sort.Slice(weights, func(i, j int) bool {
		return weights[i].weight > weights[j].weight
	})

	perTagLimit := limit
	if perTagLimit < 3 {
		perTagLimit = 3
	}

	type tagBucket struct {
		tag   string
		items []string
		index int
	}

	seen := make(map[string]struct{}, limit*2)
	buckets := make([]tagBucket, 0, len(weights))
	for _, entry := range weights {
		constraints := &recmodel.PopConstraints{IncludeTagsAny: []string{entry.tag}}
		items, err := s.store.PopularityTopK(ctx, req.OrgID, req.Namespace, cfg.HalfLifeDays, perTagLimit, constraints)
		if err != nil {
			continue
		}
		sanitized := make([]string, 0, len(items))
		for _, item := range items {
			id := strings.TrimSpace(item.ItemID)
			if id == "" {
				continue
			}
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			sanitized = append(sanitized, id)
		}
		if len(sanitized) == 0 {
			continue
		}
		buckets = append(buckets, tagBucket{tag: entry.tag, items: sanitized})
	}
	if len(buckets) == 0 {
		return nil
	}

	anchors := make([]string, 0, limit)
	for added := true; len(anchors) < limit && added; {
		added = false
		for i := range buckets {
			if len(anchors) >= limit {
				break
			}
			bucket := &buckets[i]
			if bucket.index >= len(bucket.items) {
				continue
			}
			id := bucket.items[bucket.index]
			bucket.index++
			if id == "" {
				continue
			}
			anchors = append(anchors, id)
			added = true
		}
	}
	return anchors
}

func parseConstraints(src *spectypes.RecommendConstraints) (*recmodel.PopConstraints, error) {
	if src == nil {
		return nil, nil
	}
	constraints := &recmodel.PopConstraints{
		IncludeTagsAny: src.IncludeTagsAny,
		ExcludeItemIDs: src.ExcludeItemIDs,
	}
	if len(src.PriceBetween) >= 1 {
		v := src.PriceBetween[0]
		constraints.MinPrice = &v
	}
	if len(src.PriceBetween) >= 2 {
		v := src.PriceBetween[1]
		constraints.MaxPrice = &v
	}
	if src.CreatedAfterISO != "" {
		ts, err := time.Parse(time.RFC3339, src.CreatedAfterISO)
		if err != nil {
			return nil, ValidationError{Code: "invalid_created_after", Message: "created_after must be RFC3339"}
		}
		constraints.CreatedAfter = &ts
	}
	return constraints, nil
}

func parseBlend(src *spectypes.RecommendBlend) *algorithm.BlendWeights {
	if src == nil {
		return nil
	}
	return &algorithm.BlendWeights{
		Pop:        src.Pop,
		Cooc:       src.Cooc,
		Similarity: src.Similarity,
	}
}

func applyOverrides(cfg *algorithm.Config, overrides *spectypes.Overrides) {
	if overrides == nil {
		return
	}
	if overrides.BlendAlpha != nil {
		cfg.BlendAlpha = *overrides.BlendAlpha
	}
	if overrides.BlendBeta != nil {
		cfg.BlendBeta = *overrides.BlendBeta
	}
	if overrides.BlendGamma != nil {
		cfg.BlendGamma = *overrides.BlendGamma
	}
	if overrides.ProfileBoost != nil {
		cfg.ProfileBoost = *overrides.ProfileBoost
	}
	if overrides.ProfileWindowDays != nil {
		cfg.ProfileWindowDays = float64(*overrides.ProfileWindowDays)
	}
	if overrides.ProfileTopN != nil {
		cfg.ProfileTopNTags = *overrides.ProfileTopN
	}
	if overrides.ProfileStarterBlendWeight != nil {
		weight := *overrides.ProfileStarterBlendWeight
		if weight < 0 {
			weight = 0
		} else if weight > 1 {
			weight = 1
		}
		cfg.ProfileStarterBlendWeight = weight
	}
	if overrides.MMRLambda != nil {
		cfg.MMRLambda = *overrides.MMRLambda
	}
	if overrides.BrandCap != nil {
		cfg.BrandCap = *overrides.BrandCap
	}
	if overrides.CategoryCap != nil {
		cfg.CategoryCap = *overrides.CategoryCap
	}
	if overrides.PopularityHalfLifeDays != nil {
		cfg.HalfLifeDays = float64(*overrides.PopularityHalfLifeDays)
	}
	if overrides.CoVisWindowDays != nil {
		cfg.CoVisWindowDays = *overrides.CoVisWindowDays
	}
	if overrides.PurchasedWindowDays != nil {
		cfg.PurchasedWindowDays = *overrides.PurchasedWindowDays
	}
	if overrides.RuleExcludeEvents != nil {
		cfg.RuleExcludeEvents = *overrides.RuleExcludeEvents
	}
	if overrides.PopularityFanout != nil {
		cfg.PopularityFanout = *overrides.PopularityFanout
	}
}

func (s *Service) resolveBlend(ctx context.Context, namespace string) *ResolvedBlendConfig {
	if s.blendResolver == nil {
		return nil
	}
	resolved, err := s.blendResolver.ResolveBlend(ctx, normalizeNamespace(namespace))
	if err != nil {
		return nil
	}
	return resolved
}

func (s *Service) applySegmentBlendOverrides(cfg *algorithm.Config, segment string) {
	if len(s.segmentBlendOverrides) == 0 {
		return
	}
	key := strings.ToLower(strings.TrimSpace(segment))
	if key == "" {
		return
	}
	if resolved, ok := s.segmentBlendOverrides[key]; ok {
		override := resolved
		blendOverrides(cfg, &override)
	}
}

func applySegmentProfile(cfg *algorithm.Config, profile types.SegmentProfile) {
	cfg.BlendAlpha = profile.BlendAlpha
	cfg.BlendBeta = profile.BlendBeta
	cfg.BlendGamma = profile.BlendGamma
	cfg.MMRLambda = profile.MMRLambda
	cfg.BrandCap = profile.BrandCap
	cfg.CategoryCap = profile.CategoryCap
	cfg.ProfileBoost = profile.ProfileBoost
	cfg.ProfileWindowDays = profile.ProfileWindowDays
	cfg.ProfileTopNTags = profile.ProfileTopN
	cfg.HalfLifeDays = profile.HalfLifeDays
	cfg.CoVisWindowDays = profile.CoVisWindowDays
	cfg.PurchasedWindowDays = profile.PurchasedWindowDays
	cfg.RuleExcludeEvents = profile.RuleExcludeEvents
	cfg.ExcludeEventTypes = append([]int16(nil), profile.ExcludeEventTypes...)
	cfg.BrandTagPrefixes = append([]string(nil), profile.BrandTagPrefixes...)
	cfg.CategoryTagPrefixes = append([]string(nil), profile.CategoryTagPrefixes...)
	cfg.PopularityFanout = profile.PopularityFanout
}

func convertToHTTPResponse(algoResp *algorithm.Response) spectypes.RecommendResponse {
	items := make([]spectypes.ScoredItem, 0, len(algoResp.Items))
	for _, item := range algoResp.Items {
		var explain *spectypes.ExplainBlock
		if item.Explain != nil {
			explain = mapExplainBlock(item.Explain)
		}
		items = append(items, spectypes.ScoredItem{
			ItemID:  item.ItemID,
			Score:   item.Score,
			Reasons: item.Reasons,
			Explain: explain,
		})
	}

	return spectypes.RecommendResponse{
		ModelVersion: algoResp.ModelVersion,
		Items:        items,
		SegmentID:    algoResp.SegmentID,
		ProfileID:    algoResp.ProfileID,
	}
}

// BuildAlgorithmRequest exposes request conversion for other handlers.
func BuildAlgorithmRequest(orgID uuid.UUID, req spectypes.RecommendRequest) (algorithm.Request, error) {
	return buildAlgorithmRequest(orgID, req)
}

// ApplyOverrides exposes override application for other handlers.
func ApplyOverrides(cfg *algorithm.Config, overrides *spectypes.Overrides) {
	applyOverrides(cfg, overrides)
}

// ConvertToHTTPResponse exposes response conversion for other handlers.
func ConvertToHTTPResponse(resp *algorithm.Response) spectypes.RecommendResponse {
	return convertToHTTPResponse(resp)
}

func mapExplainBlock(src *algorithm.ExplainBlock) *spectypes.ExplainBlock {
	if src == nil {
		return nil
	}
	dst := &spectypes.ExplainBlock{}

	if src.Blend != nil {
		dst.Blend = &spectypes.ExplainBlend{
			Alpha:          src.Blend.Alpha,
			Beta:           src.Blend.Beta,
			Gamma:          src.Blend.Gamma,
			PopNorm:        src.Blend.PopNorm,
			CoocNorm:       src.Blend.CoocNorm,
			SimilarityNorm: src.Blend.SimilarityNorm,
			Contributions: spectypes.ExplainBlendContribution{
				Pop:        src.Blend.Contributions.Pop,
				Cooc:       src.Blend.Contributions.Cooc,
				Similarity: src.Blend.Contributions.Similarity,
			},
		}
		if src.Blend.Raw != nil {
			dst.Blend.Raw = &spectypes.ExplainBlendRaw{
				Pop:        src.Blend.Raw.Pop,
				Cooc:       src.Blend.Raw.Cooc,
				Similarity: src.Blend.Raw.Similarity,
			}
		}
	}

	if src.Personalization != nil {
		dst.Personalization = &spectypes.ExplainPersonalization{
			Overlap:         src.Personalization.Overlap,
			BoostMultiplier: src.Personalization.BoostMultiplier,
		}
	}

	if src.MMR != nil {
		dst.MMR = &spectypes.ExplainMMR{
			Lambda:        src.MMR.Lambda,
			MaxSimilarity: src.MMR.MaxSimilarity,
			Penalty:       src.MMR.Penalty,
			Relevance:     src.MMR.Relevance,
			Rank:          src.MMR.Rank,
		}
	}

	if src.Caps != nil {
		caps := &spectypes.ExplainCaps{}
		if src.Caps.Brand != nil {
			caps.Brand = mapCapUsage(src.Caps.Brand)
		}
		if src.Caps.Category != nil {
			caps.Category = mapCapUsage(src.Caps.Category)
		}
		if caps.Brand != nil || caps.Category != nil {
			dst.Caps = caps
		}
	}

	if len(src.Anchors) > 0 {
		dst.Anchors = append([]string(nil), src.Anchors...)
	}

	if dst.Blend == nil && dst.Personalization == nil && dst.MMR == nil && dst.Caps == nil && len(dst.Anchors) == 0 {
		return nil
	}
	return dst
}

func mapCapUsage(src *algorithm.CapUsage) *spectypes.ExplainCapUsage {
	if src == nil {
		return nil
	}
	usage := &spectypes.ExplainCapUsage{
		Applied: src.Applied,
		Value:   src.Value,
	}
	if src.Limit != nil {
		limit := *src.Limit
		usage.Limit = &limit
	}
	if src.Count != nil {
		count := *src.Count
		usage.Count = &count
	}
	return usage
}
