package types

// ManualOverrideRequest represents the payload to create a manual override.
type ManualOverrideRequest struct {
	Namespace  string   `json:"namespace"`
	Surface    string   `json:"surface"`
	ItemID     string   `json:"item_id"`
	Action     string   `json:"action"`
	BoostValue *float64 `json:"boost_value,omitempty"`
	Notes      string   `json:"notes,omitempty"`
	CreatedBy  string   `json:"created_by,omitempty"`
	ExpiresAt  string   `json:"expires_at,omitempty"`
	Priority   *int     `json:"priority,omitempty"`
}

// ManualOverrideResponse represents a stored manual override.
type ManualOverrideResponse struct {
	OverrideID  string   `json:"override_id"`
	Namespace   string   `json:"namespace"`
	Surface     string   `json:"surface"`
	ItemID      string   `json:"item_id"`
	Action      string   `json:"action"`
	BoostValue  *float64 `json:"boost_value,omitempty"`
	Notes       string   `json:"notes,omitempty"`
	CreatedBy   string   `json:"created_by,omitempty"`
	CreatedAt   string   `json:"created_at"`
	ExpiresAt   string   `json:"expires_at,omitempty"`
	Status      string   `json:"status"`
	RuleID      string   `json:"rule_id,omitempty"`
	CancelledAt string   `json:"cancelled_at,omitempty"`
	CancelledBy string   `json:"cancelled_by,omitempty"`
}

// ManualOverrideCancelRequest captures optional metadata when cancelling an override.
type ManualOverrideCancelRequest struct {
	CancelledBy string `json:"cancelled_by,omitempty"`
}
