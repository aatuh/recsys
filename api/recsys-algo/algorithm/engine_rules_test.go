package algorithm

import (
	"context"
	"testing"
	"time"

	"github.com/aatuh/recsys-algo/rules"

	recmodel "github.com/aatuh/recsys-algo/model"

	"github.com/google/uuid"
)

type stubRuleStore struct {
	rules []rules.Rule
}

func (s *stubRuleStore) ListActiveRulesForScope(ctx context.Context, orgID uuid.UUID, namespace, surface, segmentID string, ts time.Time) ([]rules.Rule, error) {
	copied := make([]rules.Rule, len(s.rules))
	copy(copied, s.rules)
	return copied, nil
}

func TestBuildResponseAppliesRuleEffects(t *testing.T) {
	blockID := uuid.New()
	pinID := uuid.New()
	boostID := uuid.New()
	boostValue := 1.0

	ruleStore := &stubRuleStore{rules: []rules.Rule{
		{
			RuleID:     blockID,
			Namespace:  "default",
			Surface:    "home",
			Name:       "block-a",
			Action:     rules.RuleActionBlock,
			TargetType: rules.RuleTargetItem,
			ItemIDs:    []string{"a"},
			Priority:   100,
		},
		{
			RuleID:     pinID,
			Namespace:  "default",
			Surface:    "home",
			Name:       "pin-c",
			Action:     rules.RuleActionPin,
			TargetType: rules.RuleTargetItem,
			ItemIDs:    []string{"c"},
			Priority:   90,
		},
		{
			RuleID:     boostID,
			Namespace:  "default",
			Surface:    "home",
			Name:       "boost-b",
			Action:     rules.RuleActionBoost,
			TargetType: rules.RuleTargetItem,
			ItemIDs:    []string{"b"},
			BoostValue: &boostValue,
			Priority:   80,
		},
	}}

	mgr := rules.NewManager(ruleStore, rules.ManagerOptions{
		RefreshInterval: time.Minute,
		MaxPinSlots:     3,
		Enabled:         true,
	})

	eng := NewEngine(Config{RulesEnabled: true}, nil, mgr)

	data := &CandidateData{
		Candidates: []recmodel.ScoredItem{{ItemID: "a", Score: 1.0}, {ItemID: "b", Score: 1.0}, {ItemID: "c", Score: 1.0}},
		Tags: map[string]recmodel.ItemTags{
			"a": {ItemID: "a", Tags: []string{"tag:a"}},
			"b": {ItemID: "b", Tags: []string{"tag:b"}},
			"c": {ItemID: "c", Tags: []string{"tag:c"}},
		},
		Boosted:           map[string]bool{},
		PopNorm:           map[string]float64{},
		CoocNorm:          map[string]float64{},
		SimilarityNorm:    map[string]float64{},
		PopRaw:            map[string]float64{},
		CoocRaw:           map[string]float64{},
		SimilarityRaw:     map[string]float64{},
		SimilaritySources: map[string][]Signal{},
		ProfileOverlap:    map[string]float64{},
		ProfileMultiplier: map[string]float64{},
		MMRInfo:           map[string]MMRExplain{},
		CapsInfo:          map[string]CapsExplain{},
	}

	req := Request{
		OrgID:     uuid.New(),
		Namespace: "default",
		Surface:   "home",
		K:         3,
	}

	result, err := eng.applyRules(context.Background(), req, data)
	if err != nil {
		t.Fatalf("applyRules error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected rule result")
	}

	if len(data.Candidates) != 1 || data.Candidates[0].ItemID != "b" {
		t.Fatalf("expected only boosted candidate b remaining, got %#v", data.Candidates)
	}
	if data.Candidates[0].Score <= 1.0 {
		t.Fatalf("expected boosted score for b, got %f", data.Candidates[0].Score)
	}
	if effect := result.ItemEffects["a"]; !effect.Blocked {
		t.Fatalf("expected item a to be blocked")
	}
	if effect := result.ItemEffects["c"]; !effect.Pinned {
		t.Fatalf("expected item c to be pinned")
	}

	reasonSink := make(map[string][]string)
	resp := eng.buildResponse(
		data,
		3,
		"test_model",
		true,
		ExplainLevelTags,
		BlendWeights{Pop: 1},
		reasonSink,
		result,
	)

	if len(resp.Items) != 2 {
		t.Fatalf("expected two items in response, got %d", len(resp.Items))
	}
	if resp.Items[0].ItemID != "c" {
		t.Fatalf("expected pinned item c first, got %s", resp.Items[0].ItemID)
	}
	if resp.Items[1].ItemID != "b" {
		t.Fatalf("expected boosted item b second, got %s", resp.Items[1].ItemID)
	}
	pinnedReasons := reasonSink["c"]
	if !containsReason(pinnedReasons, "rule.pin["+pinID.String()+"]") {
		t.Fatalf("expected pin reason for c, got %v", pinnedReasons)
	}
	boostedReasons := reasonSink["b"]
	boostToken := "rule.boost:+1.00[" + boostID.String() + "]"
	if !containsReason(boostedReasons, boostToken) {
		t.Fatalf("expected boost reason %s for b, got %v", boostToken, boostedReasons)
	}
}

func containsReason(reasons []string, needle string) bool {
	for _, reason := range reasons {
		if reason == needle {
			return true
		}
	}
	return false
}
