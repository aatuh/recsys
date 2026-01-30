package exposure

import (
	"errors"
	"strings"
	"time"
)

const SchemaVersion = "v1"

// Event is the exposure logging schema.
type Event struct {
	SchemaVersion string      `json:"schema_version"`
	Timestamp     time.Time   `json:"timestamp"`
	RequestID     string      `json:"request_id,omitempty"`
	TenantID      string      `json:"tenant_id,omitempty"`
	Surface       string      `json:"surface,omitempty"`
	Segment       string      `json:"segment,omitempty"`
	AlgoVersion   string      `json:"algo_version,omitempty"`
	ConfigVersion string      `json:"config_version,omitempty"`
	RulesVersion  string      `json:"rules_version,omitempty"`
	Experiment    *Experiment `json:"experiment,omitempty"`
	Subject       *Subject    `json:"subject,omitempty"`
	Context       *Context    `json:"context,omitempty"`
	Items         []Item      `json:"items,omitempty"`
}

// Experiment captures experiment metadata.
type Experiment struct {
	ID      string `json:"id,omitempty"`
	Variant string `json:"variant,omitempty"`
}

// Subject holds pseudonymized identifiers.
type Subject struct {
	UserIDHash      string `json:"user_id_hash,omitempty"`
	AnonymousIDHash string `json:"anonymous_id_hash,omitempty"`
	SessionIDHash   string `json:"session_id_hash,omitempty"`
}

// Context captures request context attributes.
type Context struct {
	Locale  string `json:"locale,omitempty"`
	Device  string `json:"device,omitempty"`
	Country string `json:"country,omitempty"`
	Now     string `json:"now,omitempty"`
}

// Item records an exposed item and its rank.
type Item struct {
	ItemID string  `json:"item_id"`
	Rank   int     `json:"rank"`
	Score  float64 `json:"score,omitempty"`
}

// Normalize fills defaults such as schema version.
func (e *Event) Normalize() {
	if e == nil {
		return
	}
	if strings.TrimSpace(e.SchemaVersion) == "" {
		e.SchemaVersion = SchemaVersion
	}
}

// Validate checks the event shape for required fields.
func (e Event) Validate() error {
	if strings.TrimSpace(e.SchemaVersion) == "" {
		return errors.New("schema_version is required")
	}
	for _, item := range e.Items {
		if strings.TrimSpace(item.ItemID) == "" {
			return errors.New("item_id is required")
		}
		if item.Rank <= 0 {
			return errors.New("item rank must be positive")
		}
	}
	if e.Subject != nil {
		if e.Subject.UserIDHash == "" && e.Subject.AnonymousIDHash == "" && e.Subject.SessionIDHash == "" {
			return errors.New("subject hashes must be set when subject is provided")
		}
	}
	return nil
}
