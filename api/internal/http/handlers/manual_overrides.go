package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aatuh/recsys-algo/rules"
	"recsys/internal/http/common"
	"recsys/internal/services/manual"
	"recsys/internal/types"
	specstypes "recsys/specs/types"
)

// ManualOverridesHandler exposes endpoints for manual merchandising overrides.
type ManualOverridesHandler struct {
	service      *manual.Service
	rulesManager *rules.Manager
	defaultOrg   uuid.UUID
}

var errServiceUnavailable = errors.New("manual override service unavailable")

// NewManualOverridesHandler constructs the handler.
func NewManualOverridesHandler(service *manual.Service, manager *rules.Manager, defaultOrg uuid.UUID) *ManualOverridesHandler {
	return &ManualOverridesHandler{service: service, rulesManager: manager, defaultOrg: defaultOrg}
}

// ManualOverrideCreate godoc
// @Summary Create a manual boost/suppression override
// @Tags admin
// @Accept json
// @Produce json
// @Param payload body types.ManualOverrideRequest true "Manual override payload"
// @Success 201 {object} types.ManualOverrideResponse
// @Failure 400 {object} common.APIError
// @Failure 500 {object} common.APIError
// @Router /v1/admin/manual_overrides [post]
func (h *ManualOverridesHandler) ManualOverrideCreate(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		common.HttpError(w, r, errServiceUnavailable, http.StatusServiceUnavailable)
		return
	}
	var payload specstypes.ManualOverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	action := strings.ToLower(strings.TrimSpace(payload.Action))
	var actionEnum types.ManualOverrideAction
	switch action {
	case string(types.ManualOverrideActionBoost):
		actionEnum = types.ManualOverrideActionBoost
	case string(types.ManualOverrideActionSuppress):
		actionEnum = types.ManualOverrideActionSuppress
	default:
		common.BadRequest(w, r, "invalid_action", "action must be 'boost' or 'suppress'", nil)
		return
	}

	var expiresAt *time.Time
	if payload.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, payload.ExpiresAt)
		if err != nil {
			common.BadRequest(w, r, "invalid_expires_at", "expires_at must be RFC3339", nil)
			return
		}
		expiresAt = &parsed
	}

	var boostPtr *float64
	if payload.BoostValue != nil {
		val := *payload.BoostValue
		boostPtr = &val
	}

	input := manual.CreateInput{
		Namespace:  payload.Namespace,
		Surface:    payload.Surface,
		ItemID:     payload.ItemID,
		Action:     actionEnum,
		BoostValue: boostPtr,
		Notes:      payload.Notes,
		CreatedBy:  payload.CreatedBy,
		ExpiresAt:  expiresAt,
	}
	if payload.Priority != nil {
		input.Priority = payload.Priority
	}

	orig := orgIDFromHeader(r, h.defaultOrg)
	record, err := h.service.Create(r.Context(), orig, input)
	if err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	if h.rulesManager != nil {
		h.rulesManager.Invalidate(record.OrgID, record.Namespace, record.Surface)
	}

	resp := toManualOverrideResponse(*record)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// ManualOverrideList godoc
// @Summary List manual overrides
// @Tags admin
// @Produce json
// @Param namespace query string true "Namespace"
// @Param surface query string false "Surface"
// @Param status query string false "Filter by status (active,cancelled,expired)"
// @Param action query string false "Filter by action (boost,suppress)"
// @Param include_expired query boolean false "Include expired overrides"
// @Success 200 {array} types.ManualOverrideResponse
// @Failure 400 {object} common.APIError
// @Failure 500 {object} common.APIError
// @Router /v1/admin/manual_overrides [get]
func (h *ManualOverridesHandler) ManualOverrideList(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		common.HttpError(w, r, errServiceUnavailable, http.StatusServiceUnavailable)
		return
	}
	q := r.URL.Query()
	namespace := strings.TrimSpace(q.Get("namespace"))
	if namespace == "" {
		common.BadRequest(w, r, "missing_namespace", "namespace is required", nil)
		return
	}
	surface := strings.TrimSpace(q.Get("surface"))
	includeExpired := strings.EqualFold(q.Get("include_expired"), "true")

	var status types.ManualOverrideStatus
	if v := strings.TrimSpace(q.Get("status")); v != "" {
		switch strings.ToLower(v) {
		case string(types.ManualOverrideStatusActive):
			status = types.ManualOverrideStatusActive
		case string(types.ManualOverrideStatusCancelled):
			status = types.ManualOverrideStatusCancelled
		case string(types.ManualOverrideStatusExpired):
			status = types.ManualOverrideStatusExpired
			includeExpired = true
		default:
			common.BadRequest(w, r, "invalid_status", "status must be active, cancelled, or expired", nil)
			return
		}
	}

	var action types.ManualOverrideAction
	if v := strings.TrimSpace(q.Get("action")); v != "" {
		switch strings.ToLower(v) {
		case string(types.ManualOverrideActionBoost):
			action = types.ManualOverrideActionBoost
		case string(types.ManualOverrideActionSuppress):
			action = types.ManualOverrideActionSuppress
		default:
			common.BadRequest(w, r, "invalid_action", "action must be boost or suppress", nil)
			return
		}
	}

	filters := types.ManualOverrideFilters{
		Status:         status,
		Action:         action,
		IncludeExpired: includeExpired,
	}

	orig := orgIDFromHeader(r, h.defaultOrg)
	records, err := h.service.List(r.Context(), orig, namespace, surface, filters)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	resp := make([]specstypes.ManualOverrideResponse, 0, len(records))
	for _, rec := range records {
		resp = append(resp, toManualOverrideResponse(rec))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ManualOverrideCancel godoc
// @Summary Cancel a manual override
// @Tags admin
// @Accept json
// @Produce json
// @Param override_id path string true "Override ID"
// @Param payload body types.ManualOverrideCancelRequest false "Optional cancellation metadata"
// @Success 200 {object} types.ManualOverrideResponse
// @Failure 404 {object} common.APIError
// @Failure 400 {object} common.APIError
// @Failure 500 {object} common.APIError
// @Router /v1/admin/manual_overrides/{override_id}/cancel [post]
func (h *ManualOverridesHandler) ManualOverrideCancel(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		common.HttpError(w, r, errServiceUnavailable, http.StatusServiceUnavailable)
		return
	}
	overrideParam := chi.URLParam(r, "override_id")
	overrideID, err := uuid.Parse(overrideParam)
	if err != nil {
		common.BadRequest(w, r, "invalid_override_id", "override_id must be a UUID", nil)
		return
	}

	var payload specstypes.ManualOverrideCancelRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			common.HttpError(w, r, err, http.StatusBadRequest)
			return
		}
	}
	cancelledBy := strings.TrimSpace(payload.CancelledBy)
	if cancelledBy == "" {
		cancelledBy = strings.TrimSpace(r.Header.Get("X-User-Email"))
	}

	orig := orgIDFromHeader(r, h.defaultOrg)
	record, err := h.service.Cancel(r.Context(), orig, overrideID, cancelledBy)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	if record == nil {
		writeAPIError(w, r, http.StatusNotFound, "override_not_found", "override not found or already inactive")
		return
	}

	if h.rulesManager != nil {
		h.rulesManager.Invalidate(record.OrgID, record.Namespace, record.Surface)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toManualOverrideResponse(*record))
}

func toManualOverrideResponse(src types.ManualOverride) specstypes.ManualOverrideResponse {
	resp := specstypes.ManualOverrideResponse{
		OverrideID: src.OverrideID.String(),
		Namespace:  src.Namespace,
		Surface:    src.Surface,
		ItemID:     src.ItemID,
		Action:     string(src.Action),
		Notes:      src.Notes,
		CreatedBy:  src.CreatedBy,
		CreatedAt:  src.CreatedAt.Format(time.RFC3339),
		Status:     string(src.Status),
	}
	if src.BoostValue != nil {
		resp.BoostValue = src.BoostValue
	}
	if src.ExpiresAt != nil {
		resp.ExpiresAt = src.ExpiresAt.Format(time.RFC3339)
	}
	if src.RuleID != nil {
		resp.RuleID = src.RuleID.String()
	}
	if src.CancelledAt != nil {
		resp.CancelledAt = src.CancelledAt.Format(time.RFC3339)
	}
	if src.CancelledBy != "" {
		resp.CancelledBy = src.CancelledBy
	}
	return resp
}
