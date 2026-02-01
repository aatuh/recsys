package rules

import (
	"time"

	"github.com/google/uuid"
)

// RuleAction enumerates supported rule actions.
type RuleAction string

const (
	RuleActionBlock RuleAction = "BLOCK"
	RuleActionPin   RuleAction = "PIN"
	RuleActionBoost RuleAction = "BOOST"
)

// RuleTarget enumerates the supported rule target dimensions.
type RuleTarget string

const (
	RuleTargetItem     RuleTarget = "ITEM"
	RuleTargetTag      RuleTarget = "TAG"
	RuleTargetBrand    RuleTarget = "BRAND"
	RuleTargetCategory RuleTarget = "CATEGORY"
)

// Rule represents a deterministic merchandising rule.
type Rule struct {
	RuleID           uuid.UUID
	ManualOverrideID *uuid.UUID
	OrgID            uuid.UUID
	Namespace        string
	Surface          string
	Name             string
	Description      string
	Action           RuleAction
	TargetType       RuleTarget
	TargetKey        string
	ItemIDs          []string
	BoostValue       *float64
	MaxPins          *int
	SegmentID        string
	Priority         int
	Enabled          bool
	ValidFrom        *time.Time
	ValidUntil       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// RuleListFilters captures optional filters for listing rules.
type RuleListFilters struct {
	Surface    string
	SegmentID  string
	Enabled    *bool
	ActiveAt   *time.Time
	Action     *RuleAction
	TargetType *RuleTarget
}

// RuleScope represents namespace/surface (+ optional segment) lookup key.
type RuleScope struct {
	Namespace string
	Surface   string
	SegmentID string
}
