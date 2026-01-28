package main

import (
	"context"
	"fmt"

	"github.com/aatuh/recsys-algo/algorithm"
	recmodel "github.com/aatuh/recsys-algo/model"

	"github.com/google/uuid"
)

type personalizedStore struct{}

func (personalizedStore) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int, c *recmodel.PopConstraints) ([]recmodel.ScoredItem, error) {
	items := []recmodel.ScoredItem{
		{ItemID: "green_tea", Score: 10},
		{ItemID: "black_tea", Score: 9},
		{ItemID: "coffee", Score: 8},
	}
	if k > 0 && k < len(items) {
		return items[:k], nil
	}
	return items, nil
}

func (personalizedStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]recmodel.ItemTags, error) {
	tags := map[string]recmodel.ItemTags{
		"green_tea": {ItemID: "green_tea", Tags: []string{"tea", "drink"}},
		"black_tea": {ItemID: "black_tea", Tags: []string{"tea", "drink"}},
		"coffee":    {ItemID: "coffee", Tags: []string{"coffee", "drink"}},
	}
	out := make(map[string]recmodel.ItemTags, len(itemIDs))
	for _, id := range itemIDs {
		if info, ok := tags[id]; ok {
			out[id] = info
		}
	}
	return out, nil
}

func (personalizedStore) BuildUserTagProfile(ctx context.Context, orgID uuid.UUID, ns string, userID string, windowDays float64, topN int) (map[string]float64, error) {
	_ = orgID
	_ = ns
	_ = userID
	_ = windowDays
	_ = topN
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return map[string]float64{"tea": 1.0}, nil
}

func main() {
	engine := algorithm.NewEngine(algorithm.Config{
		BlendAlpha:        1,
		ProfileBoost:      0.6,
		ProfileWindowDays: 30,
		ProfileTopNTags:   5,
	}, personalizedStore{}, nil)

	resp, _, err := engine.Recommend(context.Background(), algorithm.Request{
		OrgID:          uuid.New(),
		Namespace:      "default",
		UserID:         "user-123",
		K:              3,
		IncludeReasons: true,
		ExplainLevel:   algorithm.ExplainLevelNumeric,
	})
	if err != nil {
		panic(err)
	}
	for _, item := range resp.Items {
		fmt.Printf("%s %.2f %v\n", item.ItemID, item.Score, item.Reasons)
	}
}
