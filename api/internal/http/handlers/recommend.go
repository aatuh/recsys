package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"recsys/internal/algorithm"
	"recsys/internal/http/common"
	handlerstypes "recsys/internal/http/types"
	"recsys/internal/types"

	"github.com/go-chi/chi/v5"
)

// Recommend godoc
// @Summary      Get recommendations for a user
// @Tags         ranking
// @Accept       json
// @Produce      json
// @Param        payload  body  types.RecommendRequest  true  "Recommend request"
// @Success      200      {object}  types.RecommendResponse
// @Failure      400      {object}  common.APIError
// @Router       /v1/recommendations [post]
func (h *Handler) Recommend(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.RecommendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	// Convert HTTP request to algorithm request
	algoReq, err := h.convertToAlgorithmRequest(r, req)
	if err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	// Create algorithm engine
	engine := algorithm.NewEngine(h.getAlgorithmConfig(), h.Store)

	// Get recommendations
	algoResp, err := engine.Recommend(r.Context(), algoReq)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	// Convert algorithm response to HTTP response
	httpResp := h.convertToHTTPResponse(algoResp)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(httpResp)
}

// ItemSimilar godoc
// @Summary      Get similar items
// @Tags         ranking
// @Produce      json
// @Param        item_id  path  string  true  "Item ID"
// @Param        namespace query string false "Namespace"  default(default)
// @Param        k        query int     false "Top-K"  default(20)
// @Success      200      {array}  types.ScoredItem
// @Failure      400      {object} common.APIError
// @Router       /v1/items/{item_id}/similar [get]
func (h *Handler) ItemSimilar(w http.ResponseWriter, r *http.Request) {
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
	orgID := h.defaultOrgFromHeader(r)

	// Create similar items engine
	similarEngine := algorithm.NewSimilarItemsEngine(h.Store, int(h.CoVisWindowDays))

	// Create algorithm request
	algoReq := algorithm.SimilarItemsRequest{
		OrgID:     orgID,
		ItemID:    itemID,
		Namespace: ns,
		K:         k,
	}

	// Get similar items
	algoResp, err := similarEngine.FindSimilar(r.Context(), algoReq)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	// Convert to HTTP response format
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

// convertToAlgorithmRequest converts HTTP request to algorithm request
func (h *Handler) convertToAlgorithmRequest(r *http.Request, req handlerstypes.RecommendRequest) (algorithm.Request, error) {
	orgID := h.defaultOrgFromHeader(r)

	// Convert constraints
	var constraints *types.PopConstraints
	if req.Constraints != nil {
		constraints = &types.PopConstraints{}
		if len(req.Constraints.PriceBetween) >= 1 {
			v := req.Constraints.PriceBetween[0]
			constraints.MinPrice = &v
		}
		if len(req.Constraints.PriceBetween) >= 2 {
			v := req.Constraints.PriceBetween[1]
			constraints.MaxPrice = &v
		}
		if req.Constraints.CreatedAfterISO != "" {
			ts, err := time.Parse(time.RFC3339, req.Constraints.CreatedAfterISO)
			if err != nil {
				return algorithm.Request{}, err
			}
			constraints.CreatedAfter = &ts
		}
		constraints.IncludeTagsAny = req.Constraints.IncludeTagsAny
		constraints.ExcludeItemIDs = req.Constraints.ExcludeItemIDs
	}

	// Convert blend weights
	var blend *algorithm.BlendWeights
	if req.Blend != nil {
		blend = &algorithm.BlendWeights{
			Pop:  req.Blend.Pop,
			Cooc: req.Blend.Cooc,
			ALS:  req.Blend.ALS,
		}
	}

	return algorithm.Request{
		OrgID:          orgID,
		UserID:         req.UserID,
		Namespace:      req.Namespace,
		K:              req.K,
		Constraints:    constraints,
		Blend:          blend,
		IncludeReasons: req.IncludeReasons,
	}, nil
}

// convertToHTTPResponse converts algorithm response to HTTP response
func (h *Handler) convertToHTTPResponse(algoResp *algorithm.Response) handlerstypes.RecommendResponse {
	items := make([]handlerstypes.ScoredItem, 0, len(algoResp.Items))
	for _, item := range algoResp.Items {
		items = append(items, handlerstypes.ScoredItem{
			ItemID:  item.ItemID,
			Score:   item.Score,
			Reasons: item.Reasons,
		})
	}

	return handlerstypes.RecommendResponse{
		ModelVersion: algoResp.ModelVersion,
		Items:        items,
	}
}

// getAlgorithmConfig creates algorithm configuration from handler config
func (h *Handler) getAlgorithmConfig() algorithm.Config {
	return algorithm.Config{
		BlendAlpha:           h.BlendAlpha,
		BlendBeta:            h.BlendBeta,
		BlendGamma:           h.BlendGamma,
		ProfileBoost:         h.ProfileBoost,
		ProfileWindowDays:    h.ProfileWindowDays,
		ProfileTopNTags:      h.ProfileTopNTags,
		MMRLambda:            h.MMRLambda,
		BrandCap:             h.BrandCap,
		CategoryCap:          h.CategoryCap,
		HalfLifeDays:         h.HalfLifeDays,
		CoVisWindowDays:      int(h.CoVisWindowDays),
		PurchasedWindowDays:  int(h.PurchasedWindowDays),
		RuleExcludePurchased: h.RuleExcludePurchased,
	}
}
