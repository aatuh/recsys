package algorithm

import (
	"testing"

	"recsys/internal/types"

	"github.com/stretchr/testify/require"
)

func TestMergeCandidatesRetainsFanout(t *testing.T) {
	cfg := Config{PopularityFanout: 5}
	eng := NewEngine(cfg, nil, nil)

	pop := []types.ScoredItem{
		{ItemID: "a", Score: 5},
		{ItemID: "b", Score: 4},
		{ItemID: "c", Score: 3},
		{ItemID: "d", Score: 2},
		{ItemID: "e", Score: 1},
	}

	res := eng.mergeCandidates(pop, map[string]float64{}, nil, nil, nil, 3)
	require.Len(t, res, len(pop), "merge should retain popularity fanout even when k is smaller")

	ids := make(map[string]struct{}, len(res))
	for _, cand := range res {
		ids[cand.ItemID] = struct{}{}
	}
	for _, cand := range pop {
		if _, ok := ids[cand.ItemID]; !ok {
			t.Fatalf("expected candidate %s to be retained", cand.ItemID)
		}
	}
}
