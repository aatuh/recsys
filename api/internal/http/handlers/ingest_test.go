package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"recsys/internal/http/common"
	"recsys/internal/services/ingestion"
	specstypes "recsys/specs/types"
)

type stubIngestionService struct {
	itemsErr  error
	usersErr  error
	eventsErr error

	itemsCalls  int
	usersCalls  int
	eventsCalls int

	itemsOrg  uuid.UUID
	usersOrg  uuid.UUID
	eventsOrg uuid.UUID

	itemsReq  specstypes.ItemsUpsertRequest
	usersReq  specstypes.UsersUpsertRequest
	eventsReq specstypes.EventsBatchRequest
}

func (s *stubIngestionService) UpsertItems(_ context.Context, orgID uuid.UUID, req specstypes.ItemsUpsertRequest) error {
	s.itemsCalls++
	s.itemsOrg = orgID
	s.itemsReq = req
	return s.itemsErr
}

func (s *stubIngestionService) UpsertUsers(_ context.Context, orgID uuid.UUID, req specstypes.UsersUpsertRequest) error {
	s.usersCalls++
	s.usersOrg = orgID
	s.usersReq = req
	return s.usersErr
}

func (s *stubIngestionService) InsertEvents(_ context.Context, orgID uuid.UUID, req specstypes.EventsBatchRequest) error {
	s.eventsCalls++
	s.eventsOrg = orgID
	s.eventsReq = req
	return s.eventsErr
}

func TestIngestionHandler_ItemsUpsert_Success(t *testing.T) {
	t.Parallel()

	defaultOrg := uuid.New()
	svc := &stubIngestionService{}
	handler := NewIngestionHandler(svc, defaultOrg, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/items:upsert", strings.NewReader(`{"namespace":"default","items":[{"item_id":"sku-1"}]}`))
	rec := httptest.NewRecorder()

	handler.ItemsUpsert(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected %d, got %d", http.StatusAccepted, rec.Code)
	}
	if rec.Body.String() != `{"status":"accepted"}` {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
	if svc.itemsCalls != 1 {
		t.Fatalf("expected service to be called once, got %d", svc.itemsCalls)
	}
	if svc.itemsOrg != defaultOrg {
		t.Fatalf("expected org %s, got %s", defaultOrg, svc.itemsOrg)
	}
	if len(svc.itemsReq.Items) != 1 || svc.itemsReq.Items[0].ItemID != "sku-1" {
		t.Fatalf("service received %+v", svc.itemsReq.Items)
	}
}

func TestIngestionHandler_ItemsUpsert_ValidationError(t *testing.T) {
	t.Parallel()

	svc := &stubIngestionService{
		itemsErr: ingestion.ValidationError{
			Code:    "bad_input",
			Message: "oops",
		},
	}
	handler := NewIngestionHandler(svc, uuid.New(), nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/items:upsert", strings.NewReader(`{"namespace":"default","items":[{"item_id":"sku"}]}`))
	rec := httptest.NewRecorder()

	handler.ItemsUpsert(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
	var apiErr common.APIError
	if err := json.Unmarshal(rec.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if apiErr.Code != "bad_input" {
		t.Fatalf("expected error code bad_input, got %s", apiErr.Code)
	}
}

func TestIngestionHandler_EventsBatch_ServiceError(t *testing.T) {
	t.Parallel()

	svc := &stubIngestionService{
		eventsErr: errors.New("boom"),
	}
	handler := NewIngestionHandler(svc, uuid.New(), nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/events:batch", strings.NewReader(`{"namespace":"default","events":[{"user_id":"u1","item_id":"i1","type":0}]}`))
	rec := httptest.NewRecorder()

	handler.EventsBatch(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d (body=%s)", http.StatusInternalServerError, rec.Code, rec.Body.String())
	}
	if svc.eventsCalls != 1 {
		t.Fatalf("expected service to be called once, got %d", svc.eventsCalls)
	}
}
