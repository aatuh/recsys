package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"recsys/internal/services/datamanagement"
	specstypes "recsys/specs/types"
)

type stubDataService struct {
	listUsersFn  func(context.Context, uuid.UUID, datamanagement.ListOptions) (specstypes.ListResponse, error)
	listItemsFn  func(context.Context, uuid.UUID, datamanagement.ListOptions) (specstypes.ListResponse, error)
	listEventsFn func(context.Context, uuid.UUID, datamanagement.ListOptions) (specstypes.ListResponse, error)
	delUsersFn   func(context.Context, uuid.UUID, datamanagement.DeleteOptions) (int, error)
	delItemsFn   func(context.Context, uuid.UUID, datamanagement.DeleteOptions) (int, error)
	delEventsFn  func(context.Context, uuid.UUID, datamanagement.DeleteOptions) (int, error)

	lastOrg  uuid.UUID
	lastOpts datamanagement.ListOptions
	lastDel  datamanagement.DeleteOptions
}

func (s *stubDataService) ListUsers(ctx context.Context, org uuid.UUID, opts datamanagement.ListOptions) (specstypes.ListResponse, error) {
	s.lastOrg = org
	s.lastOpts = opts
	if s.listUsersFn != nil {
		return s.listUsersFn(ctx, org, opts)
	}
	return specstypes.ListResponse{}, nil
}

func (s *stubDataService) ListItems(ctx context.Context, org uuid.UUID, opts datamanagement.ListOptions) (specstypes.ListResponse, error) {
	s.lastOrg = org
	s.lastOpts = opts
	if s.listItemsFn != nil {
		return s.listItemsFn(ctx, org, opts)
	}
	return specstypes.ListResponse{}, nil
}

func (s *stubDataService) ListEvents(ctx context.Context, org uuid.UUID, opts datamanagement.ListOptions) (specstypes.ListResponse, error) {
	s.lastOrg = org
	s.lastOpts = opts
	if s.listEventsFn != nil {
		return s.listEventsFn(ctx, org, opts)
	}
	return specstypes.ListResponse{}, nil
}

func (s *stubDataService) DeleteUsers(ctx context.Context, org uuid.UUID, opts datamanagement.DeleteOptions) (int, error) {
	s.lastOrg = org
	s.lastDel = opts
	if s.delUsersFn != nil {
		return s.delUsersFn(ctx, org, opts)
	}
	return 0, nil
}

func (s *stubDataService) DeleteItems(ctx context.Context, org uuid.UUID, opts datamanagement.DeleteOptions) (int, error) {
	s.lastOrg = org
	s.lastDel = opts
	if s.delItemsFn != nil {
		return s.delItemsFn(ctx, org, opts)
	}
	return 0, nil
}

func (s *stubDataService) DeleteEvents(ctx context.Context, org uuid.UUID, opts datamanagement.DeleteOptions) (int, error) {
	s.lastOrg = org
	s.lastDel = opts
	if s.delEventsFn != nil {
		return s.delEventsFn(ctx, org, opts)
	}
	return 0, nil
}

func TestDataManagementHandler_ListUsers_Success(t *testing.T) {
	t.Parallel()

	defaultOrg := uuid.New()
	svc := &stubDataService{
		listUsersFn: func(_ context.Context, _ uuid.UUID, opts datamanagement.ListOptions) (specstypes.ListResponse, error) {
			return specstypes.ListResponse{
				Items:   []any{map[string]any{"user_id": "u1"}},
				Total:   1,
				Limit:   opts.Limit,
				Offset:  opts.Offset,
				HasMore: false,
			}, nil
		},
	}
	handler := NewDataManagementHandler(svc, defaultOrg, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users?namespace=default&limit=50&user_id=u1", nil)
	req.Header.Set("X-Org-ID", uuid.New().String())
	rec := httptest.NewRecorder()

	handler.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}

	var resp specstypes.ListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if resp.Total != 1 || len(resp.Items) != 1 {
		t.Fatalf("unexpected payload: %+v", resp)
	}

	if svc.lastOpts.Namespace != "default" || svc.lastOpts.Filters["user_id"] != "u1" {
		t.Fatalf("service received unexpected opts: %+v", svc.lastOpts)
	}
}

func TestDataManagementHandler_ListUsers_ValidationError(t *testing.T) {
	t.Parallel()

	svc := &stubDataService{
		listUsersFn: func(_ context.Context, _ uuid.UUID, _ datamanagement.ListOptions) (specstypes.ListResponse, error) {
			return specstypes.ListResponse{}, datamanagement.ValidationError{
				Code:    "missing_namespace",
				Message: "namespace parameter is required",
			}
		},
	}
	handler := NewDataManagementHandler(svc, uuid.New(), nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.ListUsers(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
	var apiErr struct {
		Code string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if apiErr.Code != "missing_namespace" {
		t.Fatalf("expected code missing_namespace, got %s", apiErr.Code)
	}
}

func TestDataManagementHandler_DeleteItems_ServiceError(t *testing.T) {
	t.Parallel()

	svc := &stubDataService{
		delItemsFn: func(context.Context, uuid.UUID, datamanagement.DeleteOptions) (int, error) {
			return 0, errors.New("boom")
		},
	}
	handler := NewDataManagementHandler(svc, uuid.New(), nil)

	payload := specstypes.DeleteRequest{Namespace: "default"}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/items:delete", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.DeleteItems(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
