package types

import (
	"encoding/json"
	"time"
)

// SegmentProfile encapsulates ranking knobs for a segment.
type SegmentProfile struct {
	ProfileID           string
	Description         string
	BlendAlpha          float64
	BlendBeta           float64
	BlendGamma          float64
	MMRLambda           float64
	BrandCap            int
	CategoryCap         int
	ProfileBoost        float64
	ProfileWindowDays   float64
	ProfileTopN         int
	HalfLifeDays        float64
	CoVisWindowDays     int
	PurchasedWindowDays int
	RuleExcludeEvents   bool
	ExcludeEventTypes   []int16
	BrandTagPrefixes    []string
	CategoryTagPrefixes []string
	PopularityFanout    int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// SegmentRule represents a single rule expression for a segment.
type SegmentRule struct {
	RuleID      int64
	Rule        json.RawMessage
	Enabled     bool
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Segment binds a rule set to a profile.
type Segment struct {
	SegmentID   string
	Name        string
	Priority    int
	Active      bool
	ProfileID   string
	Description string
	Rules       []SegmentRule
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
