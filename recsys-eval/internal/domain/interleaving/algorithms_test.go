package interleaving

import (
	"math/rand"
	"testing"
)

func TestTeamDraftDeterministic(t *testing.T) {
	listA := []string{"a", "b", "c"}
	listB := []string{"c", "d", "e"}
	//nolint:gosec // deterministic RNG for tests
	rng := rand.New(rand.NewSource(1))
	res := TeamDraft(listA, listB, 4, rng)
	if len(res.Items) != 4 {
		t.Fatalf("expected 4 items got %d", len(res.Items))
	}
	for i := range res.Items {
		if res.Attribution[i] != "A" && res.Attribution[i] != "B" {
			t.Fatalf("invalid attribution %q", res.Attribution[i])
		}
	}
}

func TestBalancedInterleaving(t *testing.T) {
	listA := []string{"a", "b", "c"}
	listB := []string{"d", "e", "f"}
	//nolint:gosec // deterministic RNG for tests
	rng := rand.New(rand.NewSource(1))
	res := BalancedInterleaving(listA, listB, 6, rng)
	countA := 0
	countB := 0
	for _, a := range res.Attribution {
		switch a {
		case "A":
			countA++
		case "B":
			countB++
		}
	}
	if countA != countB {
		t.Fatalf("expected balanced counts got A=%d B=%d", countA, countB)
	}
}
