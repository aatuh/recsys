package bandit

import (
	"crypto/sha1"
	"encoding/hex"
	"recsys/internal/types"
	"sort"
	"strings"
)

// Decision represents a chosen policy for a request.
type Decision struct {
	PolicyID  string            `json:"policy_id"`
	Algorithm types.Algorithm   `json:"algorithm"`
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
	Algorithm types.Algorithm
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
