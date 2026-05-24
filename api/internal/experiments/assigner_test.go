package experiments

import (
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/api/internal/services/recsysvc"
)

func TestDeterministicAssignerStableVariant(t *testing.T) {
	assigner := NewDeterministicAssigner([]string{"A", "B"}, "salt")
	exp := &recsysvc.Experiment{ID: "exp-1"}
	user := recsysvc.UserRef{UserID: "user-123"}

	first := assigner.Assign(exp, user, "home", time.Now())
	second := assigner.Assign(exp, user, "home", time.Now())

	if first == nil || second == nil {
		t.Fatalf("expected assignment")
	}
	if first.Variant == "" {
		t.Fatalf("expected variant assigned")
	}
	if first.Variant != second.Variant {
		t.Fatalf("expected stable variant, got %q and %q", first.Variant, second.Variant)
	}
}

func TestDeterministicAssignerRespectsProvidedVariant(t *testing.T) {
	assigner := NewDeterministicAssigner([]string{"A", "B"}, "salt")
	exp := &recsysvc.Experiment{ID: "exp-2", Variant: "B"}
	user := recsysvc.UserRef{UserID: "user-123"}

	out := assigner.Assign(exp, user, "home", time.Now())
	if out.Variant != "B" {
		t.Fatalf("expected variant B, got %q", out.Variant)
	}
}

func TestDeterministicAssignerNoSubject(t *testing.T) {
	assigner := NewDeterministicAssigner([]string{"A", "B"}, "salt")
	exp := &recsysvc.Experiment{ID: "exp-3"}
	out := assigner.Assign(exp, recsysvc.UserRef{}, "home", time.Now())
	if out.Variant != "" {
		t.Fatalf("expected empty variant, got %q", out.Variant)
	}
}

func TestConfiguredAssignerRespectsDisabledExperiment(t *testing.T) {
	assigner := NewConfiguredAssigner([]string{"A", "B"}, "salt", []Definition{
		{ID: "exp-disabled", Enabled: false, TrafficPercent: 100},
	})

	out := assigner.Assign(&recsysvc.Experiment{ID: "exp-disabled"}, recsysvc.UserRef{UserID: "user-1"}, "home", time.Now())
	if out.Variant != "" {
		t.Fatalf("expected no assignment for disabled experiment, got %q", out.Variant)
	}
}

func TestConfiguredAssignerRespectsSurfaceAndDates(t *testing.T) {
	now := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	assigner := NewConfiguredAssigner([]string{"A", "B"}, "salt", []Definition{
		{
			ID:             "exp-home",
			Enabled:        true,
			TrafficPercent: 100,
			Surface:        "home",
			StartsAt:       now.Add(-time.Hour),
			EndsAt:         now.Add(time.Hour),
		},
	})

	wrongSurface := assigner.Assign(&recsysvc.Experiment{ID: "exp-home"}, recsysvc.UserRef{UserID: "user-1"}, "search", now)
	if wrongSurface.Variant != "" {
		t.Fatalf("expected no assignment for wrong surface, got %q", wrongSurface.Variant)
	}
	active := assigner.Assign(&recsysvc.Experiment{ID: "exp-home"}, recsysvc.UserRef{UserID: "user-1"}, "home", now)
	if active.Variant == "" {
		t.Fatalf("expected active assignment")
	}
	expired := assigner.Assign(&recsysvc.Experiment{ID: "exp-home"}, recsysvc.UserRef{UserID: "user-1"}, "home", now.Add(2*time.Hour))
	if expired.Variant != "" {
		t.Fatalf("expected no assignment after end, got %q", expired.Variant)
	}
}

func TestConfiguredAssignerRespectsTrafficAllocation(t *testing.T) {
	assigner := NewConfiguredAssigner([]string{"A", "B"}, "salt", []Definition{
		{ID: "exp-zero", Enabled: true, TrafficPercent: 0},
		{ID: "exp-full", Enabled: true, TrafficPercent: 100},
	})
	user := recsysvc.UserRef{UserID: "user-1"}

	zero := assigner.Assign(&recsysvc.Experiment{ID: "exp-zero"}, user, "home", time.Now())
	if zero.Variant != "" {
		t.Fatalf("expected no assignment at 0 traffic, got %q", zero.Variant)
	}
	full := assigner.Assign(&recsysvc.Experiment{ID: "exp-full"}, user, "home", time.Now())
	if full.Variant == "" {
		t.Fatalf("expected assignment at 100 traffic")
	}
}
