package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aatuh/recsys-algo/algorithm"
	recmodel "github.com/aatuh/recsys-algo/model"
	"github.com/aatuh/recsys-algo/rules"

	"github.com/google/uuid"
)

type rulesStore struct {
	rules []rules.Rule
}

func (r rulesStore) ListActiveRulesForScope(ctx context.Context, orgID uuid.UUID, namespace, surface, segmentID string, ts time.Time) ([]rules.Rule, error) {
	return append([]rules.Rule(nil), r.rules...), nil
}

type catalogStore struct{}

func (catalogStore) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int, c *recmodel.PopConstraints) ([]recmodel.ScoredItem, error) {
	items := []recmodel.ScoredItem{
		{ItemID: "a", Score: 10},
		{ItemID: "b", Score: 9},
		{ItemID: "c", Score: 8},
	}
	if k > 0 && k < len(items) {
		return items[:k], nil
	}
	return items, nil
}

func (catalogStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]recmodel.ItemTags, error) {
	return map[string]recmodel.ItemTags{}, nil
}

func main() {
	orgID := uuid.New()
	pinRule := rules.Rule{
		RuleID:     uuid.New(),
		OrgID:      orgID,
		Namespace:  "default",
		Surface:    "home",
		Action:     rules.RuleActionPin,
		TargetType: rules.RuleTargetItem,
		ItemIDs:    []string{"b"},
		Priority:   100,
		Enabled:    true,
	}
	boostVal := 0.2
	boostRule := rules.Rule{
		RuleID:     uuid.New(),
		OrgID:      orgID,
		Namespace:  "default",
		Surface:    "home",
		Action:     rules.RuleActionBoost,
		TargetType: rules.RuleTargetItem,
		ItemIDs:    []string{"c"},
		BoostValue: &boostVal,
		Priority:   50,
		Enabled:    true,
	}

	manager := rules.NewManager(rulesStore{rules: []rules.Rule{pinRule, boostRule}}, rules.ManagerOptions{
		Enabled:         true,
		RefreshInterval: time.Minute,
		MaxPinSlots:     2,
	})

	engine := algorithm.NewEngine(algorithm.Config{
		BlendAlpha:   1,
		RulesEnabled: true,
	}, catalogStore{}, manager)

	resp, _, err := engine.Recommend(context.Background(), algorithm.Request{
		OrgID:     orgID,
		Namespace: "default",
		Surface:   "home",
		K:         3,
	})
	if err != nil {
		panic(err)
	}
	for _, item := range resp.Items {
		fmt.Printf("%s %.2f\n", item.ItemID, item.Score)
	}
}
