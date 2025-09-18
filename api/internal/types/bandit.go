package types

import (
	"context"
	"errors"
	"strings"
)

// Algorithm is the type of bandit algorithm.
type Algorithm string

// Available algorithms.
const (
	AlgorithmThompson Algorithm = "thompson"
	AlgorithmUCB1     Algorithm = "ucb1"
)

// String returns the string representation of the algorithm.
func (a Algorithm) String() string {
	return string(a)
}

// ParseAlgorithm converts a string to an Algorithm, returning an error if
// invalid.
func ParseAlgorithm(s string) (Algorithm, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(AlgorithmThompson):
		return AlgorithmThompson, nil
	case string(AlgorithmUCB1):
		return AlgorithmUCB1, nil
	default:
		return "", errors.New("invalid algorithm: " + s)
	}
}

// PolicyConfig describes the scoring configuration used by the ranker.
type PolicyConfig struct {
	// Human-friendly id and name.
	PolicyID string `json:"policy_id"`
	Name     string `json:"name"`

	// Switch to disable policy without deleting it.
	Active bool `json:"active"`

	// Scoring knobs used by the ranker.
	BlendAlpha  float64 `json:"blend_alpha"`
	BlendBeta   float64 `json:"blend_beta"`
	BlendGamma  float64 `json:"blend_gamma"`
	MMRLambda   float64 `json:"mmr_lambda"`
	BrandCap    int     `json:"brand_cap"`
	CategoryCap int     `json:"category_cap"`

	// Personalization and filtering parameters
	ProfileBoost      float64 `json:"profile_boost"`
	RuleExcludeEvents bool    `json:"rule_exclude_events"`
	HalfLifeDays      float64 `json:"half_life_days"`
	CoVisWindowDays   int     `json:"co_vis_window_days"`
	PopularityFanout  int     `json:"popularity_fanout"`

	// Free-form field for notes.
	Notes string `json:"notes,omitempty"`
}

// Stats keeps online stats per (surface, bucket, policy, algo).
type Stats struct {
	Trials    int64   // total impressions/decisions
	Successes int64   // successes (binary reward)
	Alpha     float64 // prior alpha (Thompson)
	Beta      float64 // prior beta (Thompson)
}

// BanditStore is the minimal interface the bandit needs from persistence.
// This keeps the bandit logic independent of the concrete DB layer.
type BanditStore interface {
	// Policies

	// ListActivePolicies returns currently active policy configs for the
	// org/namespace. Order is not guaranteed.
	ListActivePolicies(
		ctx context.Context,
		orgID string,
		ns string,
	) ([]PolicyConfig, error)

	// ListPoliciesByIDs returns policy configs for the given policy IDs.
	// Missing IDs may be absent from the result.
	ListPoliciesByIDs(
		ctx context.Context,
		orgID string,
		ns string,
		ids []string,
	) ([]PolicyConfig, error)

	// Stats

	// GetStats returns current stats keyed by policy_id for the provided
	// (surface, bucket, algorithm).
	GetStats(
		ctx context.Context,
		orgID string,
		ns string,
		surface string,
		bucket string,
		algo Algorithm,
	) (map[string]Stats, error)

	// IncrementStats increments trials and optionally successes for the
	// policy. If reward is true, both trials and successes are incremented;
	// if false, only trials is incremented. The row should be created if it
	// does not exist.
	IncrementStats(
		ctx context.Context,
		orgID string,
		ns string,
		surface string,
		bucket string,
		algo Algorithm,
		policyID string,
		reward bool,
	) error

	// Logging

	// LogDecision records a bandit decision, including whether it was an
	// exploration. reqID can be used to correlate with rewards. meta is
	// optional free-form context.
	LogDecision(
		ctx context.Context,
		orgID string,
		ns string,
		surface string,
		bucket string,
		algo Algorithm,
		policyID string,
		explore bool,
		reqID string,
		meta map[string]any,
	) error

	// LogReward records an observed reward outcome for a prior decision.
	// reqID should align with the decision when available; meta is optional
	// free-form context.
	LogReward(
		ctx context.Context,
		orgID string,
		ns string,
		surface string,
		bucket string,
		algo Algorithm,
		policyID string,
		reward bool,
		reqID string,
		meta map[string]any,
	) error
}
