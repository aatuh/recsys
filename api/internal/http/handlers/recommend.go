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
	handlerstypes "recsys/internal/http/types"
	"recsys/internal/segments"
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
	start := time.Now()
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
	config := h.baseAlgorithmConfig()
	segmentProfile, segmentID, profileID, _, err := h.selectSegmentProfile(r.Context(), algoReq, req, nil)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	if segmentProfile != nil {
		applySegmentProfile(&config, *segmentProfile)
		if profileID == "" {
			profileID = segmentProfile.ProfileID
		}
	}
	applyOverrides(&config, req.Overrides)
	engine := algorithm.NewEngine(config, h.Store)

	// Get recommendations
	algoResp, traceData, err := engine.Recommend(r.Context(), algoReq)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	algoResp.SegmentID = segmentID
	algoResp.ProfileID = profileID

	// Convert algorithm response to HTTP response
	httpResp := h.convertToHTTPResponse(algoResp)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(httpResp)

	h.recordDecisionTrace(decisionTraceInput{
		Request:      r,
		HTTPRequest:  req,
		AlgoRequest:  algoReq,
		Config:       config,
		AlgoResponse: algoResp,
		HTTPResponse: httpResp,
		TraceData:    traceData,
		Duration:     time.Since(start),
	})
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
	ns := req.Namespace
	if ns == "" {
		ns = "default"
	}

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
		Namespace:      ns,
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
		SegmentID:    algoResp.SegmentID,
		ProfileID:    algoResp.ProfileID,
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

func (h *Handler) baseAlgorithmConfig() algorithm.Config {
	return algorithm.Config{
		BlendAlpha:          h.BlendAlpha,
		BlendBeta:           h.BlendBeta,
		BlendGamma:          h.BlendGamma,
		ProfileBoost:        h.ProfileBoost,
		ProfileWindowDays:   h.ProfileWindowDays,
		ProfileTopNTags:     h.ProfileTopNTags,
		MMRLambda:           h.MMRLambda,
		BrandCap:            h.BrandCap,
		CategoryCap:         h.CategoryCap,
		HalfLifeDays:        h.HalfLifeDays,
		CoVisWindowDays:     int(h.CoVisWindowDays),
		PurchasedWindowDays: int(h.PurchasedWindowDays),
		RuleExcludeEvents:   h.RuleExcludeEvents,
		ExcludeEventTypes:   append([]int16(nil), h.ExcludeEventTypes...),
		BrandTagPrefixes:    append([]string(nil), h.BrandTagPrefixes...),
		CategoryTagPrefixes: append([]string(nil), h.CategoryTagPrefixes...),
		PopularityFanout:    h.PopularityFanout,
	}
}

func applyOverrides(cfg *algorithm.Config, overrides *handlerstypes.Overrides) {
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

func (h *Handler) selectSegmentProfile(
	ctx context.Context,
	req algorithm.Request,
	httpReq handlerstypes.RecommendRequest,
	traitsOverride map[string]any,
) (*types.SegmentProfile, string, string, int64, error) {
	segmentsList, err := h.Store.ListActiveSegmentsWithRules(ctx, req.OrgID, req.Namespace)
	if err != nil {
		return nil, "", "", 0, err
	}
	if len(segmentsList) == 0 {
		return nil, "", "", 0, nil
	}

	traits := traitsOverride
	if traits == nil && req.UserID != "" {
		userRec, err := h.Store.GetUser(ctx, req.OrgID, req.Namespace, req.UserID)
		if err != nil {
			return nil, "", "", 0, err
		}
		if userRec != nil && userRec.Traits != nil {
			traits = userRec.Traits
		}
	}

	data := buildSegmentContextData(req, httpReq.Context, traits)
	now := time.Now().UTC()

	var defaultSegment *types.Segment
	for i := range segmentsList {
		seg := segmentsList[i]
		if seg.SegmentID == "default" {
			defaultSegment = &segmentsList[i]
		}
		matchedRule, ok := segmentMatches(&seg, data, now)
		if ok {
			profile, err := h.Store.GetSegmentProfile(ctx, req.OrgID, req.Namespace, seg.ProfileID)
			if err != nil {
				return nil, "", "", 0, err
			}
			profileID := ""
			if profile != nil {
				profileID = profile.ProfileID
			}
			var ruleID int64
			if matchedRule != nil {
				ruleID = matchedRule.RuleID
			}
			return profile, seg.SegmentID, profileID, ruleID, nil
		}
	}

	if defaultSegment != nil {
		profile, err := h.Store.GetSegmentProfile(ctx, req.OrgID, req.Namespace, defaultSegment.ProfileID)
		if err != nil {
			return nil, "", "", 0, err
		}
		profileID := ""
		if profile != nil {
			profileID = profile.ProfileID
		}
		return profile, defaultSegment.SegmentID, profileID, 0, nil
	}

	return nil, "", "", 0, nil
}

func buildSegmentContextData(
	req algorithm.Request,
	ctxValues map[string]any,
	traits map[string]any,
) map[string]any {
	userData := map[string]any{
		"id": req.UserID,
	}
	if traits != nil {
		userData["traits"] = traits
	}
	ctxData := map[string]any{}
	for k, v := range ctxValues {
		ctxData[k] = v
	}
	requestData := map[string]any{
		"namespace": req.Namespace,
		"k":         req.K,
	}
	if req.Blend != nil {
		requestData["blend"] = map[string]any{
			"pop":  req.Blend.Pop,
			"cooc": req.Blend.Cooc,
			"als":  req.Blend.ALS,
		}
	}
	return map[string]any{
		"user":    userData,
		"ctx":     ctxData,
		"request": requestData,
	}
}

func segmentMatches(seg *types.Segment, data map[string]any, now time.Time) (*types.SegmentRule, bool) {
	if seg == nil {
		return nil, false
	}
	if len(seg.Rules) == 0 {
		return nil, true
	}
	for i := range seg.Rules {
		rule := seg.Rules[i]
		if !rule.Enabled {
			continue
		}
		eval := segments.NewEvaluator(data, now)
		matched, err := eval.Match(rule.Rule)
		if err != nil {
			continue
		}
		if matched {
			return &seg.Rules[i], true
		}
	}
	return nil, false
}
