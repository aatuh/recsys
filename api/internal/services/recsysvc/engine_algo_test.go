package recsysvc

import (
	"testing"

	"github.com/aatuh/recsys-algo/algorithm"
	"github.com/aatuh/recsys-algo/rules"
)

func TestApplyCandidateAllowListFilters(t *testing.T) {
	engine := &AlgoEngine{}
	req := RecommendRequest{
		K: 5,
		Candidates: &Candidates{
			IncludeIDs: []string{"item_1", "item_3"},
		},
	}
	items := []Item{
		{ItemID: "item_1", Score: 1.0},
		{ItemID: "item_2", Score: 0.9},
		{ItemID: "item_3", Score: 0.8},
	}

	got, warnings := engine.applyCandidateAllowList(req, items)
	if len(got) != 2 {
		t.Fatalf("expected 2 items after include filter, got %d", len(got))
	}
	if got[0].ItemID != "item_1" || got[1].ItemID != "item_3" {
		t.Fatalf("unexpected filtered order: %#v", got)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Code != "CANDIDATES_INCLUDE_FILTERED" {
		t.Fatalf("unexpected warning code: %s", warnings[0].Code)
	}
}

func TestApplyPinnedOverridesInjectsMissing(t *testing.T) {
	engine := &AlgoEngine{}
	req := RecommendRequest{K: 3}
	items := []Item{
		{ItemID: "item_1", Score: 1.0},
		{ItemID: "item_3", Score: 0.8},
	}
	trace := &algorithm.TraceData{
		RulePinned: []rules.PinnedItem{
			{ItemID: "item_1", Score: 1.0},
			{ItemID: "item_2", Score: 0.95},
		},
	}

	got, warnings := engine.applyPinnedOverrides(req, items, trace)
	if len(got) != 3 {
		t.Fatalf("expected 3 items after pin override, got %d", len(got))
	}
	if got[0].ItemID != "item_1" || got[1].ItemID != "item_2" || got[2].ItemID != "item_3" {
		t.Fatalf("unexpected pinned order: %#v", got)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Code != "RULE_PIN_INJECTED" {
		t.Fatalf("unexpected warning code: %s", warnings[0].Code)
	}
}
