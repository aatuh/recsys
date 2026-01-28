package rules

import (
	"context"
	"strings"
	"testing"

	recmodel "github.com/aatuh/recsys-algo/model"

	"github.com/google/uuid"
)

func TestEvaluatorPrecedenceBlockPinBoost(t *testing.T) {
	pinRuleID := uuid.New()
	blockRuleID := uuid.New()

	rules := []Rule{
		{
			RuleID:     pinRuleID,
			Action:     RuleActionPin,
			TargetType: RuleTargetItem,
			ItemIDs:    []string{"a"},
			Priority:   100,
		},
		{
			RuleID:     blockRuleID,
			Action:     RuleActionBlock,
			TargetType: RuleTargetItem,
			ItemIDs:    []string{"a"},
			Priority:   50,
		},
	}

	eval := evaluator{maxPinSlots: 3}
	req := EvaluateRequest{
		Candidates: []recmodel.ScoredItem{{ItemID: "a", Score: 1.0}, {ItemID: "b", Score: 1.0}},
		ItemTags:   map[string][]string{"a": {"tag:a"}, "b": {"tag:b"}},
	}

	res, err := eval.apply(context.Background(), rules, req)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(res.Pinned) != 0 {
		t.Fatalf("expected no pinned items after block precedence, got %d", len(res.Pinned))
	}
	if len(res.Candidates) != 1 || res.Candidates[0].ItemID != "b" {
		t.Fatalf("expected only candidate b remaining, got %#v", res.Candidates)
	}
	eff, ok := res.ItemEffects["a"]
	if !ok {
		t.Fatalf("expected effects for item a")
	}
	if !eff.Blocked {
		t.Fatalf("expected item a to be blocked")
	}
	if eff.Pinned {
		t.Fatalf("expected final pinned to be false when blocked")
	}
	if len(res.ReasonTags["a"]) == 0 {
		t.Fatalf("expected reason tags for blocked item a")
	}
	found := false
	for _, tag := range res.ReasonTags["a"] {
		if strings.HasPrefix(tag, "rule.block[") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected rule.block reason tag for item a, got %v", res.ReasonTags["a"])
	}
}

func TestEvaluatorPinSlotsCap(t *testing.T) {
	rules := []Rule{
		{
			RuleID:     uuid.New(),
			Action:     RuleActionPin,
			TargetType: RuleTargetItem,
			ItemIDs:    []string{"a", "b", "c"},
			Priority:   100,
		},
	}
	eval := evaluator{maxPinSlots: 2}
	req := EvaluateRequest{
		Candidates: []recmodel.ScoredItem{{ItemID: "a", Score: 1}, {ItemID: "b", Score: 1}, {ItemID: "c", Score: 1}},
	}
	res, err := eval.apply(context.Background(), rules, req)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(res.Pinned) != 2 {
		t.Fatalf("expected 2 pinned items, got %d", len(res.Pinned))
	}
	for _, pin := range res.Pinned {
		if pin.ItemID == "c" {
			t.Fatalf("expected item c not to be pinned due to cap")
		}
	}
	if len(res.Candidates) != 1 {
		t.Fatalf("expected 1 candidate left, got %d", len(res.Candidates))
	}
}

func TestEvaluatorBoostAdjustsScore(t *testing.T) {
	boost := 0.5
	ruleID := uuid.New()
	rules := []Rule{
		{
			RuleID:     ruleID,
			Action:     RuleActionBoost,
			TargetType: RuleTargetItem,
			ItemIDs:    []string{"a"},
			BoostValue: &boost,
			Priority:   10,
		},
	}
	eval := evaluator{maxPinSlots: 3}
	req := EvaluateRequest{
		Candidates: []recmodel.ScoredItem{{ItemID: "a", Score: 1.0}},
	}
	res, err := eval.apply(context.Background(), rules, req)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(res.Candidates) != 1 || res.Candidates[0].Score != 1.5 {
		t.Fatalf("expected boosted score 1.5, got %#v", res.Candidates)
	}
	eff := res.ItemEffects["a"]
	if eff.BoostDelta != 0.5 {
		t.Fatalf("expected boost delta 0.5, got %f", eff.BoostDelta)
	}
	if len(res.ReasonTags["a"]) == 0 {
		t.Fatalf("expected boost reason tag")
	}
	found := false
	for _, tag := range res.ReasonTags["a"] {
		if strings.HasPrefix(tag, "rule.boost:+0.50[") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected formatted boost reason, got %v", res.ReasonTags["a"])
	}
}

func TestEvaluatorBoostInjectsMissingCandidate(t *testing.T) {
	boost := 2.0
	rules := []Rule{
		{
			RuleID:     uuid.New(),
			Action:     RuleActionBoost,
			TargetType: RuleTargetItem,
			ItemIDs:    []string{"new"},
			BoostValue: &boost,
			Priority:   50,
		},
	}
	eval := evaluator{maxPinSlots: 3}
	req := EvaluateRequest{
		Candidates: []recmodel.ScoredItem{{ItemID: "existing", Score: 1.0}},
	}
	res, err := eval.apply(context.Background(), rules, req)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(res.Candidates) != 2 {
		t.Fatalf("expected 2 candidates after injection, got %d", len(res.Candidates))
	}
	var injected *recmodel.ScoredItem
	for i := range res.Candidates {
		if res.Candidates[i].ItemID == "new" {
			injected = &res.Candidates[i]
			break
		}
	}
	if injected == nil {
		t.Fatalf("expected new candidate to be appended, got %#v", res.Candidates)
	}
	if injected.Score != boost {
		t.Fatalf("expected injected score %f, got %f", boost, injected.Score)
	}
	eff, ok := res.ItemEffects["new"]
	if !ok || eff.BoostDelta != boost {
		t.Fatalf("expected boost effect recorded for new item, got %#v", eff)
	}
	if len(res.ReasonTags["new"]) == 0 {
		t.Fatalf("expected reason tags for injected item")
	}
}

func TestEvaluatorPinsExternalItems(t *testing.T) {
	rules := []Rule{
		{
			RuleID:     uuid.New(),
			Action:     RuleActionPin,
			TargetType: RuleTargetItem,
			ItemIDs:    []string{"x"},
			Priority:   100,
		},
	}
	eval := evaluator{maxPinSlots: 3}
	req := EvaluateRequest{}
	res, err := eval.apply(context.Background(), rules, req)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(res.Pinned) != 1 || res.Pinned[0].ItemID != "x" {
		t.Fatalf("expected pinned item x, got %#v", res.Pinned)
	}
	if res.Pinned[0].FromCandidates {
		t.Fatalf("expected pinned item to be marked as external")
	}
}
