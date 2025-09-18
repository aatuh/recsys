package rules

import (
	"time"

	"github.com/google/uuid"
	"recsys/internal/types"
)

// EvaluateRequest carries all inputs required to evaluate rules for a response.
type EvaluateRequest struct {
	OrgID      uuid.UUID
	Namespace  string
	Surface    string
	SegmentID  string
	Now        time.Time
	Candidates []types.ScoredItem
	ItemTags   map[string][]string
	// BrandTagPrefixes and CategoryTagPrefixes help derive brand/category targets from tags.
	BrandTagPrefixes    []string
	CategoryTagPrefixes []string
}

// Match captures rule-to-item matches for auditing and dry-run responses.
type Match struct {
	RuleID  uuid.UUID
	Action  types.RuleAction
	Target  types.RuleTarget
	ItemIDs []string
}

// BoostDetail stores per-rule boost contribution for an item.
type BoostDetail struct {
	RuleID uuid.UUID
	Delta  float64
}

// ItemEffect aggregates final per-item effects after precedence resolution.
type ItemEffect struct {
	Blocked    bool
	Pinned     bool
	BoostDelta float64
	BlockRules []uuid.UUID
	PinRules   []uuid.UUID
	BoostRules []BoostDetail
}

// PinnedItem represents a pinned recommendation that should precede ranked items.
type PinnedItem struct {
	ItemID         string
	Score          float64
	FromCandidates bool
	Rules          []uuid.UUID
}

// EvaluateResult returns transformed candidates plus metadata for explain/audit.
type EvaluateResult struct {
	Candidates       []types.ScoredItem
	Pinned           []PinnedItem
	Matches          []Match
	EvaluatedRuleIDs []uuid.UUID
	ItemEffects      map[string]ItemEffect
	ReasonTags       map[string][]string
}
