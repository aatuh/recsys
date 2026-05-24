package experiments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/api/internal/services/recsysvc"
)

// Assigner determines experiment variants.
type Assigner interface {
	Assign(exp *recsysvc.Experiment, user recsysvc.UserRef, surface string, now time.Time) *recsysvc.Experiment
}

// DeterministicAssigner selects variants based on a stable hash.
type DeterministicAssigner struct {
	variants    []string
	salt        []byte
	definitions map[string]Definition
}

// Definition controls a tenant-independent experiment lifecycle.
type Definition struct {
	ID             string
	Enabled        bool
	Variants       []string
	TrafficPercent float64
	Surface        string
	StartsAt       time.Time
	EndsAt         time.Time
}

// NewDeterministicAssigner constructs a deterministic assigner.
func NewDeterministicAssigner(variants []string, salt string) *DeterministicAssigner {
	return NewConfiguredAssigner(variants, salt, nil)
}

// NewConfiguredAssigner constructs an assigner with optional experiment lifecycle definitions.
func NewConfiguredAssigner(variants []string, salt string, definitions []Definition) *DeterministicAssigner {
	clean := make([]string, 0, len(variants))
	for _, v := range variants {
		item := strings.TrimSpace(v)
		if item == "" {
			continue
		}
		clean = append(clean, item)
	}
	byID := make(map[string]Definition, len(definitions))
	for _, def := range definitions {
		id := strings.TrimSpace(def.ID)
		if id == "" {
			continue
		}
		def.ID = id
		def.Variants = cleanVariants(def.Variants)
		byID[id] = def
	}
	return &DeterministicAssigner{variants: clean, salt: []byte(salt), definitions: byID}
}

// Assign returns the experiment with a deterministic variant when missing.
func (a *DeterministicAssigner) Assign(exp *recsysvc.Experiment, user recsysvc.UserRef, surface string, now time.Time) *recsysvc.Experiment {
	if exp == nil {
		return nil
	}
	expID := strings.TrimSpace(exp.ID)
	if expID == "" {
		return exp
	}
	if strings.TrimSpace(exp.Variant) != "" {
		return exp
	}
	if a == nil {
		return exp
	}
	subject := subjectKey(user)
	if subject == "" {
		return exp
	}
	variants := a.variants
	if def, ok := a.definitions[expID]; ok {
		if !def.active(surface, now) || !a.inTraffic(expID, subject, def.TrafficPercent) {
			return exp
		}
		if len(def.Variants) > 0 {
			variants = def.Variants
		}
	}
	if len(variants) == 0 {
		return exp
	}
	variant := a.pickVariant(expID, subject, variants)
	assigned := *exp
	assigned.Variant = variant
	return &assigned
}

func (a *DeterministicAssigner) pickVariant(expID, subject string, variants []string) string {
	if len(variants) == 1 {
		return variants[0]
	}
	mac := hmac.New(sha256.New, a.salt)
	_, _ = mac.Write([]byte(expID))
	_, _ = mac.Write([]byte(":"))
	_, _ = mac.Write([]byte(subject))
	sum := mac.Sum(nil)
	value := binary.BigEndian.Uint64(sum[:8])
	idx := int(value % uint64(len(variants))) // #nosec G115 -- modulo by len(variants) bounds the value to a valid slice index.
	return variants[idx]
}

func (a *DeterministicAssigner) inTraffic(expID, subject string, trafficPercent float64) bool {
	if trafficPercent >= 100 {
		return true
	}
	if trafficPercent <= 0 {
		return false
	}
	mac := hmac.New(sha256.New, a.salt)
	_, _ = mac.Write([]byte("traffic:"))
	_, _ = mac.Write([]byte(expID))
	_, _ = mac.Write([]byte(":"))
	_, _ = mac.Write([]byte(subject))
	sum := mac.Sum(nil)
	bucket := binary.BigEndian.Uint64(sum[:8]) % 10000
	return float64(bucket) < trafficPercent*100
}

func (d Definition) active(surface string, now time.Time) bool {
	if !d.Enabled {
		return false
	}
	if d.Surface != "" && !strings.EqualFold(strings.TrimSpace(d.Surface), strings.TrimSpace(surface)) {
		return false
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	if !d.StartsAt.IsZero() && now.Before(d.StartsAt.UTC()) {
		return false
	}
	if !d.EndsAt.IsZero() && !now.Before(d.EndsAt.UTC()) {
		return false
	}
	return true
}

func cleanVariants(variants []string) []string {
	clean := make([]string, 0, len(variants))
	seen := map[string]struct{}{}
	for _, variant := range variants {
		v := strings.TrimSpace(variant)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		clean = append(clean, v)
	}
	return clean
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
