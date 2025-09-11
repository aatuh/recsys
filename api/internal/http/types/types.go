package types

type Ack struct {
	Status string `json:"status"`
}

type Item struct {
	ItemID    string    `json:"item_id" example:"i_123"`
	Available bool      `json:"available" example:"true"`
	Price     *float64  `json:"price,omitempty" example:"19.90"`
	Tags      []string  `json:"tags,omitempty"`
	Props     any       `json:"props,omitempty"`
	Embedding []float64 `json:"embedding,omitempty"`
}

type User struct {
	UserID string `json:"user_id" example:"u_123"`
	Traits any    `json:"traits,omitempty"`
}

// Optional idempotency key from the client. If set, duplicates are ignored.
// Must be unique per (org_id, namespace, source_event_id).
type Event struct {
	UserID        string  `json:"user_id" example:"u_123"`
	ItemID        string  `json:"item_id" example:"i_123"`
	Type          int16   `json:"type" example:"0"` // 0=view,1=click,2=add,3=purchase,4=custom
	Value         float64 `json:"value" example:"1"`
	TS            string  `json:"ts,omitempty" example:"2025-09-07T12:34:56Z"`
	Meta          any     `json:"meta,omitempty"`
	SourceEventID *string `json:"source_event_id,omitempty"`
}

type ItemsUpsertRequest struct {
	Namespace string `json:"namespace" example:"default"`
	Items     []Item `json:"items"`
}

type UsersUpsertRequest struct {
	Namespace string `json:"namespace" example:"default"`
	Users     []User `json:"users"`
}

type EventsBatchRequest struct {
	Namespace string  `json:"namespace" example:"default"`
	Events    []Event `json:"events"`
}

type RecommendConstraints struct {
	// Match if item.tags overlaps these (any). Empty/omitted = no tag filter.
	IncludeTagsAny []string `json:"include_tags_any,omitempty"`
	// Exclude these item IDs from results.
	ExcludeItemIDs []string `json:"exclude_item_ids,omitempty"`
	// Optional price bounds: [min, max]. Either end may be omitted.
	PriceBetween []float64 `json:"price_between,omitempty"`
	// Optional ISO8601 timestamp; only consider events on/after this instant.
	CreatedAfterISO string `json:"created_after,omitempty" example:"2025-06-01T00:00:00Z"`
}

type RecommendBlend struct {
	Pop  float64 `json:"pop" example:"0.3"`
	Cooc float64 `json:"cooc" example:"0.7"`
	ALS  float64 `json:"als,omitempty" example:"0.0"`
}

type RecommendRequest struct {
	UserID         string                `json:"user_id" example:"u_123"`
	Namespace      string                `json:"namespace" example:"default"`
	K              int                   `json:"k" example:"20"`
	Constraints    *RecommendConstraints `json:"constraints,omitempty"`
	Blend          *RecommendBlend       `json:"blend,omitempty"`
	IncludeReasons bool                  `json:"include_reasons,omitempty" example:"true"`
}

type ScoredItem struct {
	ItemID  string   `json:"item_id" example:"i_101"`
	Score   float64  `json:"score" example:"0.87"`
	Reasons []string `json:"reasons,omitempty"`
}

type RecommendResponse struct {
	ModelVersion string       `json:"model_version" example:"pop_2025-09-07_01"`
	Items        []ScoredItem `json:"items"`
}

type EventTypeConfig struct {
	Type         int16    `json:"type"`
	Name         *string  `json:"name,omitempty"`
	Weight       float64  `json:"weight"`
	HalfLifeDays *float64 `json:"half_life_days,omitempty"`
	IsActive     *bool    `json:"is_active,omitempty"`
}

type EventTypeConfigUpsertRequest struct {
	Namespace string            `json:"namespace"`
	Types     []EventTypeConfig `json:"types"`
}

type EventTypeConfigUpsertResponse struct {
	Type         int16    `json:"type"`
	Name         *string  `json:"name,omitempty"`
	Weight       float64  `json:"weight"`
	HalfLifeDays *float64 `json:"half_life_days,omitempty"`
	IsActive     bool     `json:"is_active"`
	Source       string   `json:"source"` // "tenant" or "default"
}

// List and Delete types

type ListRequest struct {
	Namespace string `json:"namespace" example:"default"`
	Limit     int    `json:"limit,omitempty" example:"100"`
	Offset    int    `json:"offset,omitempty" example:"0"`
	// Optional filters
	UserID    *string `json:"user_id,omitempty" example:"u_123"`
	ItemID    *string `json:"item_id,omitempty" example:"i_123"`
	EventType *int16  `json:"event_type,omitempty" example:"0"`
	// Date range filters
	CreatedAfter  *string `json:"created_after,omitempty" example:"2025-01-01T00:00:00Z"`
	CreatedBefore *string `json:"created_before,omitempty" example:"2025-12-31T23:59:59Z"`
}

type DeleteRequest struct {
	Namespace string `json:"namespace" example:"default"`
	// Optional filters - if not provided, deletes all data in namespace
	UserID    *string `json:"user_id,omitempty" example:"u_123"`
	ItemID    *string `json:"item_id,omitempty" example:"i_123"`
	EventType *int16  `json:"event_type,omitempty" example:"0"`
	// Date range filters
	CreatedAfter  *string `json:"created_after,omitempty" example:"2025-01-01T00:00:00Z"`
	CreatedBefore *string `json:"created_before,omitempty" example:"2025-12-31T23:59:59Z"`
}

type ListResponse struct {
	Items      []any `json:"items"`
	Total      int   `json:"total"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
	HasMore    bool  `json:"has_more"`
	NextOffset *int  `json:"next_offset,omitempty"`
}

type DeleteResponse struct {
	DeletedCount int    `json:"deleted_count"`
	Message      string `json:"message"`
}
