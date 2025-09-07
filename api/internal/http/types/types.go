package types

type Ack struct {
	Status string `json:"status"`
}

type Item struct {
	ItemID    string   `json:"item_id" example:"i_123"`
	Available bool     `json:"available" example:"true"`
	Price     *float64 `json:"price,omitempty" example:"19.90"`
	Tags      []string `json:"tags,omitempty"`
	Props     any      `json:"props,omitempty"`
}

type User struct {
	UserID string `json:"user_id" example:"u_123"`
	Traits any    `json:"traits,omitempty"`
}

type Event struct {
	UserID string  `json:"user_id" example:"u_123"`
	ItemID string  `json:"item_id" example:"i_123"`
	Type   int16   `json:"type" example:"0"` // 0=view,1=click,2=add,3=purchase,4=custom
	Value  float64 `json:"value" example:"1"`
	TS     string  `json:"ts" example:"2025-09-07T12:34:56Z"`
	Meta   any     `json:"meta,omitempty"`
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
	IncludeTagsAny  []string  `json:"include_tags_any,omitempty"`
	ExcludeItemIDs  []string  `json:"exclude_item_ids,omitempty"`
	PriceBetween    []float64 `json:"price_between,omitempty"`
	CreatedAfterISO string    `json:"created_after,omitempty" example:"2025-06-01T00:00:00Z"`
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
