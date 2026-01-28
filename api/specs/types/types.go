package types

import (
	"time"
)

// Common response types

type Ack struct {
	Status string `json:"status"`
}

// Core domain types

type Item struct {
	ItemID          string    `json:"item_id" example:"i_123"`
	Available       bool      `json:"available" example:"true"`
	Price           *float64  `json:"price,omitempty" example:"19.90"`
	Tags            []string  `json:"tags,omitempty"`
	Props           any       `json:"props,omitempty"`
	Embedding       []float64 `json:"embedding,omitempty"`
	Brand           *string   `json:"brand,omitempty" example:"Acme"`
	Category        *string   `json:"category,omitempty" example:"Shoes"`
	CategoryPath    *[]string `json:"category_path,omitempty" example:"[\"Footwear\",\"Running\"]"`
	Description     *string   `json:"description,omitempty"`
	ImageURL        *string   `json:"image_url,omitempty" example:"https://cdn.example.com/p/sku.jpg"`
	MetadataVersion *string   `json:"metadata_version,omitempty" example:"2025-10-01"`
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

// Request/Response types for ingestion

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

// Recommendation types

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
	Pop        float64 `json:"pop" example:"0.3"`
	Cooc       float64 `json:"cooc" example:"0.7"`
	Similarity float64 `json:"similarity,omitempty" example:"0.0"`
}

type Overrides struct {
	PopularityHalfLifeDays    *int     `json:"popularity_halflife_days,omitempty"`
	CoVisWindowDays           *int     `json:"covis_window_days,omitempty"`
	PopularityFanout          *int     `json:"popularity_fanout,omitempty"`
	MMRLambda                 *float64 `json:"mmr_lambda,omitempty"`
	BrandCap                  *int     `json:"brand_cap,omitempty"`
	CategoryCap               *int     `json:"category_cap,omitempty"`
	RuleExcludeEvents         *bool    `json:"rule_exclude_events,omitempty"`
	PurchasedWindowDays       *int     `json:"purchased_window_days,omitempty"`
	ProfileWindowDays         *int     `json:"profile_window_days,omitempty"`
	ProfileBoost              *float64 `json:"profile_boost,omitempty"`
	ProfileTopN               *int     `json:"profile_top_n,omitempty"`
	ProfileStarterBlendWeight *float64 `json:"profile_starter_blend_weight,omitempty"`
	BlendAlpha                *float64 `json:"blend_alpha,omitempty"`
	BlendBeta                 *float64 `json:"blend_beta,omitempty"`
	BlendGamma                *float64 `json:"blend_gamma,omitempty"`
	BanditAlgo                *string  `json:"bandit_algo,omitempty"`
}

type RecommendRequest struct {
	UserID         string                `json:"user_id" example:"u_123"`
	Namespace      string                `json:"namespace" example:"default"`
	K              int                   `json:"k" example:"20"`
	Constraints    *RecommendConstraints `json:"constraints,omitempty"`
	Blend          *RecommendBlend       `json:"blend,omitempty"`
	Overrides      *Overrides            `json:"overrides,omitempty"`
	Context        map[string]any        `json:"context,omitempty"`
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
	SegmentID    string       `json:"segment_id,omitempty" example:"vip"`
	ProfileID    string       `json:"profile_id,omitempty" example:"vip-high-novelty"`
}

// Rerank types

type RerankCandidate struct {
	ItemID string   `json:"item_id" example:"sku-123"`
	Score  *float64 `json:"score,omitempty" example:"0.42"`
}

type RerankRequest struct {
	UserID         string                `json:"user_id" example:"u_123"`
	Namespace      string                `json:"namespace" example:"default"`
	K              int                   `json:"k" example:"20"`
	Items          []RerankCandidate     `json:"items"`
	Context        map[string]any        `json:"context,omitempty"`
	IncludeReasons bool                  `json:"include_reasons,omitempty" example:"true"`
	ExplainLevel   string                `json:"explain_level,omitempty" example:"numeric" enums:"tags,numeric,full"`
	Blend          *RecommendBlend       `json:"blend,omitempty"`
	Overrides      *Overrides            `json:"overrides,omitempty"`
	Constraints    *RecommendConstraints `json:"constraints,omitempty"`
}

type ExplainLLMRequest struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Namespace  string `json:"namespace"`
	Surface    string `json:"surface"`
	SegmentID  string `json:"segment_id,omitempty"`
	From       string `json:"from"`
	To         string `json:"to"`
	Question   string `json:"question,omitempty"`
}

type ExplainLLMResponse struct {
	Markdown string         `json:"markdown"`
	Facts    map[string]any `json:"facts"`
	Cache    string         `json:"cache"`
	Model    string         `json:"model"`
	Warnings []string       `json:"warnings,omitempty"`
}

// Explanation types

type ExplainBlendContribution struct {
	Pop        float64 `json:"pop,omitempty"`
	Cooc       float64 `json:"cooc,omitempty"`
	Similarity float64 `json:"similarity,omitempty"`
}

type ExplainBlendRaw struct {
	Pop        float64 `json:"pop,omitempty"`
	Cooc       float64 `json:"cooc,omitempty"`
	Similarity float64 `json:"similarity,omitempty"`
}

type ExplainBlend struct {
	Alpha          float64                  `json:"alpha,omitempty"`
	Beta           float64                  `json:"beta,omitempty"`
	Gamma          float64                  `json:"gamma,omitempty"`
	PopNorm        float64                  `json:"pop_norm,omitempty"`
	CoocNorm       float64                  `json:"cooc_norm,omitempty"`
	SimilarityNorm float64                  `json:"similarity_norm,omitempty"`
	Contributions  ExplainBlendContribution `json:"contrib"`
	Raw            *ExplainBlendRaw         `json:"raw,omitempty"`
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

// Event type configuration

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

// Audit types

type AuditDecisionListResponse struct {
	Decisions []AuditDecisionSummary `json:"decisions"`
}

type AuditDecisionsSearchRequest struct {
	Namespace string `json:"namespace"`
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
	UserHash  string `json:"user_hash,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type AuditDecisionSummary struct {
	DecisionID string                `json:"decision_id"`
	Namespace  string                `json:"namespace"`
	Ts         time.Time             `json:"ts"`
	Surface    string                `json:"surface,omitempty"`
	RequestID  string                `json:"request_id,omitempty"`
	UserHash   string                `json:"user_hash,omitempty"`
	K          *int                  `json:"k,omitempty"`
	FinalItems []AuditTraceFinalItem `json:"final_items"`
	Extras     map[string]any        `json:"extras,omitempty"`
}

type AuditDecisionDetail struct {
	DecisionID  string                   `json:"decision_id"`
	OrgID       string                   `json:"org_id"`
	Namespace   string                   `json:"namespace"`
	Ts          time.Time                `json:"ts"`
	Surface     string                   `json:"surface,omitempty"`
	RequestID   string                   `json:"request_id,omitempty"`
	UserHash    string                   `json:"user_hash,omitempty"`
	K           *int                     `json:"k,omitempty"`
	Constraints *AuditTraceConstraints   `json:"constraints,omitempty"`
	Config      AuditTraceConfig         `json:"effective_config"`
	Bandit      *AuditTraceBandit        `json:"bandit,omitempty"`
	Candidates  []AuditTraceCandidate    `json:"candidates_pre"`
	FinalItems  []AuditTraceFinalItem    `json:"final_items"`
	MMR         []AuditTraceMMR          `json:"mmr_info,omitempty"`
	Caps        map[string]AuditTraceCap `json:"caps,omitempty"`
	Extras      map[string]any           `json:"extras,omitempty"`
}

type AuditTraceConstraints struct {
	IncludeTagsAny []string  `json:"include_tags_any,omitempty"`
	ExcludeItemIDs []string  `json:"exclude_item_ids,omitempty"`
	PriceBetween   []float64 `json:"price_between,omitempty"`
	CreatedAfter   string    `json:"created_after,omitempty"`
}

type AuditTraceConfig struct {
	Alpha               float64 `json:"alpha"`
	Beta                float64 `json:"beta"`
	Gamma               float64 `json:"gamma"`
	ProfileBoost        float64 `json:"profile_boost"`
	ProfileWindowDays   float64 `json:"profile_window_days"`
	ProfileTopN         int     `json:"profile_top_n"`
	MMRLambda           float64 `json:"mmr_lambda"`
	BrandCap            int     `json:"brand_cap"`
	CategoryCap         int     `json:"category_cap"`
	HalfLifeDays        float64 `json:"half_life_days"`
	CoVisWindowDays     int     `json:"co_vis_window_days"`
	PurchasedWindowDays int     `json:"purchased_window_days"`
	RuleExcludeEvents   bool    `json:"rule_exclude_events"`
	PopularityFanout    int     `json:"popularity_fanout"`
}

type AuditTraceBandit struct {
	ChosenPolicyID string            `json:"chosen_policy_id"`
	Algorithm      string            `json:"algorithm"`
	BucketKey      string            `json:"bucket_key,omitempty"`
	Explore        bool              `json:"explore"`
	RequestID      string            `json:"request_id,omitempty"`
	Explain        map[string]string `json:"explain,omitempty"`
	Experiment     string            `json:"experiment,omitempty"`
	Variant        string            `json:"variant,omitempty"`
}

type AuditTraceCandidate struct {
	ItemID string  `json:"item_id"`
	Score  float64 `json:"score"`
}

type AuditTraceFinalItem struct {
	ItemID  string   `json:"item_id"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons,omitempty"`
}

type AuditTraceMMR struct {
	PickIndex      int     `json:"pick_index"`
	ItemID         string  `json:"item_id"`
	MaxSimilarity  float64 `json:"max_sim,omitempty"`
	Relevance      float64 `json:"relevance,omitempty"`
	Penalty        float64 `json:"penalty,omitempty"`
	BrandCapHit    bool    `json:"brand_cap_hit,omitempty"`
	CategoryCapHit bool    `json:"category_cap_hit,omitempty"`
}

type AuditTraceCap struct {
	Brand    *AuditTraceCapUsage `json:"brand,omitempty"`
	Category *AuditTraceCapUsage `json:"category,omitempty"`
}

type AuditTraceCapUsage struct {
	Applied bool   `json:"applied"`
	Limit   *int   `json:"limit,omitempty"`
	Count   *int   `json:"count,omitempty"`
	Value   string `json:"value,omitempty"`
}

type RecommendationConfigPayload struct {
	HalfLifeDays                  float64                                `json:"half_life_days"`
	CoVisWindowDays               float64                                `json:"co_vis_window_days"`
	PopularityFanout              int                                    `json:"popularity_fanout"`
	MaxK                          int                                    `json:"max_k"`
	MaxFanout                     int                                    `json:"max_fanout"`
	MaxExcludeIDs                 int                                    `json:"max_exclude_ids"`
	MaxAnchorsInjected            int                                    `json:"max_anchors_injected"`
	MMRLambda                     float64                                `json:"mmr_lambda"`
	BrandCap                      int                                    `json:"brand_cap"`
	CategoryCap                   int                                    `json:"category_cap"`
	RuleExcludeEvents             bool                                   `json:"rule_exclude_events"`
	ExcludeEventTypes             []int16                                `json:"exclude_event_types,omitempty"`
	BrandTagPrefixes              []string                               `json:"brand_tag_prefixes,omitempty"`
	CategoryTagPrefixes           []string                               `json:"category_tag_prefixes,omitempty"`
	PurchasedWindowDays           float64                                `json:"purchased_window_days"`
	ProfileWindowDays             float64                                `json:"profile_window_days"`
	ProfileBoost                  float64                                `json:"profile_boost"`
	ProfileTopNTags               int                                    `json:"profile_top_n_tags"`
	ProfileMinEventsForBoost      int                                    `json:"profile_min_events_for_boost"`
	ProfileColdStartMultiplier    float64                                `json:"profile_cold_start_multiplier"`
	ProfileStarterBlendWeight     float64                                `json:"profile_starter_blend_weight"`
	MMRPresets                    map[string]float64                     `json:"mmr_presets,omitempty"`
	BlendAlpha                    float64                                `json:"blend_alpha"`
	BlendBeta                     float64                                `json:"blend_beta"`
	BlendGamma                    float64                                `json:"blend_gamma"`
	NewUserBlendAlpha             *float64                               `json:"new_user_blend_alpha,omitempty"`
	NewUserBlendBeta              *float64                               `json:"new_user_blend_beta,omitempty"`
	NewUserBlendGamma             *float64                               `json:"new_user_blend_gamma,omitempty"`
	NewUserMMRLambda              *float64                               `json:"new_user_mmr_lambda,omitempty"`
	NewUserPopFanout              *int                                   `json:"new_user_pop_fanout,omitempty"`
	BanditExperiment              BanditExperimentConfig                 `json:"bandit_experiment"`
	RulesEnabled                  bool                                   `json:"rules_enabled"`
	CoverageCacheTTLSeconds       float64                                `json:"coverage_cache_ttl_seconds"`
	CoverageLongTailHintThreshold float64                                `json:"coverage_long_tail_hint_threshold"`
	SegmentProfiles               map[string]SegmentProfileConfigPayload `json:"segment_profiles,omitempty"`
}

type BanditExperimentConfig struct {
	Enabled        bool     `json:"enabled"`
	HoldoutPercent float64  `json:"holdout_percent"`
	Label          string   `json:"label,omitempty"`
	Surfaces       []string `json:"surfaces,omitempty"`
}

type SegmentProfileConfigPayload struct {
	BlendAlpha                float64  `json:"blend_alpha"`
	BlendBeta                 float64  `json:"blend_beta"`
	BlendGamma                float64  `json:"blend_gamma"`
	MMRLambda                 *float64 `json:"mmr_lambda,omitempty"`
	PopularityFanout          *int     `json:"popularity_fanout,omitempty"`
	ProfileStarterBlendWeight *float64 `json:"profile_starter_blend_weight,omitempty"`
}

type RecommendationConfigMetadata struct {
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by,omitempty"`
	Source    string    `json:"source,omitempty"`
	Notes     string    `json:"notes,omitempty"`
}

type RecommendationConfigDocument struct {
	Namespace string                       `json:"namespace"`
	Config    *RecommendationConfigPayload `json:"config"`
	Metadata  RecommendationConfigMetadata `json:"metadata"`
}

type RecommendationConfigUpdateRequest struct {
	Namespace string                       `json:"namespace"`
	Config    *RecommendationConfigPayload `json:"config"`
	Author    string                       `json:"author,omitempty"`
	Notes     string                       `json:"notes,omitempty"`
}
