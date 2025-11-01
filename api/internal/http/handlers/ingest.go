package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"recsys/internal/http/common"
	"recsys/internal/services/ingestion"
	specstypes "recsys/specs/types"
)

// IngestionService captures the ingestion domain contract required by the HTTP adapter.
type IngestionService interface {
	UpsertItems(ctx context.Context, orgID uuid.UUID, req specstypes.ItemsUpsertRequest) error
	UpsertUsers(ctx context.Context, orgID uuid.UUID, req specstypes.UsersUpsertRequest) error
	InsertEvents(ctx context.Context, orgID uuid.UUID, req specstypes.EventsBatchRequest) error
}

// IngestionHandler exposes ingestion routes backed by the ingestion service.
type IngestionHandler struct {
	service    IngestionService
	defaultOrg uuid.UUID
	logger     *zap.Logger
}

// NewIngestionHandler constructs a handler backed by the provided service.
func NewIngestionHandler(svc IngestionService, defaultOrg uuid.UUID, logger *zap.Logger) *IngestionHandler {
	h := &IngestionHandler{
		service:    svc,
		defaultOrg: defaultOrg,
		logger:     logger,
	}
	if h.logger == nil {
		h.logger = zap.NewNop()
	}
	return h
}

// ItemsUpsert godoc
// @Summary      Upsert items (batch)
// @Description  Create or update items by opaque IDs.
// @Tags         ingestion
// @Accept       json
// @Produce      json
// @Param        payload  body  types.ItemsUpsertRequest  true  "Items upsert"
// @Success      202      {object}  types.Ack
// @Failure      400      {object}  common.APIError
// @Router       /v1/items:upsert [post]
// @ID upsertItems
func (h *IngestionHandler) ItemsUpsert(w http.ResponseWriter, r *http.Request) {
	var req specstypes.ItemsUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}
	orgID := orgIDFromHeader(r, h.defaultOrg)

	if h.service == nil {
		common.HttpErrorWithLogger(w, r, errors.New("ingestion handler: service is nil"), http.StatusInternalServerError, h.logger)
		return
	}
	if err := h.service.UpsertItems(r.Context(), orgID, req); err != nil {
		var vErr ingestion.ValidationError
		if errors.As(err, &vErr) {
			common.BadRequest(w, r, vErr.Code, vErr.Message, vErr.Details)
			return
		}
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	writeAccepted(w)
}

// UsersUpsert godoc
// @Summary      Upsert users (batch)
// @Tags         ingestion
// @Accept       json
// @Produce      json
// @Param        payload  body  types.UsersUpsertRequest  true  "Users upsert"
// @Success      202      {object}  types.Ack
// @Failure      400      {object}  common.APIError
// @Router       /v1/users:upsert [post]
// @ID upsertUsers
func (h *IngestionHandler) UsersUpsert(w http.ResponseWriter, r *http.Request) {
	var req specstypes.UsersUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}
	orgID := orgIDFromHeader(r, h.defaultOrg)

	if h.service == nil {
		common.HttpErrorWithLogger(w, r, errors.New("ingestion handler: service is nil"), http.StatusInternalServerError, h.logger)
		return
	}
	if err := h.service.UpsertUsers(r.Context(), orgID, req); err != nil {
		var vErr ingestion.ValidationError
		if errors.As(err, &vErr) {
			common.BadRequest(w, r, vErr.Code, vErr.Message, vErr.Details)
			return
		}
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	writeAccepted(w)
}

// EventsBatch godoc
// @Summary      Ingest events (batch)
// @Tags         ingestion
// @Accept       json
// @Produce      json
// @Param        payload  body  types.EventsBatchRequest  true  "Events batch"
// @Success      202      {object}  types.Ack
// @Failure      400      {object}  common.APIError
// @Router       /v1/events:batch [post]
// @ID batchEvents
func (h *IngestionHandler) EventsBatch(w http.ResponseWriter, r *http.Request) {
	var req specstypes.EventsBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}
	orgID := orgIDFromHeader(r, h.defaultOrg)

	if h.service == nil {
		common.HttpErrorWithLogger(w, r, errors.New("ingestion handler: service is nil"), http.StatusInternalServerError, h.logger)
		return
	}
	if err := h.service.InsertEvents(r.Context(), orgID, req); err != nil {
		var vErr ingestion.ValidationError
		if errors.As(err, &vErr) {
			common.BadRequest(w, r, vErr.Code, vErr.Message, vErr.Details)
			return
		}
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	writeAccepted(w)
}

func writeAccepted(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"accepted"}`))
}
