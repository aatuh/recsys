package experiments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"strings"

	"recsys/internal/services/recsysvc"
)

// Assigner determines experiment variants.
type Assigner interface {
	Assign(exp *recsysvc.Experiment, user recsysvc.UserRef) *recsysvc.Experiment
}

// DeterministicAssigner selects variants based on a stable hash.
type DeterministicAssigner struct {
	variants []string
	salt     []byte
}

// NewDeterministicAssigner constructs a deterministic assigner.
func NewDeterministicAssigner(variants []string, salt string) *DeterministicAssigner {
	clean := make([]string, 0, len(variants))
	for _, v := range variants {
		item := strings.TrimSpace(v)
		if item == "" {
			continue
		}
		clean = append(clean, item)
	}
	return &DeterministicAssigner{variants: clean, salt: []byte(salt)}
}

// Assign returns the experiment with a deterministic variant when missing.
func (a *DeterministicAssigner) Assign(exp *recsysvc.Experiment, user recsysvc.UserRef) *recsysvc.Experiment {
	if exp == nil {
		return nil
	}
	if strings.TrimSpace(exp.ID) == "" {
		return exp
	}
	if strings.TrimSpace(exp.Variant) != "" {
		return exp
	}
	if a == nil || len(a.variants) == 0 {
		return exp
	}
	subject := subjectKey(user)
	if subject == "" {
		return exp
	}
	variant := a.pickVariant(exp.ID, subject)
	assigned := *exp
	assigned.Variant = variant
	return &assigned
}

func (a *DeterministicAssigner) pickVariant(expID, subject string) string {
	if len(a.variants) == 1 {
		return a.variants[0]
	}
	mac := hmac.New(sha256.New, a.salt)
	_, _ = mac.Write([]byte(expID))
	_, _ = mac.Write([]byte(":"))
	_, _ = mac.Write([]byte(subject))
	sum := mac.Sum(nil)
	value := binary.BigEndian.Uint64(sum[:8])
	idx := int(value % uint64(len(a.variants)))
	return a.variants[idx]
}

func subjectKey(user recsysvc.UserRef) string {
	if v := strings.TrimSpace(user.UserID); v != "" {
		return v
	}
	if v := strings.TrimSpace(user.SessionID); v != "" {
		return v
	}
	return strings.TrimSpace(user.AnonymousID)
}
