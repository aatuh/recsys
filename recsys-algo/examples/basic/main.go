package main

import (
	"context"
	"fmt"

	"github.com/aatuh/recsys-suite/api/recsys-algo/algorithm"
	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"

	"github.com/google/uuid"
)

type basicStore struct{}

func (basicStore) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int, c *recmodel.PopConstraints) ([]recmodel.ScoredItem, error) {
	items := []recmodel.ScoredItem{
		{ItemID: "a", Score: 10},
		{ItemID: "b", Score: 8},
		{ItemID: "c", Score: 6},
	}
	if k > 0 && k < len(items) {
		return items[:k], nil
	}
	return items, nil
}

func (basicStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]recmodel.ItemTags, error) {
	tags := map[string]recmodel.ItemTags{
		"a": {ItemID: "a", Tags: []string{"category:books"}},
		"b": {ItemID: "b", Tags: []string{"category:movies"}},
		"c": {ItemID: "c", Tags: []string{"category:games"}},
	}
	out := make(map[string]recmodel.ItemTags, len(itemIDs))
	for _, id := range itemIDs {
		if info, ok := tags[id]; ok {
			out[id] = info
		}
	}
	return out, nil
}

func main() {
	engine := algorithm.NewEngine(algorithm.Config{BlendAlpha: 1}, basicStore{}, nil)
	resp, _, err := engine.Recommend(context.Background(), algorithm.Request{
		OrgID:     uuid.New(),
		Namespace: "default",
		K:         3,
	})
	if err != nil {
		panic(err)
	}
	for _, item := range resp.Items {
		fmt.Printf("%s %.2f\n", item.ItemID, item.Score)
	}
}
