package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"recsys/internal/algorithm"
	"recsys/internal/http/common"
	handlerstypes "recsys/internal/http/types"
	"recsys/internal/types"
)

// SegmentProfilesList godoc
// @Summary      List segment profiles
// @Tags         config
// @Produce      json
// @Param        namespace query string false "Namespace" default(default)
// @Success      200 {object} handlerstypes.SegmentProfilesListResponse
// @Failure      500 {object} common.APIError
// @Router       /v1/segment-profiles [get]
func (h *Handler) SegmentProfilesList(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		ns = "default"
	}
	orgID := h.defaultOrgFromHeader(r)

	profiles, err := h.Store.ListSegmentProfiles(r.Context(), orgID, ns)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	resp := handlerstypes.SegmentProfilesListResponse{
		Namespace: ns,
		Profiles:  make([]handlerstypes.SegmentProfile, 0, len(profiles)),
	}
	for _, profile := range profiles {
		resp.Profiles = append(resp.Profiles, mapSegmentProfile(profile))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// SegmentProfilesUpsert godoc
// @Summary      Upsert segment profiles
// @Tags         config
// @Accept       json
// @Produce      json
// @Param        payload body handlerstypes.SegmentProfilesUpsertRequest true "Profiles"
// @Success      200 {object} handlerstypes.Ack
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/segment-profiles:upsert [post]
func (h *Handler) SegmentProfilesUpsert(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.SegmentProfilesUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if len(req.Profiles) == 0 {
		common.BadRequest(w, r, "empty_profiles", "profiles must not be empty", nil)
		return
	}

	orgID := h.defaultOrgFromHeader(r)
	internalProfiles := make([]types.SegmentProfile, 0, len(req.Profiles))
	for _, profile := range req.Profiles {
		if profile.ProfileID == "" {
			common.BadRequest(w, r, "missing_profile_id", "profile_id is required", nil)
			return
		}
		internalProfiles = append(internalProfiles, mapSegmentProfileToInternal(profile))
	}

	if err := h.Store.UpsertSegmentProfiles(r.Context(), orgID, req.Namespace, internalProfiles); err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(handlerstypes.Ack{Status: "ok"})
}

// SegmentProfilesDelete godoc
// @Summary      Delete segment profiles
// @Tags         config
// @Accept       json
// @Produce      json
// @Param        payload body handlerstypes.IDListRequest true "IDs"
// @Success      200 {object} handlerstypes.Ack
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/segment-profiles:delete [post]
func (h *Handler) SegmentProfilesDelete(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.IDListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if len(req.IDs) == 0 {
		common.BadRequest(w, r, "empty_ids", "ids must not be empty", nil)
		return
	}

	orgID := h.defaultOrgFromHeader(r)
	if _, err := h.Store.DeleteSegmentProfiles(r.Context(), orgID, req.Namespace, req.IDs); err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(handlerstypes.Ack{Status: "ok"})
}

// SegmentsList godoc
// @Summary      List segments with rules
// @Tags         config
// @Produce      json
// @Param        namespace query string false "Namespace" default(default)
// @Success      200 {object} handlerstypes.SegmentsListResponse
// @Failure      500 {object} common.APIError
// @Router       /v1/segments [get]
func (h *Handler) SegmentsList(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		ns = "default"
	}
	orgID := h.defaultOrgFromHeader(r)

	segments, err := h.Store.ListSegmentsWithRules(r.Context(), orgID, ns)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	resp := handlerstypes.SegmentsListResponse{
		Namespace: ns,
		Segments:  make([]handlerstypes.Segment, 0, len(segments)),
	}
	for _, seg := range segments {
		resp.Segments = append(resp.Segments, mapSegment(seg))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// SegmentsUpsert godoc
// @Summary      Upsert a segment and its rules
// @Tags         config
// @Accept       json
// @Produce      json
// @Param        payload body handlerstypes.SegmentsUpsertRequest true "Segment"
// @Success      200 {object} handlerstypes.Ack
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/segments:upsert [post]
func (h *Handler) SegmentsUpsert(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.SegmentsUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}

	seg := req.Segment
	if seg.SegmentID == "" {
		common.BadRequest(w, r, "missing_segment_id", "segment_id is required", nil)
		return
	}
	if seg.ProfileID == "" {
		common.BadRequest(w, r, "missing_profile_id", "profile_id is required", nil)
		return
	}

	internalSeg := types.Segment{
		SegmentID:   seg.SegmentID,
		Name:        seg.Name,
		Priority:    seg.Priority,
		Active:      seg.Active,
		ProfileID:   seg.ProfileID,
		Description: seg.Description,
	}
	for _, rule := range seg.Rules {
		if !json.Valid(rule.Rule) {
			common.BadRequest(w, r, "invalid_rule", "rule must be valid JSON", nil)
			return
		}
		internalSeg.Rules = append(internalSeg.Rules, types.SegmentRule{
			RuleID:      valueOrZero(rule.RuleID),
			Rule:        rule.Rule,
			Enabled:     rule.Enabled,
			Description: rule.Description,
		})
	}

	orgID := h.defaultOrgFromHeader(r)
	if err := h.Store.UpsertSegmentWithRules(r.Context(), orgID, req.Namespace, internalSeg); err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(handlerstypes.Ack{Status: "ok"})
}

// SegmentsDelete godoc
// @Summary      Delete segments
// @Tags         config
// @Accept       json
// @Produce      json
// @Param        payload body handlerstypes.IDListRequest true "IDs"
// @Success      200 {object} handlerstypes.Ack
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/segments:delete [post]
func (h *Handler) SegmentsDelete(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.IDListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if len(req.IDs) == 0 {
		common.BadRequest(w, r, "empty_ids", "ids must not be empty", nil)
		return
	}

	orgID := h.defaultOrgFromHeader(r)
	if _, err := h.Store.DeleteSegments(r.Context(), orgID, req.Namespace, req.IDs); err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(handlerstypes.Ack{Status: "ok"})
}

// SegmentsDryRun godoc
// @Summary      Simulate segment selection for context
// @Tags         config
// @Accept       json
// @Produce      json
// @Param        payload body handlerstypes.SegmentDryRunRequest true "Dry run"
// @Success      200 {object} handlerstypes.SegmentDryRunResponse
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/segments:dry-run [post]
func (h *Handler) SegmentsDryRun(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.SegmentDryRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}

	orgID := h.defaultOrgFromHeader(r)
	algoReq := algorithm.Request{
		OrgID:     orgID,
		UserID:    req.UserID,
		Namespace: req.Namespace,
	}
	_, segmentID, profileID, ruleID, err := h.selectSegmentProfile(r.Context(), algoReq, handlerstypes.RecommendRequest{Namespace: req.Namespace, Context: req.Context}, req.Traits)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	matched := segmentID != ""
	resp := handlerstypes.SegmentDryRunResponse{
		Matched:   matched,
		SegmentID: segmentID,
		ProfileID: profileID,
	}
	if ruleID != 0 {
		resp.MatchedRuleID = &ruleID
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func mapSegmentProfile(src types.SegmentProfile) handlerstypes.SegmentProfile {
	return handlerstypes.SegmentProfile{
		ProfileID:           src.ProfileID,
		Description:         src.Description,
		BlendAlpha:          src.BlendAlpha,
		BlendBeta:           src.BlendBeta,
		BlendGamma:          src.BlendGamma,
		MMRLambda:           src.MMRLambda,
		BrandCap:            src.BrandCap,
		CategoryCap:         src.CategoryCap,
		ProfileBoost:        src.ProfileBoost,
		ProfileWindowDays:   src.ProfileWindowDays,
		ProfileTopN:         src.ProfileTopN,
		HalfLifeDays:        src.HalfLifeDays,
		CoVisWindowDays:     src.CoVisWindowDays,
		PurchasedWindowDays: src.PurchasedWindowDays,
		RuleExcludeEvents:   src.RuleExcludeEvents,
		ExcludeEventTypes:   append([]int16(nil), src.ExcludeEventTypes...),
		BrandTagPrefixes:    append([]string(nil), src.BrandTagPrefixes...),
		CategoryTagPrefixes: append([]string(nil), src.CategoryTagPrefixes...),
		PopularityFanout:    src.PopularityFanout,
		CreatedAt:           ptrTime(src.CreatedAt),
		UpdatedAt:           ptrTime(src.UpdatedAt),
	}
}

func mapSegmentProfileToInternal(src handlerstypes.SegmentProfile) types.SegmentProfile {
	return types.SegmentProfile{
		ProfileID:           src.ProfileID,
		Description:         src.Description,
		BlendAlpha:          src.BlendAlpha,
		BlendBeta:           src.BlendBeta,
		BlendGamma:          src.BlendGamma,
		MMRLambda:           src.MMRLambda,
		BrandCap:            src.BrandCap,
		CategoryCap:         src.CategoryCap,
		ProfileBoost:        src.ProfileBoost,
		ProfileWindowDays:   src.ProfileWindowDays,
		ProfileTopN:         src.ProfileTopN,
		HalfLifeDays:        src.HalfLifeDays,
		CoVisWindowDays:     src.CoVisWindowDays,
		PurchasedWindowDays: src.PurchasedWindowDays,
		RuleExcludeEvents:   src.RuleExcludeEvents,
		ExcludeEventTypes:   append([]int16(nil), src.ExcludeEventTypes...),
		BrandTagPrefixes:    append([]string(nil), src.BrandTagPrefixes...),
		CategoryTagPrefixes: append([]string(nil), src.CategoryTagPrefixes...),
		PopularityFanout:    src.PopularityFanout,
	}
}

func mapSegment(src types.Segment) handlerstypes.Segment {
	out := handlerstypes.Segment{
		SegmentID:   src.SegmentID,
		Name:        src.Name,
		Priority:    src.Priority,
		Active:      src.Active,
		ProfileID:   src.ProfileID,
		Description: src.Description,
		CreatedAt:   ptrTime(src.CreatedAt),
		UpdatedAt:   ptrTime(src.UpdatedAt),
	}
	for _, rule := range src.Rules {
		copyRule := handlerstypes.SegmentRule{
			Rule:        append(json.RawMessage(nil), rule.Rule...),
			Enabled:     rule.Enabled,
			Description: rule.Description,
		}
		if rule.RuleID != 0 {
			copyRule.RuleID = &rule.RuleID
		}
		out.Rules = append(out.Rules, copyRule)
	}
	return out
}

func ptrTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func valueOrZero(id *int64) int64 {
	if id == nil {
		return 0
	}
	return *id
}
