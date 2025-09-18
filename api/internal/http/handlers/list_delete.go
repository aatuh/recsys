package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"recsys/internal/http/common"
	"recsys/specs/types"

	"github.com/google/uuid"
)

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
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		common.BadRequest(w, r, "missing_namespace", "namespace parameter is required", nil)
		return
	}

	limit := 100 // default
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	offset := 0 // default
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Build filters
	filters := make(map[string]interface{})
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if createdAfter := r.URL.Query().Get("created_after"); createdAfter != "" {
		filters["created_after"] = createdAfter
	}
	if createdBefore := r.URL.Query().Get("created_before"); createdBefore != "" {
		filters["created_before"] = createdBefore
	}

	orgID := h.DefaultOrg
	if s := r.Header.Get("X-Org-ID"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			orgID = id
		}
	}

	users, total, err := h.Store.ListUsers(r.Context(), orgID, namespace, limit, offset, filters)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	hasMore := offset+limit < total
	var nextOffset *int
	if hasMore {
		next := offset + limit
		nextOffset = &next
	}

	response := types.ListResponse{
		Items:      convertToAnySlice(users),
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		common.BadRequest(w, r, "missing_namespace", "namespace parameter is required", nil)
		return
	}

	limit := 100 // default
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	offset := 0 // default
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Build filters
	filters := make(map[string]interface{})
	if itemID := r.URL.Query().Get("item_id"); itemID != "" {
		filters["item_id"] = itemID
	}
	if createdAfter := r.URL.Query().Get("created_after"); createdAfter != "" {
		filters["created_after"] = createdAfter
	}
	if createdBefore := r.URL.Query().Get("created_before"); createdBefore != "" {
		filters["created_before"] = createdBefore
	}

	orgID := h.DefaultOrg
	if s := r.Header.Get("X-Org-ID"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			orgID = id
		}
	}

	items, total, err := h.Store.ListItems(r.Context(), orgID, namespace, limit, offset, filters)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	hasMore := offset+limit < total
	var nextOffset *int
	if hasMore {
		next := offset + limit
		nextOffset = &next
	}

	response := types.ListResponse{
		Items:      convertToAnySlice(items),
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		common.BadRequest(w, r, "missing_namespace", "namespace parameter is required", nil)
		return
	}

	limit := 100 // default
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	offset := 0 // default
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Build filters
	filters := make(map[string]interface{})
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
	if createdAfter := r.URL.Query().Get("created_after"); createdAfter != "" {
		filters["created_after"] = createdAfter
	}
	if createdBefore := r.URL.Query().Get("created_before"); createdBefore != "" {
		filters["created_before"] = createdBefore
	}

	orgID := h.DefaultOrg
	if s := r.Header.Get("X-Org-ID"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			orgID = id
		}
	}

	events, total, err := h.Store.ListEvents(r.Context(), orgID, namespace, limit, offset, filters)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	hasMore := offset+limit < total
	var nextOffset *int
	if hasMore {
		next := offset + limit
		nextOffset = &next
	}

	response := types.ListResponse{
		Items:      convertToAnySlice(events),
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
func (h *Handler) DeleteUsers(w http.ResponseWriter, r *http.Request) {
	var req types.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	orgID := h.DefaultOrg
	if s := r.Header.Get("X-Org-ID"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			orgID = id
		}
	}

	// Build filters
	filters := make(map[string]interface{})
	if req.UserID != nil {
		filters["user_id"] = *req.UserID
	}
	if req.CreatedAfter != nil {
		filters["created_after"] = *req.CreatedAfter
	}
	if req.CreatedBefore != nil {
		filters["created_before"] = *req.CreatedBefore
	}

	deletedCount, err := h.Store.DeleteUsers(r.Context(), orgID, req.Namespace, filters)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	response := types.DeleteResponse{
		DeletedCount: deletedCount,
		Message:      "Users deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
func (h *Handler) DeleteItems(w http.ResponseWriter, r *http.Request) {
	var req types.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	orgID := h.DefaultOrg
	if s := r.Header.Get("X-Org-ID"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			orgID = id
		}
	}

	// Build filters
	filters := make(map[string]interface{})
	if req.ItemID != nil {
		filters["item_id"] = *req.ItemID
	}
	if req.CreatedAfter != nil {
		filters["created_after"] = *req.CreatedAfter
	}
	if req.CreatedBefore != nil {
		filters["created_before"] = *req.CreatedBefore
	}

	deletedCount, err := h.Store.DeleteItems(r.Context(), orgID, req.Namespace, filters)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	response := types.DeleteResponse{
		DeletedCount: deletedCount,
		Message:      "Items deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
func (h *Handler) DeleteEvents(w http.ResponseWriter, r *http.Request) {
	var req types.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	orgID := h.DefaultOrg
	if s := r.Header.Get("X-Org-ID"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			orgID = id
		}
	}

	// Build filters
	filters := make(map[string]interface{})
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

	deletedCount, err := h.Store.DeleteEvents(r.Context(), orgID, req.Namespace, filters)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	response := types.DeleteResponse{
		DeletedCount: deletedCount,
		Message:      "Events deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper function to convert []map[string]interface{} to []any
func convertToAnySlice(items []map[string]interface{}) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}
