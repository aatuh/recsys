package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"recsys/internal/algorithm"
	"recsys/internal/http/common"
	handlerstypes "recsys/internal/http/types"
	"recsys/internal/types"

	"github.com/go-chi/chi/v5"
)

// CreatedAfterParseError represents an error parsing the CreatedAfterISO field
type CreatedAfterParseError struct {
	Err error
}

func (e *CreatedAfterParseError) Error() string {
	return e.Err.Error()
}

func (e *CreatedAfterParseError) Unwrap() error {
	return e.Err
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
func (h *Handler) Recommend(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.RecommendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	// Convert HTTP request to algorithm request
	algoReq, err := h.convertToAlgorithmRequest(r, req)
	if err != nil {
		// Handle specific parsing errors with proper error codes
		var parseErr *CreatedAfterParseError
		if errors.As(err, &parseErr) {
			common.BadRequest(
				w, r, "invalid_created_after",
				"created_after must be RFC3339", nil,
			)
			return
		}
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	// Create algorithm engine
	config, err := h.getAlgorithmConfig(req.Overrides)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	engine := algorithm.NewEngine(config, h.Store)

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
// @Success      200      {array}  internal_http_types.ScoredItem
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
				// Return a specific error that can be handled by the caller
				return algorithm.Request{}, &CreatedAfterParseError{Err: err}
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
		ExplainLevel:   algorithm.NormalizeExplainLevel(req.ExplainLevel),
	}, nil
}

// convertToHTTPResponse converts algorithm response to HTTP response
func (h *Handler) convertToHTTPResponse(algoResp *algorithm.Response) handlerstypes.RecommendResponse {
	items := make([]handlerstypes.ScoredItem, 0, len(algoResp.Items))
	for _, item := range algoResp.Items {
		var explain *handlerstypes.ExplainBlock
		if item.Explain != nil {
			explain = mapExplainBlock(item.Explain)
		}
		items = append(items, handlerstypes.ScoredItem{
			ItemID:  item.ItemID,
			Score:   item.Score,
			Reasons: item.Reasons,
			Explain: explain,
		})
	}

	return handlerstypes.RecommendResponse{
		ModelVersion: algoResp.ModelVersion,
		Items:        items,
	}
}

// api/internal/http/handlers/recommend.go

func mapExplainBlock(src *algorithm.ExplainBlock) *handlerstypes.ExplainBlock {
	if src == nil {
		return nil
	}

	dst := &handlerstypes.ExplainBlock{}

	if src.Blend != nil {
		dst.Blend = &handlerstypes.ExplainBlend{
			Alpha:    src.Blend.Alpha,
			Beta:     src.Blend.Beta,
			Gamma:    src.Blend.Gamma,
			PopNorm:  src.Blend.PopNorm,
			CoocNorm: src.Blend.CoocNorm,
			EmbNorm:  src.Blend.EmbNorm,
			Contributions: handlerstypes.ExplainBlendContribution{
				Pop:  src.Blend.Contributions.Pop,
				Cooc: src.Blend.Contributions.Cooc,
				Emb:  src.Blend.Contributions.Emb,
			},
		}
		if src.Blend.Raw != nil {
			dst.Blend.Raw = &handlerstypes.ExplainBlendRaw{
				Pop:  src.Blend.Raw.Pop,
				Cooc: src.Blend.Raw.Cooc,
				Emb:  src.Blend.Raw.Emb,
			}
		}
	}

	if src.Personalization != nil {
		dst.Personalization = &handlerstypes.ExplainPersonalization{
			Overlap:         src.Personalization.Overlap,
			BoostMultiplier: src.Personalization.BoostMultiplier,
			Raw: func() *handlerstypes.ExplainPersonalizationRaw {
				if src.Personalization.Raw == nil {
					return nil
				}
				return &handlerstypes.ExplainPersonalizationRaw{
					ProfileBoost: src.Personalization.Raw.ProfileBoost,
				}
			}(),
		}
	}

	if src.MMR != nil {
		dst.MMR = &handlerstypes.ExplainMMR{
			Lambda:        src.MMR.Lambda,
			MaxSimilarity: src.MMR.MaxSimilarity,
			Penalty:       src.MMR.Penalty,
			Relevance:     src.MMR.Relevance,
			Rank:          src.MMR.Rank,
		}
	}

	if src.Caps != nil {
		c := &handlerstypes.ExplainCaps{}
		if src.Caps.Brand != nil {
			c.Brand = mapCapUsage(src.Caps.Brand)
		}
		if src.Caps.Category != nil {
			c.Category = mapCapUsage(src.Caps.Category)
		}
		if c.Brand != nil || c.Category != nil {
			dst.Caps = c
		}
	}

	// IMPORTANT: copy anchors; add placeholder defensively if empty.
	if len(src.Anchors) > 0 {
		dst.Anchors = append([]string(nil), src.Anchors...)
	} else {
		dst.Anchors = []string{"(no_recent_activity)"}
	}

	// Only drop the block if truly empty and no anchors.
	if dst.Blend == nil && dst.Personalization == nil &&
		dst.MMR == nil && dst.Caps == nil && len(dst.Anchors) == 0 {
		return nil
	}
	return dst
}

func mapCapUsage(src *algorithm.CapUsage) *handlerstypes.ExplainCapUsage {
	if src == nil {
		return nil
	}
	usage := &handlerstypes.ExplainCapUsage{
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

// getAlgorithmConfig creates algorithm configuration from handler config and overrides
func (h *Handler) getAlgorithmConfig(
	overrides *handlerstypes.Overrides,
) (algorithm.Config, error) {
	config := algorithm.Config{
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
		PopularityFanout:     h.PopularityFanout,
	}

	// Apply overrides if provided
	if overrides != nil {
		if overrides.BlendAlpha != nil {
			config.BlendAlpha = *overrides.BlendAlpha
		}
		if overrides.BlendBeta != nil {
			config.BlendBeta = *overrides.BlendBeta
		}
		if overrides.BlendGamma != nil {
			config.BlendGamma = *overrides.BlendGamma
		}
		if overrides.ProfileBoost != nil {
			config.ProfileBoost = *overrides.ProfileBoost
		}
		if overrides.ProfileWindowDays != nil {
			config.ProfileWindowDays = float64(*overrides.ProfileWindowDays)
		}
		if overrides.ProfileTopN != nil {
			config.ProfileTopNTags = *overrides.ProfileTopN
		}
		if overrides.MMRLambda != nil {
			config.MMRLambda = *overrides.MMRLambda
		}
		if overrides.BrandCap != nil {
			config.BrandCap = *overrides.BrandCap
		}
		if overrides.CategoryCap != nil {
			config.CategoryCap = *overrides.CategoryCap
		}
		if overrides.PopularityHalfLifeDays != nil {
			config.HalfLifeDays = float64(*overrides.PopularityHalfLifeDays)
		}
		if overrides.CoVisWindowDays != nil {
			config.CoVisWindowDays = *overrides.CoVisWindowDays
		}
		if overrides.PurchasedWindowDays != nil {
			config.PurchasedWindowDays = *overrides.PurchasedWindowDays
		}
		if overrides.RuleExcludePurchased != nil {
			config.RuleExcludePurchased = *overrides.RuleExcludePurchased
		}
		if overrides.PopularityFanout != nil {
			config.PopularityFanout = *overrides.PopularityFanout
		}
	}

	return config, nil
}
