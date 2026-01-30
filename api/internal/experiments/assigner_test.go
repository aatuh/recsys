package experiments

import (
	"testing"

	"recsys/internal/services/recsysvc"
)

func TestDeterministicAssignerStableVariant(t *testing.T) {
	assigner := NewDeterministicAssigner([]string{"A", "B"}, "salt")
	exp := &recsysvc.Experiment{ID: "exp-1"}
	user := recsysvc.UserRef{UserID: "user-123"}

	first := assigner.Assign(exp, user)
	second := assigner.Assign(exp, user)

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

	out := assigner.Assign(exp, user)
	if out.Variant != "B" {
		t.Fatalf("expected variant B, got %q", out.Variant)
	}
}

func TestDeterministicAssignerNoSubject(t *testing.T) {
	assigner := NewDeterministicAssigner([]string{"A", "B"}, "salt")
	exp := &recsysvc.Experiment{ID: "exp-3"}
	out := assigner.Assign(exp, recsysvc.UserRef{})
	if out.Variant != "" {
		t.Fatalf("expected empty variant, got %q", out.Variant)
	}
}
