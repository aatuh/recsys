package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"recsys/internal/http/common"
	"recsys/internal/store"
	"recsys/internal/types"
	handlerstypes "recsys/specs/types"

	"github.com/aatuh/recsys-algo/algorithm"

	"github.com/google/uuid"
)

// SegmentsHandler exposes segment and segment-profile management endpoints.
type SegmentsHandler struct {
	store      *store.Store
	defaultOrg uuid.UUID
}

// NewSegmentsHandler constructs an HTTP handler for segment operations.
func NewSegmentsHandler(st *store.Store, defaultOrg uuid.UUID) *SegmentsHandler {
	return &SegmentsHandler{store: st, defaultOrg: defaultOrg}
}

// SegmentProfilesList godoc
// @Summary      List segment profiles
// @Tags         config
// @Produce      json
// @Param        namespace query string false "Namespace" default(default)
// @Success      200 {object} types.SegmentProfilesListResponse
// @Failure      500 {object} common.APIError
// @Router       /v1/segment-profiles [get]
func (h *SegmentsHandler) SegmentProfilesList(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		ns = "default"
	}
	orgID := orgIDFromHeader(r, h.defaultOrg)

	profiles, err := h.store.ListSegmentProfiles(r.Context(), orgID, ns)
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
// @Param        payload body types.SegmentProfilesUpsertRequest true "Profiles"
// @Success      200 {object} types.Ack
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/segment-profiles:upsert [post]
// @ID segmentProfilesUpsert
func (h *SegmentsHandler) SegmentProfilesUpsert(w http.ResponseWriter, r *http.Request) {
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

	orgID := orgIDFromHeader(r, h.defaultOrg)
	internalProfiles := make([]types.SegmentProfile, 0, len(req.Profiles))
	for _, profile := range req.Profiles {
		if profile.ProfileID == "" {
			common.BadRequest(w, r, "missing_profile_id", "profile_id is required", nil)
			return
		}
		internalProfiles = append(internalProfiles, mapSegmentProfileToInternal(profile))
	}

	if err := h.store.UpsertSegmentProfiles(r.Context(), orgID, req.Namespace, internalProfiles); err != nil {
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
// @Param        payload body types.IDListRequest true "IDs"
// @Success      200 {object} types.Ack
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/segment-profiles:delete [post]
// @ID segmentProfilesDelete
func (h *SegmentsHandler) SegmentProfilesDelete(w http.ResponseWriter, r *http.Request) {
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

	orgID := orgIDFromHeader(r, h.defaultOrg)
	if _, err := h.store.DeleteSegmentProfiles(r.Context(), orgID, req.Namespace, req.IDs); err != nil {
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
// @Success      200 {object} types.SegmentsListResponse
// @Failure      500 {object} common.APIError
// @Router       /v1/segments [get]
func (h *SegmentsHandler) SegmentsList(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		ns = "default"
	}
	orgID := orgIDFromHeader(r, h.defaultOrg)

	segments, err := h.store.ListSegmentsWithRules(r.Context(), orgID, ns)
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
// @Param        payload body types.SegmentsUpsertRequest true "Segment"
// @Success      200 {object} types.Ack
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/segments:upsert [post]
// @ID segmentsUpsert
func (h *SegmentsHandler) SegmentsUpsert(w http.ResponseWriter, r *http.Request) {
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

	orgID := orgIDFromHeader(r, h.defaultOrg)
	if err := h.store.UpsertSegmentWithRules(r.Context(), orgID, req.Namespace, internalSeg); err != nil {
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
// @Param        payload body types.IDListRequest true "IDs"
// @Success      200 {object} types.Ack
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/segments:delete [post]
// @ID segmentsDelete
func (h *SegmentsHandler) SegmentsDelete(w http.ResponseWriter, r *http.Request) {
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

	orgID := orgIDFromHeader(r, h.defaultOrg)
	if _, err := h.store.DeleteSegments(r.Context(), orgID, req.Namespace, req.IDs); err != nil {
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
// @Param        payload body types.SegmentDryRunRequest true "Dry run"
// @Success      200 {object} types.SegmentDryRunResponse
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/segments:dry-run [post]
// @ID segmentsDryRun
func (h *SegmentsHandler) SegmentsDryRun(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.SegmentDryRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}

	orgID := orgIDFromHeader(r, h.defaultOrg)
	algoReq := algorithm.Request{
		OrgID:     orgID,
		UserID:    req.UserID,
		Namespace: req.Namespace,
	}
	sel, ruleID, err := resolveSegmentSelection(r.Context(), h.store, algoReq, handlerstypes.RecommendRequest{Namespace: req.Namespace, Context: req.Context}, req.Traits)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	matched := sel.SegmentID != ""
	resp := handlerstypes.SegmentDryRunResponse{
		Matched:   matched,
		SegmentID: sel.SegmentID,
		ProfileID: sel.ProfileID,
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
