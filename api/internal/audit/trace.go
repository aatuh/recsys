package audit

import (
	"time"

	"github.com/google/uuid"
)

// Trace captures a single recommendation decision for audit purposes.
type Trace struct {
	DecisionID  uuid.UUID           `json:"decision_id"`
	OrgID       string              `json:"org_id"`
	Timestamp   time.Time           `json:"ts"`
	Namespace   string              `json:"namespace"`
	Surface     string              `json:"surface,omitempty"`
	RequestID   string              `json:"request_id,omitempty"`
	UserHash    string              `json:"user_hash,omitempty"`
	K           int                 `json:"k,omitempty"`
	Constraints *TraceConstraints   `json:"constraints,omitempty"`
	Config      TraceConfig         `json:"effective_config"`
	Bandit      *TraceBandit        `json:"bandit,omitempty"`
	Candidates  []TraceCandidate    `json:"candidates_pre"`
	FinalItems  []TraceFinalItem    `json:"final_items"`
	MMR         []TraceMMR          `json:"mmr_info,omitempty"`
	Caps        map[string]TraceCap `json:"caps,omitempty"`
	Extras      map[string]any      `json:"extras,omitempty"`
}

// TraceConstraints mirrors request-level constraints applied during ranking.
type TraceConstraints struct {
	IncludeTagsAny []string  `json:"include_tags_any,omitempty"`
	ExcludeItemIDs []string  `json:"exclude_item_ids,omitempty"`
	PriceBetween   []float64 `json:"price_between,omitempty"`
	CreatedAfter   string    `json:"created_after,omitempty"`
}

// TraceConfig reflects the effective algorithm knobs used for the decision.
type TraceConfig struct {
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

// TraceBandit captures contextual bandit decisions when used.
type TraceBandit struct {
	ChosenPolicyID string            `json:"chosen_policy_id"`
	Algorithm      string            `json:"algorithm"`
	BucketKey      string            `json:"bucket_key,omitempty"`
	Explore        bool              `json:"explore"`
	RequestID      string            `json:"request_id,omitempty"`
	Explain        map[string]string `json:"explain,omitempty"`
}

// TraceCandidate captures score snapshots before MMR/caps enforcement.
type TraceCandidate struct {
	ItemID string  `json:"item_id"`
	Score  float64 `json:"score"`
}

// TraceFinalItem captures the delivered item list and reasons.
type TraceFinalItem struct {
	ItemID  string   `json:"item_id"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons,omitempty"`
}

// TraceMMR captures greedy selection metadata for MMR decisions.
type TraceMMR struct {
	PickIndex      int     `json:"pick_index"`
	ItemID         string  `json:"item_id"`
	MaxSimilarity  float64 `json:"max_sim,omitempty"`
	Relevance      float64 `json:"relevance,omitempty"`
	Penalty        float64 `json:"penalty,omitempty"`
	BrandCapHit    bool    `json:"brand_cap_hit,omitempty"`
	CategoryCapHit bool    `json:"category_cap_hit,omitempty"`
}

// TraceCap aggregates brand/category cap usage per item.
type TraceCap struct {
	Brand    *TraceCapUsage `json:"brand,omitempty"`
	Category *TraceCapUsage `json:"category,omitempty"`
}

// TraceCapUsage provides per-dimension cap metadata.
type TraceCapUsage struct {
	Applied bool   `json:"applied"`
	Limit   *int   `json:"limit,omitempty"`
	Count   *int   `json:"count,omitempty"`
	Value   string `json:"value,omitempty"`
}
