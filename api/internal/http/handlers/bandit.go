package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"recsys/internal/algorithm"
	"recsys/internal/bandit"
	"recsys/internal/http/common"
	"recsys/internal/http/types"
	internaltypes "recsys/internal/types"
)

// @Summary Upsert bandit policies
// @Tags bandit
// @Accept json
// @Produce json
// @Param payload body types.BanditPoliciesUpsertRequest true "Policies"
// @Success 202 {object} types.Ack
// @Router /v1/bandit/policies:upsert [post]
// @ID upsertBanditPolicies
func (h *Handler) BanditPoliciesUpsert(w http.ResponseWriter, r *http.Request) {
	var req types.BanditPoliciesUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	if req.Namespace == "" {
		common.BadRequest(
			w,
			r,
			"missing_namespace",
			"namespace is required",
			nil,
		)
		return
	}
	orgID := h.defaultOrgFromHeader(r).String()
	rows := make([]internaltypes.PolicyConfig, 0, len(req.Policies))
	for _, p := range req.Policies {
		rows = append(rows, internaltypes.PolicyConfig{
			PolicyID:    p.PolicyID,
			Name:        p.Name,
			Active:      p.Active,
			BlendAlpha:  p.BlendAlpha,
			BlendBeta:   p.BlendBeta,
			BlendGamma:  p.BlendGamma,
			MMRLambda:   p.MMRLambda,
			BrandCap:    p.BrandCap,
			CategoryCap: p.CategoryCap,
			Notes:       p.Notes,
		})
	}
	if err := h.Store.UpsertBanditPolicies(
		r.Context(), orgID, req.Namespace, rows,
	); err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(types.Ack{Status: "accepted"})
}

// @Summary List all bandit policies (active and inactive)
// @Tags bandit
// @Produce json
// @Param namespace query string true "Namespace"
// @Success 200 {array} types.BanditPolicy
// @Router /v1/bandit/policies [get]
func (h *Handler) BanditPoliciesList(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		common.BadRequest(
			w, r, "missing_namespace", "namespace is required", nil,
		)
		return
	}
	orgID := h.defaultOrgFromHeader(r).String()
	rows, err := h.Store.ListAllPolicies(r.Context(), orgID, ns)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	out := make([]types.BanditPolicy, 0, len(rows))
	for _, p := range rows {
		out = append(out, types.BanditPolicy{
			PolicyID:    p.PolicyID,
			Name:        p.Name,
			Active:      p.Active,
			BlendAlpha:  p.BlendAlpha,
			BlendBeta:   p.BlendBeta,
			BlendGamma:  p.BlendGamma,
			MMRLambda:   p.MMRLambda,
			BrandCap:    p.BrandCap,
			CategoryCap: p.CategoryCap,
			Notes:       p.Notes,
		})
	}
	_ = json.NewEncoder(w).Encode(out)
}

// @Summary Decide best policy for this request context
// @Tags bandit
// @Accept json
// @Produce json
// @Param payload body types.BanditDecideRequest true "Decision request"
// @Success 200 {object} types.BanditDecideResponse
// @Router /v1/bandit/decide [post]
func (h *Handler) BanditDecide(w http.ResponseWriter, r *http.Request) {
	var req types.BanditDecideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	if req.Namespace == "" || req.Surface == "" {
		common.BadRequest(
			w,
			r,
			"missing_fields",
			"namespace and surface are required",
			nil,
		)
		return
	}
	orgID := h.defaultOrgFromHeader(r).String()
	bucket := bandit.BucketKeyFromContext(req.Context)

	mgr := h.newBanditStoreManager(h.banditAlgoOverride(req.Algorithm))
	dec, err := mgr.Decide(
		r.Context(),
		orgID,
		req.Namespace,
		req.Surface,
		bucket,
		req.CandidatePolicyIDs,
		req.RequestID,
	)
	if err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	err = json.NewEncoder(w).Encode(types.BanditDecideResponse{
		PolicyID:  dec.PolicyID,
		Algorithm: string(dec.Algorithm),
		Surface:   dec.Surface,
		BucketKey: dec.BucketKey,
		Explore:   dec.Explore,
		Explain:   dec.Explain,
	})
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
}

// @Summary Report binary reward for a previous decision
// @Tags bandit
// @Accept json
// @Produce json
// @Param payload body types.BanditRewardRequest true "Reward request"
// @Success 202 {object} types.Ack
// @Router /v1/bandit/reward [post]
func (h *Handler) BanditReward(w http.ResponseWriter, r *http.Request) {
	var req types.BanditRewardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	if req.Namespace == "" ||
		req.Surface == "" ||
		req.PolicyID == "" ||
		req.BucketKey == "" {
		common.BadRequest(
			w,
			r,
			"missing_fields",
			"namespace, surface, policy_id, bucket_key required",
			nil,
		)
		return
	}
	orgID := h.defaultOrgFromHeader(r).String()
	mgr := h.newBanditStoreManager(h.banditAlgoOverride(req.Algorithm))
	err := mgr.Reward(r.Context(), orgID, req.Namespace, bandit.RewardInput{
		PolicyID:  req.PolicyID,
		Surface:   req.Surface,
		BucketKey: req.BucketKey,
		Reward:    req.Reward,
		Algorithm: h.banditAlgoOverride(req.Algorithm),
	}, req.RequestID)
	if err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(types.Ack{Status: "accepted"})
}

// @Summary Recommend with bandit-selected policy
// @Tags ranking
// @Accept json
// @Produce json
// @Param payload body types.RecommendWithBanditRequest true "Request"
// @Success 200 {object} types.RecommendWithBanditResponse
// @Router /v1/bandit/recommendations [post]
func (h *Handler) RecommendWithBandit(w http.ResponseWriter, r *http.Request) {
	var req types.RecommendWithBanditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	if req.Namespace == "" || req.Surface == "" {
		common.BadRequest(
			w,
			r,
			"missing_fields",
			"namespace and surface are required",
			nil,
		)
		return
	}

	algoReq, err := h.convertToAlgorithmRequest(r, req.RecommendRequest)
	if err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	orgID := h.defaultOrgFromHeader(r).String()
	bucket := bandit.BucketKeyFromContext(req.Context)

	// Decide policy.
	mgr := h.newBanditStoreManager(h.banditAlgoOverride(req.Algorithm))
	dec, err := mgr.Decide(
		r.Context(),
		orgID,
		req.Namespace,
		req.Surface,
		bucket,
		req.CandidatePolicyIDs,
		"",
	)
	if err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	// Run ranker with chosen policy knobs.
	cfg, err := h.getAlgorithmConfig(nil)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	cfg.BlendAlpha = 0
	cfg.BlendBeta = 0
	cfg.BlendGamma = 0
	cfg.MMRLambda = 0
	cfg.BrandCap = 0
	cfg.CategoryCap = 0

	// Fetch policy to populate knobs.
	pl, err := h.Store.ListPoliciesByIDs(
		r.Context(), orgID, req.Namespace, []string{dec.PolicyID},
	)
	if err == nil && len(pl) == 1 {
		cfg.BlendAlpha = pl[0].BlendAlpha
		cfg.BlendBeta = pl[0].BlendBeta
		cfg.BlendGamma = pl[0].BlendGamma
		cfg.MMRLambda = pl[0].MMRLambda
		cfg.BrandCap = pl[0].BrandCap
		cfg.CategoryCap = pl[0].CategoryCap
	} else {
		// If not found, fall back to env-based defaults already in cfg.
	}

	engine := algorithm.NewEngine(cfg, h.Store)
	algoResp, err := engine.Recommend(r.Context(), algoReq)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	out := types.RecommendWithBanditResponse{
		RecommendResponse: h.convertToHTTPResponse(algoResp),
		ChosenPolicyID:    dec.PolicyID,
		Algorithm:         string(dec.Algorithm),
		BanditBucket:      dec.BucketKey,
		Explore:           dec.Explore,
		BanditExplain:     dec.Explain,
	}
	_ = json.NewEncoder(w).Encode(out)
}

// banditAlgoOverride overrides the algorithm if provided.
func (h *Handler) banditAlgoOverride(s string) internaltypes.Algorithm {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case string(internaltypes.AlgorithmThompson):
		return internaltypes.AlgorithmThompson
	case string(internaltypes.AlgorithmUCB1):
		return internaltypes.AlgorithmUCB1
	default:
		return h.BanditAlgo
	}
}

// newBanditStoreManager creates a new bandit manager that wraps the handlers
// store.
func (h *Handler) newBanditStoreManager(
	algo internaltypes.Algorithm,
) *bandit.Manager {
	// The store already implements needed methods in bandit.go.
	wrapped := &banditStoreAdapter{Store: h.Store}
	return bandit.NewManager(wrapped, algo)
}

// banditStoreAdapter adapts the handlers store.
type banditStoreAdapter struct {
	Store internaltypes.BanditStore
}

// ListActivePolicies lists active policies.
func (a *banditStoreAdapter) ListActivePolicies(
	ctx context.Context,
	orgID string,
	ns string,
) ([]internaltypes.PolicyConfig, error) {
	return a.Store.ListActivePolicies(ctx, orgID, ns)
}

// ListPoliciesByIDs lists policies by IDs.
func (a *banditStoreAdapter) ListPoliciesByIDs(
	ctx context.Context,
	orgID, ns string,
	ids []string,
) ([]internaltypes.PolicyConfig, error) {
	return a.Store.ListPoliciesByIDs(ctx, orgID, ns, ids)
}

// GetStats gets stats.
func (a *banditStoreAdapter) GetStats(
	ctx context.Context,
	orgID string,
	ns string,
	surface string,
	bucket string,
	algo internaltypes.Algorithm,
) (map[string]internaltypes.Stats, error) {
	return a.Store.GetStats(ctx, orgID, ns, surface, bucket, algo)
}

// IncrementStats increments stats.
func (a *banditStoreAdapter) IncrementStats(
	ctx context.Context,
	orgID string,
	ns string,
	surface string,
	bucket string,
	algo internaltypes.Algorithm,
	policyID string,
	reward bool,
) error {
	return a.Store.IncrementStats(
		ctx, orgID, ns, surface, bucket, algo, policyID, reward,
	)
}

// LogDecision logs a decision.
func (a *banditStoreAdapter) LogDecision(
	ctx context.Context,
	orgID string,
	ns string,
	surface string,
	bucket string,
	algo internaltypes.Algorithm,
	policyID string,
	explore bool,
	reqID string,
	meta map[string]any,
) error {
	return a.Store.LogDecision(
		ctx, orgID, ns, surface, bucket, algo, policyID, explore, reqID, meta,
	)
}

// LogReward logs a reward.
func (a *banditStoreAdapter) LogReward(
	ctx context.Context,
	orgID, ns, surface, bucket string,
	algo internaltypes.Algorithm,
	policyID string,
	reward bool,
	reqID string,
	meta map[string]any,
) error {
	return a.Store.LogReward(
		ctx, orgID, ns, surface, bucket, algo, policyID, reward, reqID, meta,
	)
}
