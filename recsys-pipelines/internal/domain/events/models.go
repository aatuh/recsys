package events

import "time"

// ExposureEvent is emitted when an item is shown/served in a recommendation
// response. This is sufficient for popularity and co-occurrence pipelines.
type ExposureEvent struct {
	Version int       `json:"v"`
	TS      time.Time `json:"ts"`

	Tenant  string `json:"tenant"`
	Surface string `json:"surface"`

	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id,omitempty"`

	ItemID string `json:"item_id"`
	Rank   int    `json:"rank,omitempty"`
}
