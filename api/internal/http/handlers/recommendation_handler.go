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
	service    RecommendationService
	store      *store.Store
	config     RecommendationConfig
	tracer     *decisionTracer
	logger     *zap.Logger
	defaultOrg uuid.UUID
}

func NewRecommendationHandler(service RecommendationService, st *store.Store, cfg RecommendationConfig, tracer *decisionTracer, defaultOrg uuid.UUID, logger *zap.Logger) *RecommendationHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RecommendationHandler{
		service:    service,
		store:      st,
		config:     cfg,
		tracer:     tracer,
		logger:     logger,
		defaultOrg: defaultOrg,
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result.Response); err != nil {
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
