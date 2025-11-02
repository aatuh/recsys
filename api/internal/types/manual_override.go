package types

import (
	"time"

	"github.com/google/uuid"
)

// ManualOverrideAction represents the supported manual merchandising actions.
type ManualOverrideAction string

const (
	ManualOverrideActionBoost     ManualOverrideAction = "boost"
	ManualOverrideActionSuppress  ManualOverrideAction = "suppress"
)

// ManualOverrideStatus captures lifecycle stages for manual overrides.
type ManualOverrideStatus string

const (
	ManualOverrideStatusActive    ManualOverrideStatus = "active"
	ManualOverrideStatusCancelled ManualOverrideStatus = "cancelled"
	ManualOverrideStatusExpired   ManualOverrideStatus = "expired"
)

// ManualOverride represents an ad-hoc merchandising adjustment.
type ManualOverride struct {
	OverrideID uuid.UUID
	OrgID      uuid.UUID
	Namespace  string
	Surface    string
	ItemID     string
	Action     ManualOverrideAction
	BoostValue *float64
	Notes      string
	CreatedBy  string
	CreatedAt  time.Time
	ExpiresAt  *time.Time
	RuleID     *uuid.UUID
	Status     ManualOverrideStatus
	CancelledAt *time.Time
	CancelledBy string
}

// ManualOverrideCreate captures the payload required to create an override.
type ManualOverrideCreate struct {
	Namespace string
	Surface   string
	ItemID    string
	Action    ManualOverrideAction
	BoostValue *float64
	Notes     string
	CreatedBy string
	ExpiresAt *time.Time
}

// ManualOverrideFilters controls list queries.
type ManualOverrideFilters struct {
	Status ManualOverrideStatus
	Action ManualOverrideAction
	IncludeExpired bool
}
