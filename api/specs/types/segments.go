package types

import (
	"encoding/json"
	"time"
)

type SegmentProfile struct {
	ProfileID           string     `json:"profile_id"`
	Description         string     `json:"description,omitempty"`
	BlendAlpha          float64    `json:"blend_alpha"`
	BlendBeta           float64    `json:"blend_beta"`
	BlendGamma          float64    `json:"blend_gamma"`
	MMRLambda           float64    `json:"mmr_lambda"`
	BrandCap            int        `json:"brand_cap"`
	CategoryCap         int        `json:"category_cap"`
	ProfileBoost        float64    `json:"profile_boost"`
	ProfileWindowDays   float64    `json:"profile_window_days"`
	ProfileTopN         int        `json:"profile_top_n"`
	HalfLifeDays        float64    `json:"half_life_days"`
	CoVisWindowDays     int        `json:"co_vis_window_days"`
	PurchasedWindowDays int        `json:"purchased_window_days"`
	RuleExcludeEvents   bool       `json:"rule_exclude_events"`
	ExcludeEventTypes   []int16    `json:"exclude_event_types,omitempty"`
	BrandTagPrefixes    []string   `json:"brand_tag_prefixes,omitempty"`
	CategoryTagPrefixes []string   `json:"category_tag_prefixes,omitempty"`
	PopularityFanout    int        `json:"popularity_fanout"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}

type SegmentProfilesListResponse struct {
	Namespace string           `json:"namespace"`
	Profiles  []SegmentProfile `json:"profiles"`
}

type SegmentProfilesUpsertRequest struct {
	Namespace string           `json:"namespace"`
	Profiles  []SegmentProfile `json:"profiles"`
}

type SegmentRule struct {
	RuleID      *int64          `json:"rule_id,omitempty"`
	Rule        json.RawMessage `json:"rule"`
	Enabled     bool            `json:"enabled"`
	Description string          `json:"description,omitempty"`
}

type Segment struct {
	SegmentID   string        `json:"segment_id"`
	Name        string        `json:"name"`
	Priority    int           `json:"priority"`
	Active      bool          `json:"active"`
	ProfileID   string        `json:"profile_id"`
	Description string        `json:"description,omitempty"`
	Rules       []SegmentRule `json:"rules,omitempty"`
	CreatedAt   *time.Time    `json:"created_at,omitempty"`
	UpdatedAt   *time.Time    `json:"updated_at,omitempty"`
}

type SegmentsListResponse struct {
	Namespace string    `json:"namespace"`
	Segments  []Segment `json:"segments"`
}

type SegmentsUpsertRequest struct {
	Namespace string  `json:"namespace"`
	Segment   Segment `json:"segment"`
}

type IDListRequest struct {
	Namespace string   `json:"namespace"`
	IDs       []string `json:"ids"`
}

type SegmentDryRunRequest struct {
	Namespace string         `json:"namespace"`
	UserID    string         `json:"user_id,omitempty"`
	Context   map[string]any `json:"context,omitempty"`
	Traits    map[string]any `json:"traits,omitempty"`
}

type SegmentDryRunResponse struct {
	Matched       bool   `json:"matched"`
	SegmentID     string `json:"segment_id,omitempty"`
	ProfileID     string `json:"profile_id,omitempty"`
	MatchedRuleID *int64 `json:"matched_rule_id,omitempty"`
}
