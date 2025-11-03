package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"recsys/internal/algorithm"
	"recsys/internal/http/common"
	policymetrics "recsys/internal/observability/policy"
	"recsys/internal/services/recommendation"
	"recsys/internal/store"
	handlerstypes "recsys/specs/types"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RecommendationService captures the recommendation domain operations used by the HTTP layer.
type RecommendationService interface {
	Recommend(ctx context.Context, orgID uuid.UUID, req handlerstypes.RecommendRequest, baseCfg algorithm.Config, selector recommendation.SegmentSelector) (*recommendation.Result, error)
}

// RecommendationHandler exposes ranking endpoints backed by the recommendation service.
type RecommendationHandler struct {
	service       RecommendationService
	store         *store.Store
	config        RecommendationConfig
	tracer        *decisionTracer
	logger        *zap.Logger
	defaultOrg    uuid.UUID
	policyMetrics *policymetrics.Metrics
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

func NewRecommendationHandler(service RecommendationService, st *store.Store, cfg RecommendationConfig, tracer *decisionTracer, defaultOrg uuid.UUID, logger *zap.Logger, policyMetrics *policymetrics.Metrics) *RecommendationHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RecommendationHandler{
		service:       service,
		store:         st,
		config:        cfg,
		tracer:        tracer,
		logger:        logger,
		defaultOrg:    defaultOrg,
		policyMetrics: policyMetrics,
	}
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

	result, err := h.service.Recommend(r.Context(), orgID, req, h.config.BaseConfig(), h.segmentSelector())
	if err != nil {
		var vErr recommendation.ValidationError
		if errors.As(err, &vErr) {
			common.BadRequest(w, r, vErr.Code, vErr.Message, vErr.Details)
			return
		}
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	respBody := recommendationResponseEnvelope{
		RecommendResponse: result.Response,
	}
	if trace := buildTraceDebugPayload(result.TraceData); trace != nil {
		respBody.Trace = trace
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(respBody); err != nil {
		h.logger.Error("write recommend response", zap.Error(err))
		return
	}

	if h.tracer != nil {
		h.tracer.Record(decisionTraceInput{
			Request:      r,
			HTTPRequest:  req,
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

	similarEngine := algorithm.NewSimilarItemsEngine(h.store, int(h.config.CoVisWindowDays))
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

func buildTraceDebugPayload(traceData *algorithm.TraceData) *traceDebugPayload {
	if traceData == nil {
		return nil
	}

	extras := make(map[string]any)
	if traceData.Policy != nil {
		extras["policy"] = traceData.Policy
	}
	if len(traceData.StarterProfile) > 0 {
		extras["starter_profile"] = traceData.StarterProfile
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
