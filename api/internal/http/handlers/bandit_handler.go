package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"recsys/internal/algorithm"
	"recsys/internal/bandit"
	"recsys/internal/http/common"
	"recsys/internal/services/recommendation"
	"recsys/internal/store"
	internaltypes "recsys/internal/types"
	handlerstypes "recsys/specs/types"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BanditHandler struct {
	store      *store.Store
	service    RecommendationService
	config     *RecommendationConfigManager
	tracer     *decisionTracer
	logger     *zap.Logger
	defaultOrg uuid.UUID
	banditAlgo internaltypes.Algorithm
}

func NewBanditHandler(store *store.Store, service RecommendationService, cfg *RecommendationConfigManager, tracer *decisionTracer, defaultOrg uuid.UUID, banditAlgo internaltypes.Algorithm, logger *zap.Logger) *BanditHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BanditHandler{
		store:      store,
		service:    service,
		config:     cfg,
		tracer:     tracer,
		logger:     logger,
		defaultOrg: defaultOrg,
		banditAlgo: banditAlgo,
	}
}

func (h *BanditHandler) currentConfig() RecommendationConfig {
	if h == nil || h.config == nil {
		return RecommendationConfig{}
	}
	return h.config.Current()
}

// BanditPoliciesUpsert godoc
// @Summary Upsert bandit policies
// @Tags bandit
// @Accept json
// @Produce json
// @Param payload body types.BanditPoliciesUpsertRequest true "Policies"
// @Success 202 {object} types.Ack
// @Router /v1/bandit/policies:upsert [post]
// @ID upsertBanditPolicies
func (h *BanditHandler) BanditPoliciesUpsert(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.BanditPoliciesUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}
	if req.Namespace == "" {
		common.BadRequest(w, r, "missing_namespace", "namespace is required", nil)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg).String()
	rows := make([]internaltypes.PolicyConfig, 0, len(req.Policies))
	for _, p := range req.Policies {
		rows = append(rows, internaltypes.PolicyConfig{
			PolicyID:          p.PolicyID,
			Name:              p.Name,
			Active:            p.Active,
			BlendAlpha:        p.BlendAlpha,
			BlendBeta:         p.BlendBeta,
			BlendGamma:        p.BlendGamma,
			MMRLambda:         p.MMRLambda,
			BrandCap:          p.BrandCap,
			CategoryCap:       p.CategoryCap,
			ProfileBoost:      p.ProfileBoost,
			RuleExcludeEvents: p.RuleExcludeEvents,
			HalfLifeDays:      p.HalfLifeDays,
			CoVisWindowDays:   p.CoVisWindowDays,
			PopularityFanout:  p.PopularityFanout,
			Notes:             p.Notes,
		})
	}

	if err := h.store.UpsertBanditPolicies(r.Context(), orgID, req.Namespace, rows); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(handlerstypes.Ack{Status: "accepted"})
}

// BanditPoliciesList godoc
// @Summary List all bandit policies (active and inactive)
// @Tags bandit
// @Produce json
// @Param namespace query string true "Namespace"
// @Success 200 {array} types.BanditPolicy
// @Router /v1/bandit/policies [get]
func (h *BanditHandler) BanditPoliciesList(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		common.BadRequest(w, r, "missing_namespace", "namespace is required", nil)
		return
	}
	orgID := orgIDFromHeader(r, h.defaultOrg).String()

	rows, err := h.store.ListAllPolicies(r.Context(), orgID, ns)
	if err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	out := make([]handlerstypes.BanditPolicy, 0, len(rows))
	for _, p := range rows {
		out = append(out, handlerstypes.BanditPolicy{
			PolicyID:          p.PolicyID,
			Name:              p.Name,
			Active:            p.Active,
			BlendAlpha:        p.BlendAlpha,
			BlendBeta:         p.BlendBeta,
			BlendGamma:        p.BlendGamma,
			MMRLambda:         p.MMRLambda,
			BrandCap:          p.BrandCap,
			CategoryCap:       p.CategoryCap,
			ProfileBoost:      p.ProfileBoost,
			RuleExcludeEvents: p.RuleExcludeEvents,
			HalfLifeDays:      p.HalfLifeDays,
			CoVisWindowDays:   p.CoVisWindowDays,
			PopularityFanout:  p.PopularityFanout,
			Notes:             p.Notes,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// BanditDecide godoc
// @Summary Decide best policy for this request context
// @Tags bandit
// @Accept json
// @Produce json
// @Param payload body types.BanditDecideRequest true "Decision request"
// @Success 200 {object} types.BanditDecideResponse
// @Router /v1/bandit/decide [post]
func (h *BanditHandler) BanditDecide(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.BanditDecideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}
	if req.Namespace == "" || req.Surface == "" {
		common.BadRequest(w, r, "missing_fields", "namespace and surface are required", nil)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg).String()
	bucket := bandit.BucketKeyFromContext(req.Context)

	mgr := h.newBanditStoreManager(h.banditAlgoOverride(req.Algorithm))
	dec, err := mgr.Decide(r.Context(), orgID, req.Namespace, req.Surface, bucket, req.CandidatePolicyIDs, req.RequestID)
	if err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}

	resp := handlerstypes.BanditDecideResponse{
		PolicyID:   dec.PolicyID,
		Algorithm:  string(dec.Algorithm),
		Surface:    dec.Surface,
		BucketKey:  dec.BucketKey,
		Explore:    dec.Explore,
		Explain:    dec.Explain,
		Experiment: dec.Experiment,
		Variant:    dec.Variant,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// BanditReward godoc
// @Summary Report binary reward for a previous decision
// @Tags bandit
// @Accept json
// @Produce json
// @Param payload body types.BanditRewardRequest true "Reward request"
// @Success 202 {object} types.Ack
// @Router /v1/bandit/reward [post]
func (h *BanditHandler) BanditReward(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.BanditRewardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}
	if req.Namespace == "" || req.Surface == "" || req.PolicyID == "" || req.BucketKey == "" {
		common.BadRequest(w, r, "missing_fields", "namespace, surface, policy_id, bucket_key required", nil)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg).String()
	mgr := h.newBanditStoreManager(h.banditAlgoOverride(req.Algorithm))
	var rewardMeta map[string]any
	if req.Experiment != "" || req.Variant != "" {
		rewardMeta = make(map[string]any, 2)
		if req.Experiment != "" {
			rewardMeta["experiment"] = req.Experiment
		}
		if req.Variant != "" {
			rewardMeta["variant"] = req.Variant
		}
	}
	err := mgr.Reward(r.Context(), orgID, req.Namespace, bandit.RewardInput{
		PolicyID:  req.PolicyID,
		Surface:   req.Surface,
		BucketKey: req.BucketKey,
		Reward:    req.Reward,
		Algorithm: h.banditAlgoOverride(req.Algorithm),
		Meta:      rewardMeta,
	}, req.RequestID)
	if err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(handlerstypes.Ack{Status: "accepted"})
}

// RecommendWithBandit godoc
// @Summary Recommend with bandit-selected policy
// @Tags ranking
// @Accept json
// @Produce json
// @Param payload body types.RecommendWithBanditRequest true "Request"
// @Success 200 {object} types.RecommendWithBanditResponse
// @Router /v1/bandit/recommendations [post]
func (h *BanditHandler) RecommendWithBandit(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req handlerstypes.RecommendWithBanditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}
	if req.Namespace == "" || req.Surface == "" {
		common.BadRequest(w, r, "missing_fields", "namespace and surface are required", nil)
		return
	}

	orgUUID := orgIDFromHeader(r, h.defaultOrg)
	orgID := orgUUID.String()

	bucket := bandit.BucketKeyFromContext(req.Context)
	mgr := h.newBanditStoreManager(h.banditAlgoOverride(req.Algorithm))
	dec, err := mgr.Decide(r.Context(), orgID, req.Namespace, req.Surface, bucket, req.CandidatePolicyIDs, "")
	if err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}

	cfg := h.currentConfig()
	algoCfg := cfg.BaseConfig()
	// Reset policy-driven fields prior to applying overrides.
	algoCfg.BlendAlpha = 0
	algoCfg.BlendBeta = 0
	algoCfg.BlendGamma = 0
	algoCfg.MMRLambda = 0
	algoCfg.BrandCap = 0
	algoCfg.CategoryCap = 0
	algoCfg.ProfileBoost = 0
	algoCfg.RuleExcludeEvents = false
	algoCfg.HalfLifeDays = 0
	algoCfg.CoVisWindowDays = 0
	algoCfg.PopularityFanout = 0

	if dec.PolicyID != "" {
		policies, err := h.store.ListPoliciesByIDs(r.Context(), orgID, req.Namespace, []string{dec.PolicyID})
		if err == nil && len(policies) == 1 {
			pl := policies[0]
			algoCfg.BlendAlpha = pl.BlendAlpha
			algoCfg.BlendBeta = pl.BlendBeta
			algoCfg.BlendGamma = pl.BlendGamma
			algoCfg.MMRLambda = pl.MMRLambda
			algoCfg.BrandCap = pl.BrandCap
			algoCfg.CategoryCap = pl.CategoryCap
			algoCfg.ProfileBoost = pl.ProfileBoost
			algoCfg.RuleExcludeEvents = pl.RuleExcludeEvents
			algoCfg.HalfLifeDays = pl.HalfLifeDays
			algoCfg.CoVisWindowDays = pl.CoVisWindowDays
			algoCfg.PopularityFanout = pl.PopularityFanout
		}
	}

	result, err := h.service.Recommend(r.Context(), orgUUID, req.RecommendRequest, algoCfg, h.segmentSelector())
	if err != nil {
		var vErr recommendation.ValidationError
		if errors.As(err, &vErr) {
			common.BadRequest(w, r, vErr.Code, vErr.Message, vErr.Details)
			return
		}
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	resp := handlerstypes.RecommendWithBanditResponse{
		RecommendResponse: result.Response,
		ChosenPolicyID:    dec.PolicyID,
		Algorithm:         string(dec.Algorithm),
		BanditBucket:      dec.BucketKey,
		Explore:           dec.Explore,
		BanditExplain:     dec.Explain,
		RequestID:         req.RequestID,
		BanditExperiment:  dec.Experiment,
		BanditVariant:     dec.Variant,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)

	var banditCtx *banditTraceContext
	if dec.PolicyID != "" {
		banditCtx = &banditTraceContext{
			ChosenPolicyID: dec.PolicyID,
			Algorithm:      string(dec.Algorithm),
			BucketKey:      dec.BucketKey,
			Explore:        dec.Explore,
			Explain:        dec.Explain,
			Experiment:     dec.Experiment,
			Variant:        dec.Variant,
		}
		if req.RequestID != "" {
			banditCtx.RequestID = req.RequestID
		} else if v := req.Context["request_id"]; v != "" {
			banditCtx.RequestID = v
		}
	}

	if h.tracer != nil {
		h.tracer.Record(decisionTraceInput{
			Request:      r,
			HTTPRequest:  req.RecommendRequest,
			AlgoRequest:  result.AlgoRequest,
			Config:       algoCfg,
			AlgoResponse: result.AlgoResponse,
			HTTPResponse: result.Response,
			TraceData:    result.TraceData,
			Duration:     time.Since(start),
			Surface:      req.Surface,
			Bandit:       banditCtx,
		})
	}
}

func (h *BanditHandler) segmentSelector() recommendation.SegmentSelector {
	return func(ctx context.Context, req algorithm.Request, httpReq handlerstypes.RecommendRequest) (recommendation.SegmentSelection, error) {
		sel, _, err := resolveSegmentSelection(ctx, h.store, req, httpReq, nil)
		return sel, err
	}
}

func (h *BanditHandler) banditAlgoOverride(s string) internaltypes.Algorithm {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case string(internaltypes.AlgorithmThompson):
		return internaltypes.AlgorithmThompson
	case string(internaltypes.AlgorithmUCB1):
		return internaltypes.AlgorithmUCB1
	default:
		return h.banditAlgo
	}
}

func (h *BanditHandler) newBanditStoreManager(algo internaltypes.Algorithm) *bandit.Manager {
	wrapped := &banditStoreAdapter{Store: h.store}
	cfg := h.currentConfig()
	exp := bandit.ExperimentConfig{
		Enabled:        cfg.BanditExperiment.Enabled,
		HoldoutPercent: cfg.BanditExperiment.HoldoutPercent,
		Label:          cfg.BanditExperiment.Label,
		Surfaces:       make(map[string]struct{}, len(cfg.BanditExperiment.Surfaces)),
	}
	for k := range cfg.BanditExperiment.Surfaces {
		exp.Surfaces[k] = struct{}{}
	}
	return bandit.NewManager(wrapped, algo, bandit.WithExperiment(exp))
}

type banditStoreAdapter struct {
	Store internaltypes.BanditStore
}

func (a *banditStoreAdapter) ListActivePolicies(ctx context.Context, orgID string, ns string) ([]internaltypes.PolicyConfig, error) {
	return a.Store.ListActivePolicies(ctx, orgID, ns)
}

func (a *banditStoreAdapter) ListPoliciesByIDs(ctx context.Context, orgID, ns string, ids []string) ([]internaltypes.PolicyConfig, error) {
	return a.Store.ListPoliciesByIDs(ctx, orgID, ns, ids)
}

func (a *banditStoreAdapter) GetStats(ctx context.Context, orgID string, ns string, surface string, bucket string, algo internaltypes.Algorithm) (map[string]internaltypes.Stats, error) {
	return a.Store.GetStats(ctx, orgID, ns, surface, bucket, algo)
}

func (a *banditStoreAdapter) IncrementStats(ctx context.Context, orgID string, ns string, surface string, bucket string, algo internaltypes.Algorithm, policyID string, reward bool) error {
	return a.Store.IncrementStats(ctx, orgID, ns, surface, bucket, algo, policyID, reward)
}

func (a *banditStoreAdapter) LogDecision(ctx context.Context, orgID string, ns string, surface string, bucket string, algo internaltypes.Algorithm, policyID string, explore bool, reqID string, meta map[string]any) error {
	return a.Store.LogDecision(ctx, orgID, ns, surface, bucket, algo, policyID, explore, reqID, meta)
}

func (a *banditStoreAdapter) LogReward(ctx context.Context, orgID, ns, surface, bucket string, algo internaltypes.Algorithm, policyID string, reward bool, reqID string, meta map[string]any) error {
	return a.Store.LogReward(ctx, orgID, ns, surface, bucket, algo, policyID, reward, reqID, meta)
}
