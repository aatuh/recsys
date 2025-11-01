package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"recsys/internal/http/common"
	"recsys/internal/services/datamanagement"
	specstypes "recsys/specs/types"
)

// DataManagementService defines the domain contract required by the HTTP adapter.
type DataManagementService interface {
	ListUsers(ctx context.Context, orgID uuid.UUID, opts datamanagement.ListOptions) (specstypes.ListResponse, error)
	ListItems(ctx context.Context, orgID uuid.UUID, opts datamanagement.ListOptions) (specstypes.ListResponse, error)
	ListEvents(ctx context.Context, orgID uuid.UUID, opts datamanagement.ListOptions) (specstypes.ListResponse, error)
	DeleteUsers(ctx context.Context, orgID uuid.UUID, opts datamanagement.DeleteOptions) (int, error)
	DeleteItems(ctx context.Context, orgID uuid.UUID, opts datamanagement.DeleteOptions) (int, error)
	DeleteEvents(ctx context.Context, orgID uuid.UUID, opts datamanagement.DeleteOptions) (int, error)
}

// DataManagementHandler exposes list/delete routes backed by the service.
type DataManagementHandler struct {
	service    DataManagementService
	defaultOrg uuid.UUID
	logger     *zap.Logger
}

// NewDataManagementHandler constructs a handler for data-management endpoints.
func NewDataManagementHandler(svc DataManagementService, defaultOrg uuid.UUID, logger *zap.Logger) *DataManagementHandler {
	h := &DataManagementHandler{
		service:    svc,
		defaultOrg: defaultOrg,
		logger:     logger,
	}
	if h.logger == nil {
		h.logger = zap.NewNop()
	}
	return h
}

// ListUsers godoc
// @Summary      List users with pagination and filtering
// @Description  Get a paginated list of users with optional filtering by user_id, date range, etc.
// @Tags         data-management
// @Accept       json
// @Produce      json
// @Param        namespace  query  string  true   "Namespace"
// @Param        limit      query  int     false  "Limit (default: 100, max: 1000)"
// @Param        offset     query  int     false  "Offset (default: 0)"
// @Param        user_id    query  string  false  "Filter by user ID"
// @Param        created_after   query  string  false  "Filter by creation date (ISO8601)"
// @Param        created_before  query  string  false  "Filter by creation date (ISO8601)"
// @Success      200        {object}  types.ListResponse
// @Failure      400        {object}  common.APIError
// @Router       /v1/users [get]
// @ID listUsers
func (h *DataManagementHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	filters := make(map[string]any)
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	addCreatedFilters(r, filters)

	h.list(w, r, func(ctx context.Context, orgID uuid.UUID, opts datamanagement.ListOptions) (specstypes.ListResponse, error) {
		opts.Filters = filters
		return h.service.ListUsers(ctx, orgID, opts)
	})
}

// ListItems godoc
// @Summary      List items with pagination and filtering
// @Description  Get a paginated list of items with optional filtering by item_id, date range, etc.
// @Tags         data-management
// @Accept       json
// @Produce      json
// @Param        namespace  query  string  true   "Namespace"
// @Param        limit      query  int     false  "Limit (default: 100, max: 1000)"
// @Param        offset     query  int     false  "Offset (default: 0)"
// @Param        item_id    query  string  false  "Filter by item ID"
// @Param        created_after   query  string  false  "Filter by creation date (ISO8601)"
// @Param        created_before  query  string  false  "Filter by creation date (ISO8601)"
// @Success      200        {object}  types.ListResponse
// @Failure      400        {object}  common.APIError
// @Router       /v1/items [get]
// @ID listItems
func (h *DataManagementHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	filters := make(map[string]any)
	if itemID := r.URL.Query().Get("item_id"); itemID != "" {
		filters["item_id"] = itemID
	}
	addCreatedFilters(r, filters)

	h.list(w, r, func(ctx context.Context, orgID uuid.UUID, opts datamanagement.ListOptions) (specstypes.ListResponse, error) {
		opts.Filters = filters
		return h.service.ListItems(ctx, orgID, opts)
	})
}

// ListEvents godoc
// @Summary      List events with pagination and filtering
// @Description  Get a paginated list of events with optional filtering by user_id, item_id, event_type, date range, etc.
// @Tags         data-management
// @Accept       json
// @Produce      json
// @Param        namespace  query  string  true   "Namespace"
// @Param        limit      query  int     false  "Limit (default: 100, max: 1000)"
// @Param        offset     query  int     false  "Offset (default: 0)"
// @Param        user_id    query  string  false  "Filter by user ID"
// @Param        item_id    query  string  false  "Filter by item ID"
// @Param        event_type query  int     false  "Filter by event type"
// @Param        created_after   query  string  false  "Filter by creation date (ISO8601)"
// @Param        created_before  query  string  false  "Filter by creation date (ISO8601)"
// @Success      200        {object}  types.ListResponse
// @Failure      400        {object}  common.APIError
// @Router       /v1/events [get]
// @ID listEvents
func (h *DataManagementHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	filters := make(map[string]any)
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if itemID := r.URL.Query().Get("item_id"); itemID != "" {
		filters["item_id"] = itemID
	}
	if eventType := r.URL.Query().Get("event_type"); eventType != "" {
		if parsed, err := strconv.ParseInt(eventType, 10, 16); err == nil {
			filters["event_type"] = int16(parsed)
		}
	}
	addCreatedFilters(r, filters)

	h.list(w, r, func(ctx context.Context, orgID uuid.UUID, opts datamanagement.ListOptions) (specstypes.ListResponse, error) {
		opts.Filters = filters
		return h.service.ListEvents(ctx, orgID, opts)
	})
}

// DeleteUsers godoc
// @Summary      Delete users with optional filtering
// @Description  Delete users based on filters. If no filters provided, deletes all users in namespace.
// @Tags         data-management
// @Accept       json
// @Produce      json
// @Param        payload  body  types.DeleteRequest  true  "Delete request"
// @Success      200      {object}  types.DeleteResponse
// @Failure      400      {object}  common.APIError
// @Router       /v1/users:delete [post]
// @ID deleteUsers
func (h *DataManagementHandler) DeleteUsers(w http.ResponseWriter, r *http.Request) {
	h.delete(w, r, "Users deleted successfully", func(ctx context.Context, orgID uuid.UUID, opts datamanagement.DeleteOptions) (int, error) {
		return h.service.DeleteUsers(ctx, orgID, opts)
	})
}

// DeleteItems godoc
// @Summary      Delete items with optional filtering
// @Description  Delete items based on filters. If no filters provided, deletes all items in namespace.
// @Tags         data-management
// @Accept       json
// @Produce      json
// @Param        payload  body  types.DeleteRequest  true  "Delete request"
// @Success      200      {object}  types.DeleteResponse
// @Failure      400      {object}  common.APIError
// @Router       /v1/items:delete [post]
// @ID deleteItems
func (h *DataManagementHandler) DeleteItems(w http.ResponseWriter, r *http.Request) {
	h.delete(w, r, "Items deleted successfully", func(ctx context.Context, orgID uuid.UUID, opts datamanagement.DeleteOptions) (int, error) {
		return h.service.DeleteItems(ctx, orgID, opts)
	})
}

// DeleteEvents godoc
// @Summary      Delete events with optional filtering
// @Description  Delete events based on filters. If no filters provided, deletes all events in namespace.
// @Tags         data-management
// @Accept       json
// @Produce      json
// @Param        payload  body  types.DeleteRequest  true  "Delete request"
// @Success      200      {object}  types.DeleteResponse
// @Failure      400      {object}  common.APIError
// @Router       /v1/events:delete [post]
// @ID deleteEvents
func (h *DataManagementHandler) DeleteEvents(w http.ResponseWriter, r *http.Request) {
	h.delete(w, r, "Events deleted successfully", func(ctx context.Context, orgID uuid.UUID, opts datamanagement.DeleteOptions) (int, error) {
		return h.service.DeleteEvents(ctx, orgID, opts)
	})
}

type listInvoker func(ctx context.Context, orgID uuid.UUID, opts datamanagement.ListOptions) (specstypes.ListResponse, error)

func (h *DataManagementHandler) list(w http.ResponseWriter, r *http.Request, invoke listInvoker) {
	orgID := orgIDFromHeader(r, h.defaultOrg)
	limit := parseLimitParam(r.URL.Query().Get("limit"), 100, 1000)
	offset := parseOffsetParam(r.URL.Query().Get("offset"), 0)

	opts := datamanagement.ListOptions{
		Namespace: r.URL.Query().Get("namespace"),
		Limit:     limit,
		Offset:    offset,
		Filters:   make(map[string]any),
	}

	resp, err := invoke(r.Context(), orgID, opts)
	if err != nil {
		var vErr datamanagement.ValidationError
		if errors.As(err, &vErr) {
			common.BadRequest(w, r, vErr.Code, vErr.Message, vErr.Details)
			return
		}
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

type deleteInvoker func(ctx context.Context, orgID uuid.UUID, opts datamanagement.DeleteOptions) (int, error)

func (h *DataManagementHandler) delete(w http.ResponseWriter, r *http.Request, successMsg string, invoke deleteInvoker) {
	var req specstypes.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg)
	opts := datamanagement.DeleteOptions{
		Namespace: req.Namespace,
		Filters:   buildDeleteFilters(req),
	}

	count, err := invoke(r.Context(), orgID, opts)
	if err != nil {
		var vErr datamanagement.ValidationError
		if errors.As(err, &vErr) {
			common.BadRequest(w, r, vErr.Code, vErr.Message, vErr.Details)
			return
		}
		common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, h.logger)
		return
	}

	response := specstypes.DeleteResponse{
		DeletedCount: count,
		Message:      successMsg,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func addCreatedFilters(r *http.Request, filters map[string]any) {
	if createdAfter := r.URL.Query().Get("created_after"); createdAfter != "" {
		filters["created_after"] = createdAfter
	}
	if createdBefore := r.URL.Query().Get("created_before"); createdBefore != "" {
		filters["created_before"] = createdBefore
	}
}

func parseLimitParam(val string, def, max int) int {
	if val == "" {
		return def
	}
	if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 && (max <= 0 || parsed <= max) {
		return parsed
	}
	return def
}

func parseOffsetParam(val string, def int) int {
	if val == "" {
		return def
	}
	if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
		return parsed
	}
	return def
}

func buildDeleteFilters(req specstypes.DeleteRequest) map[string]any {
	filters := make(map[string]any)
	if req.UserID != nil {
		filters["user_id"] = *req.UserID
	}
	if req.ItemID != nil {
		filters["item_id"] = *req.ItemID
	}
	if req.EventType != nil {
		filters["event_type"] = *req.EventType
	}
	if req.CreatedAfter != nil {
		filters["created_after"] = *req.CreatedAfter
	}
	if req.CreatedBefore != nil {
		filters["created_before"] = *req.CreatedBefore
	}
	return filters
}
