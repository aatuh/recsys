package types

// RulePayload represents the payload for creating or updating a rule.
type RulePayload struct {
	Namespace   string   `json:"namespace"`
	Surface     string   `json:"surface"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Action      string   `json:"action"`
	TargetType  string   `json:"target_type"`
	TargetKey   string   `json:"target_key,omitempty"`
	ItemIDs     []string `json:"item_ids,omitempty"`
	BoostValue  *float64 `json:"boost_value,omitempty"`
	MaxPins     *int     `json:"max_pins,omitempty"`
	SegmentID   string   `json:"segment_id,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	Enabled     *bool    `json:"enabled,omitempty"`
	ValidFrom   string   `json:"valid_from,omitempty"`
	ValidUntil  string   `json:"valid_until,omitempty"`
}

// RuleResponse represents a persisted merchandising rule.
type RuleResponse struct {
	RuleID      string   `json:"rule_id"`
	Namespace   string   `json:"namespace"`
	Surface     string   `json:"surface"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Action      string   `json:"action"`
	TargetType  string   `json:"target_type"`
	TargetKey   string   `json:"target_key,omitempty"`
	ItemIDs     []string `json:"item_ids,omitempty"`
	BoostValue  *float64 `json:"boost_value,omitempty"`
	MaxPins     *int     `json:"max_pins,omitempty"`
	SegmentID   string   `json:"segment_id,omitempty"`
	Priority    int      `json:"priority"`
	Enabled     bool     `json:"enabled"`
	ValidFrom   string   `json:"valid_from,omitempty"`
	ValidUntil  string   `json:"valid_until,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// RulesListResponse wraps a set of rule responses.
type RulesListResponse struct {
	Rules []RuleResponse `json:"rules"`
}

// RuleMatchResponse describes matched rule metadata in dry-run responses.
type RuleMatchResponse struct {
	RuleID  string   `json:"rule_id"`
	Action  string   `json:"action"`
	Target  string   `json:"target_type"`
	ItemIDs []string `json:"item_ids"`
}

// RuleItemEffectResponse captures per-item rule effects.
type RuleItemEffectResponse struct {
	Blocked    bool    `json:"blocked"`
	Pinned     bool    `json:"pinned"`
	BoostDelta float64 `json:"boost_delta"`
}

// RuleDryRunPinnedItem describes pinned item metadata in dry-run output.
type RuleDryRunPinnedItem struct {
	ItemID         string   `json:"item_id"`
	RuleIDs        []string `json:"rule_ids"`
	FromCandidates bool     `json:"from_candidates"`
}

// RuleDryRunRequest represents a dry-run evaluation payload.
type RuleDryRunRequest struct {
	Namespace string   `json:"namespace"`
	Surface   string   `json:"surface"`
	SegmentID string   `json:"segment_id,omitempty"`
	Items     []string `json:"items"`
}

// RuleDryRunResponse summarises dry-run evaluation results.
type RuleDryRunResponse struct {
	RulesEvaluated []string                          `json:"rules_evaluated"`
	RulesMatched   []RuleMatchResponse               `json:"rules_matched"`
	ItemEffects    map[string]RuleItemEffectResponse `json:"rule_effects_per_item"`
	ReasonTags     map[string][]string               `json:"reason_tags,omitempty"`
	PinnedPreview  []RuleDryRunPinnedItem            `json:"pinned_items,omitempty"`
}
