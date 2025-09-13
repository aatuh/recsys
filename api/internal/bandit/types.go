package bandit

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
)

type Algorithm string

const (
	AlgorithmThompson Algorithm = "thompson"
	AlgorithmUCB1     Algorithm = "ucb1"
)

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

// Decision represents a chosen policy for a request.
type Decision struct {
	PolicyID  string            `json:"policy_id"`
	Algorithm Algorithm         `json:"algorithm"`
	Surface   string            `json:"surface"`
	BucketKey string            `json:"bucket_key"`
	Explore   bool              `json:"explore"`
	Explain   map[string]string `json:"explain"`
}

// RewardInput is the minimal data needed to update stats.
type RewardInput struct {
	PolicyID  string
	Surface   string
	BucketKey string
	Reward    bool
	Algorithm Algorithm
}

// BucketKeyFromContext builds a canonical key from simple context.
func BucketKeyFromContext(ctx map[string]string) string {
	if ctx == nil {
		return "ctx:empty"
	}
	// Only a few known keys for MVP, but keep others deterministically.
	keys := make([]string, 0, len(ctx))
	for k := range ctx {
		keys = append(keys, strings.ToLower(strings.TrimSpace(k)))
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := strings.ToLower(strings.TrimSpace(ctx[k]))
		parts = append(parts, k+"="+v)
	}
	raw := "ctx:" + strings.Join(parts, "|")
	// Keep readable, but protect against excessive length with a suffix hash.
	if len(raw) <= 200 {
		return raw
	}
	sum := sha1.Sum([]byte(raw))
	return raw[:200] + "|h=" + hex.EncodeToString(sum[:8])
}
