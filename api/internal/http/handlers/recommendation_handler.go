package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"recsys/internal/algorithm"
	"recsys/internal/http/common"
	policymetrics "recsys/internal/observability/policy"
	"recsys/internal/rules"
	"recsys/internal/services/recommendation"
	"recsys/internal/store"
	"recsys/internal/types"
	handlerstypes "recsys/specs/types"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RecommendationService captures the recommendation domain operations used by the HTTP layer.
type RecommendationService interface {
	Recommend(ctx context.Context, orgID uuid.UUID, req handlerstypes.RecommendRequest, baseCfg algorithm.Config, selector recommendation.SegmentSelector) (*recommendation.Result, error)
	Rerank(ctx context.Context, orgID uuid.UUID, req handlerstypes.RerankRequest, baseCfg algorithm.Config, selector recommendation.SegmentSelector) (*recommendation.Result, error)
}

// RecommendationHandler exposes ranking endpoints backed by the recommendation service.
type RecommendationHandler struct {
	service       RecommendationService
	store         *store.Store
	config        *RecommendationConfigManager
	tracer        *decisionTracer
	logger        *zap.Logger
	defaultOrg    uuid.UUID
	policyMetrics *policymetrics.Metrics
	coverage      *coverageTracker
}

// RecommendationPresetsResponse describes preset configuration payloads.
type RecommendationPresetsResponse struct {
	MMRPresets map[string]float64 `json:"mmr_presets"`
}

type recommendationResponseEnvelope struct {
	handlerstypes.RecommendResponse
	Trace *traceDebugPayload `json:"trace,omitempty"`
}

type traceDebugPayload struct {
	Extras map[string]any `json:"extras,omitempty"`
}

type traceSourceMetric struct {
	Count      int     `json:"count"`
	DurationMS float64 `json:"duration_ms"`
}

func NewRecommendationHandler(service RecommendationService, st *store.Store, cfg *RecommendationConfigManager, tracer *decisionTracer, defaultOrg uuid.UUID, logger *zap.Logger, policyMetrics *policymetrics.Metrics) *RecommendationHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	handler := &RecommendationHandler{
		service:       service,
		store:         st,
		config:        cfg,
		tracer:        tracer,
		logger:        logger,
		defaultOrg:    defaultOrg,
		policyMetrics: policyMetrics,
	}
	if st != nil && cfg != nil {
		initial := cfg.Current()
		handler.coverage = newCoverageTracker(st, initial.CoverageCacheTTL, initial.CoverageLongTailHintThreshold)
	}
	return handler
}

func (h *RecommendationHandler) ApplyCoverageConfig(cfg RecommendationConfig) {
	if h == nil || h.coverage == nil {
		return
	}
	h.coverage.applyConfig(cfg.CoverageCacheTTL, cfg.CoverageLongTailHintThreshold)
}

func (h *RecommendationHandler) currentConfig() RecommendationConfig {
	if h == nil || h.config == nil {
		return RecommendationConfig{}
	}
	return h.config.Current()
}

// Recommend godoc
// @Summary      Get recommendations for a user
// @Tags         ranking
// @Accept       json
// @Produce      json
// @Param        payload  body  types.RecommendRequest  true  "Recommend request"
// @Success      200      {object}  types.RecommendResponse
// @Failure      400      {object}  common.APIError
// @Router       /v1/recommendations [post]
func (h *RecommendationHandler) Recommend(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req handlerstypes.RecommendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg)
	cfg := h.currentConfig()

	result, err := h.service.Recommend(r.Context(), orgID, req, cfg.BaseConfig(), h.segmentSelector())
	if err != nil {
		var vErr recommendation.ValidationError
		if errors.As(err, &vErr) {
			common.BadRequest(w, r, vErr.Code, vErr.Message, vErr.Details)
			return
		}
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	h.respondWithRecommendation(w, r, req, orgID, start, result)
}

// Rerank godoc
// @Summary      Rerank a candidate set
// @Tags         ranking
// @Accept       json
// @Produce      json
// @Param        payload  body  types.RerankRequest  true  "Rerank request"
// @Success      200      {object}  types.RecommendResponse
// @Failure      400      {object}  common.APIError
// @Router       /v1/rerank [post]
func (h *RecommendationHandler) Rerank(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req handlerstypes.RerankRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg)
	cfg := h.currentConfig()

	result, err := h.service.Rerank(r.Context(), orgID, req, cfg.BaseConfig(), h.segmentSelector())
	if err != nil {
		var vErr recommendation.ValidationError
		if errors.As(err, &vErr) {
			common.BadRequest(w, r, vErr.Code, vErr.Message, vErr.Details)
			return
		}
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	h.respondWithRecommendation(w, r, req, orgID, start, result)
}

// ItemSimilar godoc
// @Summary      Get similar items
// @Tags         ranking
// @Produce      json
// @Param        item_id  path  string  true  "Item ID"
// @Param        namespace query string false "Namespace"  default(default)
// @Param        k        query int     false "Top-K"  default(20)
// @Success      200      {array}  specs_types.ScoredItem
// @Failure      400      {object} common.APIError
// @Router       /v1/items/{item_id}/similar [get]
func (h *RecommendationHandler) ItemSimilar(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "item_id")
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		ns = "default"
	}
	k := 20
	if s := r.URL.Query().Get("k"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			k = v
		}
	}

	orgID := orgIDFromHeader(r, h.defaultOrg)

	cfg := h.currentConfig()
	similarEngine := algorithm.NewSimilarItemsEngine(h.store, int(cfg.CoVisWindowDays))
	algoReq := algorithm.SimilarItemsRequest{
		OrgID:     orgID,
		ItemID:    itemID,
		Namespace: ns,
		K:         k,
	}

	algoResp, err := similarEngine.FindSimilar(r.Context(), algoReq)
	if err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	httpResp := make([]handlerstypes.ScoredItem, 0, len(algoResp.Items))
	for _, item := range algoResp.Items {
		httpResp = append(httpResp, handlerstypes.ScoredItem{
			ItemID:  item.ItemID,
			Score:   item.Score,
			Reasons: item.Reasons,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(httpResp)
}

func (h *RecommendationHandler) segmentSelector() recommendation.SegmentSelector {
	return func(ctx context.Context, req algorithm.Request, httpReq handlerstypes.RecommendRequest) (recommendation.SegmentSelection, error) {
		sel, _, err := resolveSegmentSelection(ctx, h.store, req, httpReq, nil)
		return sel, err
	}
}

func (h *RecommendationHandler) respondWithRecommendation(w http.ResponseWriter, r *http.Request, httpReq any, orgID uuid.UUID, start time.Time, result *recommendation.Result) {
	respBody := recommendationResponseEnvelope{RecommendResponse: result.Response}
	if trace := buildTraceDebugPayload(result.TraceData); trace != nil {
		respBody.Trace = trace
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(respBody); err != nil {
		h.logger.Error("write recommend response", zap.Error(err))
		return
	}

	if h.tracer != nil {
		traceReq := normalizeTraceRequest(httpReq)
		h.tracer.Record(decisionTraceInput{
			Request:      r,
			HTTPRequest:  traceReq,
			AlgoRequest:  result.AlgoRequest,
			Config:       result.AlgoConfig,
			AlgoResponse: result.AlgoResponse,
			HTTPResponse: result.Response,
			TraceData:    result.TraceData,
			Duration:     time.Since(start),
			Surface:      result.AlgoRequest.Surface,
		})
	}

	if result.TraceData != nil && result.TraceData.Policy != nil {
		summary := result.TraceData.Policy
		if h.policyMetrics != nil {
			h.policyMetrics.Observe(result.AlgoRequest, summary)
			if len(result.TraceData.ManualOverrideHits) > 0 {
				h.policyMetrics.ObserveOverrides(result.AlgoRequest, result.TraceData.ManualOverrideHits)
			}
		}
		if summary.RuleBoostCount > 0 || summary.RulePinCount > 0 || summary.RuleBlockCount > 0 {
			h.logger.Info("policy_rule_actions",
				zap.String("namespace", result.AlgoRequest.Namespace),
				zap.String("surface", result.AlgoRequest.Surface),
				zap.Int("k", result.AlgoRequest.K),
				zap.Int("boost_count", summary.RuleBoostCount),
				zap.Int("pin_count", summary.RulePinCount),
				zap.Int("block_count", summary.RuleBlockCount),
				zap.Strings("rule_ids", collectRuleIDs(result.TraceData.RuleMatches)),
			)
		}
		if summary.ConstraintLeakCount > 0 {
			h.logger.Warn("policy_constraint_leak",
				zap.String("namespace", result.AlgoRequest.Namespace),
				zap.String("surface", result.AlgoRequest.Surface),
				zap.Int("k", result.AlgoRequest.K),
				zap.Int("leak_count", summary.ConstraintLeakCount),
				zap.Strings("leaked_items", summary.ConstraintLeakIDs),
			)
		}
		if summary.RuleBoostExposure > 0 || summary.RulePinExposure > 0 {
			h.logger.Info("policy_rule_exposure",
				zap.String("namespace", result.AlgoRequest.Namespace),
				zap.String("surface", result.AlgoRequest.Surface),
				zap.Int("k", result.AlgoRequest.K),
				zap.Int("boost_exposure", summary.RuleBoostExposure),
				zap.Int("pin_exposure", summary.RulePinExposure),
				zap.Int("boost_injected", summary.RuleBoostInjected),
			)
		}
		if summary.RuleBoostCount > 0 && summary.RuleBoostExposure == 0 {
			h.logger.Warn("policy_rule_zero_effect",
				zap.String("namespace", result.AlgoRequest.Namespace),
				zap.String("surface", result.AlgoRequest.Surface),
				zap.Int("k", result.AlgoRequest.K),
				zap.String("type", "boost"),
				zap.Strings("rule_ids", collectRuleIDsByAction(result.TraceData.RuleMatches, types.RuleActionBoost)),
			)
		}
		if summary.RulePinCount > 0 && summary.RulePinExposure == 0 {
			h.logger.Warn("policy_rule_zero_effect",
				zap.String("namespace", result.AlgoRequest.Namespace),
				zap.String("surface", result.AlgoRequest.Surface),
				zap.Int("k", result.AlgoRequest.K),
				zap.String("type", "pin"),
				zap.Strings("rule_ids", collectRuleIDsByAction(result.TraceData.RuleMatches, types.RuleActionPin)),
			)
		}
		if summary.RuleBlockExposure > 0 {
			h.logger.Info("policy_rule_block_exposure",
				zap.String("namespace", result.AlgoRequest.Namespace),
				zap.String("surface", result.AlgoRequest.Surface),
				zap.Int("k", result.AlgoRequest.K),
				zap.Int("blocked_items", summary.RuleBlockExposure),
			)
		}
	}
	if result.TraceData != nil && len(result.TraceData.ManualOverrideHits) > 0 {
		h.logger.Info("manual_override_activity",
			zap.String("namespace", result.AlgoRequest.Namespace),
			zap.String("surface", result.AlgoRequest.Surface),
			zap.Int("k", result.AlgoRequest.K),
			zap.Any("overrides", summarizeOverrideHits(result.TraceData.ManualOverrideHits)),
		)
	}
	if result.TraceData != nil && len(result.TraceData.StarterProfile) > 0 {
		h.logStarterDiversity(r.Context(), orgID, result)
	}

	if h.coverage != nil && h.policyMetrics != nil && len(result.Response.Items) > 0 {
		itemIDs := make([]string, 0, len(result.Response.Items))
		for _, item := range result.Response.Items {
			if trimmed := strings.TrimSpace(item.ItemID); trimmed != "" {
				itemIDs = append(itemIDs, trimmed)
			}
		}
		if len(itemIDs) > 0 {
			snapshot, err := h.coverage.snapshot(r.Context(), orgID, result.AlgoRequest.Namespace, itemIDs)
			if err != nil {
				h.logger.Warn("coverage_snapshot_failed",
					zap.String("namespace", result.AlgoRequest.Namespace),
					zap.String("surface", result.AlgoRequest.Surface),
					zap.Error(err),
				)
			} else {
				h.policyMetrics.ObserveCoverage(result.AlgoRequest, itemIDs, snapshot.LongTailFlags, snapshot.TotalCatalog)
			}
		}
	}

	if len(result.SourceStats) > 0 {
		for source, metric := range result.SourceStats {
			h.logger.Info("candidate_source_metrics",
				zap.String("source", source),
				zap.Int("count", metric.Count),
				zap.Float64("duration_ms", durationMillis(metric.Duration)),
				zap.String("surface", result.AlgoRequest.Surface),
				zap.String("namespace", result.AlgoRequest.Namespace),
				zap.Int("k", result.AlgoRequest.K),
			)
		}
	}
}

// RecommendationPresets godoc
// @Summary     List recommendation presets
// @Tags        recommendation
// @Produce     json
// @Success     200 {object} RecommendationPresetsResponse
// @Failure     500 {object} common.APIError
// @Router      /v1/admin/recommendation/presets [get]
func (h *RecommendationHandler) RecommendationPresets(w http.ResponseWriter, r *http.Request) {
	resp := RecommendationPresetsResponse{MMRPresets: make(map[string]float64)}
	cfg := h.currentConfig()
	for surface, value := range cfg.MMRPresets {
		resp.MMRPresets[surface] = value
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("write recommendation presets", zap.Error(err))
	}
}

func buildTraceDebugPayload(traceData *algorithm.TraceData) *traceDebugPayload {
	if traceData == nil {
		return nil
	}

	extras := make(map[string]any)
	if traceData.Policy != nil {
		extras["policy"] = traceData.Policy
	}
	if len(traceData.ManualOverrideHits) > 0 {
		overrides := make([]map[string]any, 0, len(traceData.ManualOverrideHits))
		for _, hit := range traceData.ManualOverrideHits {
			entry := map[string]any{
				"override_id": hit.OverrideID.String(),
				"rule_id":     hit.RuleID.String(),
				"action":      strings.ToLower(string(hit.Action)),
				"matched":     len(hit.MatchedItems),
			}
			if len(hit.MatchedItems) > 0 {
				items := append([]string(nil), hit.MatchedItems...)
				entry["matched_items"] = items
			}
			if len(hit.BlockedItems) > 0 {
				items := append([]string(nil), hit.BlockedItems...)
				entry["blocked_items"] = items
				entry["blocked"] = items
			}
			if len(hit.PinnedItems) > 0 {
				items := append([]string(nil), hit.PinnedItems...)
				entry["pinned_items"] = items
				entry["pinned"] = items
			}
			if len(hit.BoostedItems) > 0 {
				items := append([]string(nil), hit.BoostedItems...)
				entry["boosted_items"] = items
				entry["boosted"] = items
			}
			if len(hit.ServedItems) > 0 {
				items := append([]string(nil), hit.ServedItems...)
				entry["served_items"] = items
				entry["served"] = items
			}
			overrides = append(overrides, entry)
		}
		extras["manual_overrides"] = overrides
	}
	if len(traceData.StarterProfile) > 0 {
		extras["starter_profile"] = traceData.StarterProfile
		if traceData.StarterBlendWeight > 0 {
			extras["starter_profile_weight"] = traceData.StarterBlendWeight
		}
	}
	if traceData.RecentEventCount >= 0 {
		extras["recent_event_count"] = traceData.RecentEventCount
	}
	if len(traceData.SourceMetrics) > 0 {
		sources := make(map[string]traceSourceMetric, len(traceData.SourceMetrics))
		for source, metric := range traceData.SourceMetrics {
			sources[source] = traceSourceMetric{
				Count:      metric.Count,
				DurationMS: durationMillis(metric.Duration),
			}
		}
		extras["candidate_sources"] = sources
	}

	if len(extras) == 0 {
		return nil
	}
	return &traceDebugPayload{Extras: extras}
}

func durationMillis(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

func collectRuleIDs(matches []rules.Match) []string {
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		id := strings.ToLower(match.RuleID.String())
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func normalizeTraceRequest(raw any) handlerstypes.RecommendRequest {
	switch req := raw.(type) {
	case handlerstypes.RecommendRequest:
		return req
	case handlerstypes.RerankRequest:
		return handlerstypes.RecommendRequest{
			UserID:         req.UserID,
			Namespace:      req.Namespace,
			K:              req.K,
			Constraints:    req.Constraints,
			Blend:          req.Blend,
			Overrides:      req.Overrides,
			Context:        req.Context,
			IncludeReasons: req.IncludeReasons,
			ExplainLevel:   req.ExplainLevel,
		}
	default:
		return handlerstypes.RecommendRequest{}
	}
}

func collectRuleIDsByAction(matches []rules.Match, action types.RuleAction) []string {
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		if match.Action != action {
			continue
		}
		id := strings.ToLower(match.RuleID.String())
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func summarizeOverrideHits(hits []rules.OverrideHit) []map[string]any {
	if len(hits) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(hits))
	for _, hit := range hits {
		entry := map[string]any{
			"override_id": hit.OverrideID.String(),
			"rule_id":     hit.RuleID.String(),
			"action":      strings.ToLower(string(hit.Action)),
			"matched":     len(hit.MatchedItems),
		}
		if len(hit.BlockedItems) > 0 {
			entry["blocked"] = append([]string(nil), hit.BlockedItems...)
		}
		if len(hit.PinnedItems) > 0 {
			entry["pinned"] = append([]string(nil), hit.PinnedItems...)
		}
		if len(hit.BoostedItems) > 0 {
			entry["boosted"] = append([]string(nil), hit.BoostedItems...)
		}
		if len(hit.ServedItems) > 0 {
			entry["served"] = append([]string(nil), hit.ServedItems...)
		}
		out = append(out, entry)
	}
	return out
}

func (h *RecommendationHandler) logStarterDiversity(ctx context.Context, orgID uuid.UUID, result *recommendation.Result) {
	if h == nil || h.store == nil || result == nil {
		return
	}
	trace := result.TraceData
	if trace == nil || len(trace.StarterProfile) == 0 {
		return
	}
	if len(result.Response.Items) == 0 {
		return
	}
	itemIDs := make([]string, 0, len(result.Response.Items))
	for _, item := range result.Response.Items {
		id := strings.TrimSpace(item.ItemID)
		if id != "" {
			itemIDs = append(itemIDs, id)
		}
	}
	if len(itemIDs) == 0 {
		return
	}
	tags, err := h.store.ListItemsTags(ctx, orgID, result.AlgoRequest.Namespace, itemIDs)
	if err != nil || len(tags) == 0 {
		h.logger.Debug("starter_diversity_tags_failed",
			zap.String("namespace", result.AlgoRequest.Namespace),
			zap.String("surface", result.AlgoRequest.Surface),
			zap.Error(err),
		)
		return
	}
	prefixes := result.AlgoConfig.CategoryTagPrefixes
	if len(prefixes) == 0 {
		cfg := h.currentConfig()
		prefixes = cfg.CategoryTagPrefixes
	}
	counts := make(map[string]int)
	total := 0
	for _, item := range result.Response.Items {
		id := strings.TrimSpace(item.ItemID)
		if id == "" {
			continue
		}
		info, ok := tags[id]
		if !ok {
			continue
		}
		category := deriveCategoryFromTags(info, prefixes)
		if category == "" {
			category = "unknown"
		}
		counts[category]++
		total++
	}
	if total == 0 {
		return
	}
	entropy := categoryEntropy(counts, total)
	h.logger.Info("starter_diversity",
		zap.String("namespace", result.AlgoRequest.Namespace),
		zap.String("surface", result.AlgoRequest.Surface),
		zap.String("segment_id", result.AlgoRequest.SegmentID),
		zap.Int("k", result.AlgoRequest.K),
		zap.Float64("starter_weight", trace.StarterBlendWeight),
		zap.Int("recent_event_count", trace.RecentEventCount),
		zap.Int("unique_categories", len(counts)),
		zap.Float64("entropy", entropy),
		zap.Any("category_counts", counts),
	)
}

func deriveCategoryFromTags(info types.ItemTags, prefixes []string) string {
	if len(info.Tags) == 0 {
		return ""
	}
	normalized := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		p := strings.ToLower(strings.TrimSpace(prefix))
		p = strings.TrimSuffix(p, ":")
		if p != "" {
			normalized = append(normalized, p)
		}
	}
	for _, tag := range info.Tags {
		lower := strings.ToLower(strings.TrimSpace(tag))
		if lower == "" {
			continue
		}
		for _, prefix := range normalized {
			withColon := prefix + ":"
			if strings.HasPrefix(lower, withColon) {
				value := strings.TrimPrefix(lower, withColon)
				value = strings.Trim(value, " _-")
				if value != "" {
					return value
				}
			}
			if strings.HasPrefix(lower, prefix) && lower != prefix {
				value := strings.TrimPrefix(lower, prefix)
				value = strings.Trim(value, ": _-")
				if value != "" {
					return value
				}
			}
		}
	}
	return ""
}

func categoryEntropy(counts map[string]int, total int) float64 {
	if total <= 0 {
		return 0
	}
	totalF := float64(total)
	entropy := 0.0
	for _, count := range counts {
		if count <= 0 {
			continue
		}
		p := float64(count) / totalF
		entropy -= p * math.Log2(p)
	}
	return entropy
}
