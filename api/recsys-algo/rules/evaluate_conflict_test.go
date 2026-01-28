package rules

import (
	"context"
	"testing"

	recmodel "github.com/aatuh/recsys-algo/model"

	"github.com/google/uuid"
)

func TestEvaluatorConflictsRespectPriority(t *testing.T) {
	boostHigh := 1.0
	boostLow := 0.25
	pinHigh := uuid.New()
	pinLow := uuid.New()

	rules := []Rule{
		{
			RuleID:     uuid.New(),
			Action:     RuleActionBlock,
			TargetType: RuleTargetItem,
			ItemIDs:    []string{"blk"},
			Priority:   200,
		},
		{
			RuleID:     uuid.New(),
			Action:     RuleActionBoost,
			TargetType: RuleTargetItem,
			ItemIDs:    []string{"a"},
			BoostValue: &boostLow,
			Priority:   50,
		},
		{
			RuleID:     uuid.New(),
			Action:     RuleActionBoost,
			TargetType: RuleTargetItem,
			ItemIDs:    []string{"a"},
			BoostValue: &boostHigh,
			Priority:   150,
		},
		{
			RuleID:     pinLow,
			Action:     RuleActionPin,
			TargetType: RuleTargetItem,
			ItemIDs:    []string{"pin"},
			Priority:   50,
		},
		{
			RuleID:     pinHigh,
			Action:     RuleActionPin,
			TargetType: RuleTargetItem,
			ItemIDs:    []string{"pin"},
			Priority:   300,
		},
	}

	req := EvaluateRequest{
		Candidates: []recmodel.ScoredItem{
			{ItemID: "a", Score: 1.0},
			{ItemID: "blk", Score: 1.0},
			{ItemID: "pin", Score: 0.01},
		},
	}
	eval := evaluator{maxPinSlots: 1}
	res, err := eval.apply(context.Background(), rules, req)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(res.Candidates) != 1 || res.Candidates[0].ItemID != "a" {
		t.Fatalf("expected only boosted candidate a to survive, got %#v", res.Candidates)
	}
	if res.Candidates[0].Score <= 2.0 {
		t.Fatalf("expected high-priority boost applied, got %f", res.Candidates[0].Score)
	}
	eff := res.ItemEffects["a"]
	if len(eff.BoostRules) != 2 {
		t.Fatalf("expected only high priority boost recorded, got %#v", eff.BoostRules)
	}

	if _, blocked := res.ItemEffects["blk"]; !blocked {
		t.Fatalf("blocked item should record effects")
	}
	if len(res.Candidates) != 1 {
		t.Fatalf("blocked item should be excluded from candidates")
	}

	if len(res.Pinned) != 1 || res.Pinned[0].ItemID != "pin" {
		t.Fatalf("expected pin item pinned once, got %#v", res.Pinned)
	}
	if len(res.Pinned[0].Rules) != 1 || res.Pinned[0].Rules[0] != pinHigh {
		t.Fatalf("expected pin from highest priority rule, got %#v", res.Pinned[0].Rules)
	}
}
