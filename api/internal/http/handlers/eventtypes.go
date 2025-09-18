package handlers

import (
	"encoding/json"
	"net/http"
	"recsys/internal/http/common"
	"recsys/internal/store"
	"recsys/specs/types"

	"github.com/google/uuid"
)

// @Summary Upsert tenant event-type config
// @Tags config
// @Accept json
// @Produce json
// @Param payload body types.EventTypeConfigUpsertRequest true "Event types"
// @Success 202 {object} types.Ack
// @Router /v1/event-types:upsert [post]
// @ID upsertEventTypes
func (h *Handler) EventTypesUpsert(w http.ResponseWriter, r *http.Request) {
	var req types.EventTypeConfigUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	orgID := h.defaultOrgFromHeader(r)
	rows := make([]store.EventTypeConfig, 0, len(req.Types))
	for _, t := range req.Types {
		if t.Weight <= 0 {
			common.Unprocessable(w, r, "invalid_weight", "weight must be > 0", nil)
			return
		}
		if t.HalfLifeDays != nil && *t.HalfLifeDays <= 0 {
			common.Unprocessable(w, r, "invalid_half_life", "half_life_days must be > 0", nil)
			return
		}
		rows = append(rows, store.EventTypeConfig{
			Type:         t.Type,
			Name:         t.Name,
			Weight:       t.Weight,
			HalfLifeDays: t.HalfLifeDays,
			IsActive:     t.IsActive,
		})
	}
	if err := h.Store.UpsertEventTypeConfig(r.Context(), orgID, req.Namespace, rows); err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"accepted"}`))
}

// @Summary List effective event-type config
// @Tags config
// @Produce json
// @Param namespace query string true "Namespace"
// @Success 200 {array} types.EventTypeConfigUpsertResponse
// @Router /v1/event-types [get]
func (h *Handler) EventTypesList(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		common.BadRequest(w, r, "missing_namespace", "namespace is required", nil)
		return
	}
	orgID := h.defaultOrgFromHeader(r)
	rows, err := h.Store.ListEventTypeConfigEffective(r.Context(), orgID, ns)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	out := make([]types.EventTypeConfigUpsertResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.EventTypeConfigUpsertResponse{
			Type: r.Type, Name: r.Name, Weight: r.Weight, HalfLifeDays: r.HalfLifeDays, IsActive: r.IsActive, Source: r.Source,
		})
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *Handler) defaultOrgFromHeader(r *http.Request) uuid.UUID {
	org := r.Header.Get("X-Org-ID")
	if id, err := uuid.Parse(org); err == nil {
		return id
	}
	return h.DefaultOrg
}
