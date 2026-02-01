package dataset

import "time"

// Exposure is a single recommendation exposure (served list).
type Exposure struct {
	RequestID string            `json:"request_id"`
	UserID    string            `json:"user_id"`
	Timestamp time.Time         `json:"ts"`
	Items     []ExposedItem     `json:"items"`
	Context   map[string]string `json:"context,omitempty"`
	LatencyMs *float64          `json:"latency_ms,omitempty"`
	Error     *bool             `json:"error,omitempty"`
}

// ExposedItem is a recommended item in an exposure list.
type ExposedItem struct {
	ItemID            string   `json:"item_id"`
	Rank              int      `json:"rank"`
	Propensity        *float64 `json:"propensity,omitempty"`
	LoggingPropensity *float64 `json:"logging_propensity,omitempty"`
	TargetPropensity  *float64 `json:"target_propensity,omitempty"`
}

// RankList represents a ranked list for interleaving comparisons.
type RankList struct {
	RequestID string            `json:"request_id"`
	UserID    string            `json:"user_id"`
	Timestamp time.Time         `json:"ts"`
	Items     []string          `json:"items"`
	Context   map[string]string `json:"context,omitempty"`
}

// Outcome is a user action tied to an exposure.
type Outcome struct {
	RequestID string    `json:"request_id"`
	UserID    string    `json:"user_id"`
	ItemID    string    `json:"item_id"`
	EventType string    `json:"event_type"` // click | conversion
	Value     float64   `json:"value,omitempty"`
	Timestamp time.Time `json:"ts"`
}

// Assignment is an experiment assignment event.
type Assignment struct {
	ExperimentID string            `json:"experiment_id"`
	Variant      string            `json:"variant"`
	RequestID    string            `json:"request_id"`
	UserID       string            `json:"user_id"`
	Timestamp    time.Time         `json:"ts"`
	Context      map[string]string `json:"context,omitempty"`
}
