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

type Overrides struct {
	PopularityHalfLifeDays *int     `json:"popularity_halflife_days,omitempty"`
	CoVisWindowDays        *int     `json:"covis_window_days,omitempty"`
	PopularityFanout       *int     `json:"popularity_fanout,omitempty"`
	MMRLambda              *float64 `json:"mmr_lambda,omitempty"`
	BrandCap               *int     `json:"brand_cap,omitempty"`
	CategoryCap            *int     `json:"category_cap,omitempty"`
	RuleExcludePurchased   *bool    `json:"rule_exclude_purchased,omitempty"`
	PurchasedWindowDays    *int     `json:"purchased_window_days,omitempty"`
	ProfileWindowDays      *int     `json:"profile_window_days,omitempty"`
	ProfileBoost           *float64 `json:"profile_boost,omitempty"`
	ProfileTopN            *int     `json:"profile_top_n,omitempty"`
	BlendAlpha             *float64 `json:"blend_alpha,omitempty"`
	BlendBeta              *float64 `json:"blend_beta,omitempty"`
	BlendGamma             *float64 `json:"blend_gamma,omitempty"`
	BanditAlgo             *string  `json:"bandit_algo,omitempty"`
}

type RecommendRequest struct {
	UserID         string                `json:"user_id" example:"u_123"`
	Namespace      string                `json:"namespace" example:"default"`
	K              int                   `json:"k" example:"20"`
	Constraints    *RecommendConstraints `json:"constraints,omitempty"`
	Blend          *RecommendBlend       `json:"blend,omitempty"`
	Overrides      *Overrides            `json:"overrides,omitempty"`
	IncludeReasons bool                  `json:"include_reasons,omitempty" example:"true"`
	ExplainLevel   string                `json:"explain_level,omitempty" example:"numeric" enums:"tags,numeric,full"`
}

type ScoredItem struct {
	ItemID  string        `json:"item_id" example:"i_101"`
	Score   float64       `json:"score" example:"0.87"`
	Reasons []string      `json:"reasons,omitempty"`
	Explain *ExplainBlock `json:"explain,omitempty"`
}

type RecommendResponse struct {
	ModelVersion string       `json:"model_version" example:"pop_2025-09-07_01"`
	Items        []ScoredItem `json:"items"`
}

type ExplainBlendContribution struct {
	Pop  float64 `json:"pop,omitempty"`
	Cooc float64 `json:"cooc,omitempty"`
	Emb  float64 `json:"emb,omitempty"`
}

type ExplainBlendRaw struct {
	Pop  float64 `json:"pop,omitempty"`
	Cooc float64 `json:"cooc,omitempty"`
	Emb  float64 `json:"emb,omitempty"`
}

type ExplainBlend struct {
	Alpha         float64                  `json:"alpha,omitempty"`
	Beta          float64                  `json:"beta,omitempty"`
	Gamma         float64                  `json:"gamma,omitempty"`
	PopNorm       float64                  `json:"pop_norm,omitempty"`
	CoocNorm      float64                  `json:"cooc_norm,omitempty"`
	EmbNorm       float64                  `json:"emb_norm,omitempty"`
	Contributions ExplainBlendContribution `json:"contrib"`
	Raw           *ExplainBlendRaw         `json:"raw,omitempty"`
}

type ExplainPersonalizationRaw struct {
	ProfileBoost float64 `json:"profile_boost,omitempty"`
}

type ExplainPersonalization struct {
	Overlap         float64                    `json:"overlap,omitempty"`
	BoostMultiplier float64                    `json:"boost_multiplier,omitempty"`
	Raw             *ExplainPersonalizationRaw `json:"raw,omitempty"`
}

type ExplainMMR struct {
	Lambda        float64 `json:"lambda,omitempty"`
	MaxSimilarity float64 `json:"max_sim,omitempty"`
	Penalty       float64 `json:"penalty,omitempty"`
	Relevance     float64 `json:"relevance,omitempty"`
	Rank          int     `json:"rank,omitempty"`
}

type ExplainCapUsage struct {
	Applied bool   `json:"applied"`
	Limit   *int   `json:"limit,omitempty"`
	Count   *int   `json:"count,omitempty"`
	Value   string `json:"value,omitempty"`
}

type ExplainCaps struct {
	Brand    *ExplainCapUsage `json:"brand,omitempty"`
	Category *ExplainCapUsage `json:"category,omitempty"`
}

type ExplainBlock struct {
	Blend           *ExplainBlend           `json:"blend,omitempty"`
	Personalization *ExplainPersonalization `json:"personalization,omitempty"`
	MMR             *ExplainMMR             `json:"mmr,omitempty"`
	Caps            *ExplainCaps            `json:"caps,omitempty"`
	Anchors         []string                `json:"anchors,omitempty"`
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
