package algorithm

import (
	"testing"

	recmodel "github.com/aatuh/recsys-algo/model"

	"github.com/stretchr/testify/require"
)

func TestMergeCandidatesRetainsFanout(t *testing.T) {
	cfg := Config{PopularityFanout: 5}
	eng := NewEngine(cfg, nil, nil)

	pop := []recmodel.ScoredItem{
		{ItemID: "a", Score: 5},
		{ItemID: "b", Score: 4},
		{ItemID: "c", Score: 3},
		{ItemID: "d", Score: 2},
		{ItemID: "e", Score: 1},
	}

	merged, _ := eng.mergeCandidates(pop, map[string]float64{}, map[string]float64{}, map[string]float64{}, 3)
	require.Len(t, merged, len(pop), "merge should retain popularity fanout even when k is smaller")

	ids := make(map[string]struct{}, len(merged))
	for _, cand := range merged {
		ids[cand.ItemID] = struct{}{}
	}
	for _, cand := range pop {
		if _, ok := ids[cand.ItemID]; !ok {
			t.Fatalf("expected candidate %s to be retained", cand.ItemID)
		}
	}
}

func TestMergeCandidatesIncludesOtherSources(t *testing.T) {
	cfg := Config{PopularityFanout: 8}
	eng := NewEngine(cfg, nil, nil)

	pop := []recmodel.ScoredItem{
		{ItemID: "a", Score: 8},
		{ItemID: "b", Score: 7},
		{ItemID: "c", Score: 6},
		{ItemID: "d", Score: 5},
		{ItemID: "e", Score: 4},
		{ItemID: "f", Score: 3},
		{ItemID: "g", Score: 2},
		{ItemID: "h", Score: 1},
	}
	collab := map[string]float64{"x": 10, "y": 9}

	merged, sources := eng.mergeCandidates(pop, collab, nil, nil, 4)

	hasCollab := false
	popCount := 0
	popSet := make(map[string]struct{}, len(pop))
	for _, cand := range pop {
		popSet[cand.ItemID] = struct{}{}
	}
	for _, cand := range merged {
		if _, ok := popSet[cand.ItemID]; ok {
			popCount++
		}
		if cand.ItemID == "x" || cand.ItemID == "y" {
			hasCollab = true
			if set := sources[cand.ItemID]; set != nil {
				if _, ok := set[SignalCollaborative]; ok {
					continue
				}
				t.Fatalf("expected collaborative source for %s", cand.ItemID)
			}
			t.Fatalf("expected source set for %s", cand.ItemID)
		}
	}

	reserveSlots := len(pop) / 4
	if reserveSlots < 1 {
		reserveSlots = 1
	}
	if reserveSlots > minInt(20, len(pop)) {
		reserveSlots = minInt(20, len(pop))
	}

	if !hasCollab {
		t.Fatalf("expected collaborative candidates to be included")
	}
	if popCount < len(pop)-reserveSlots {
		t.Fatalf("expected at least %d pop candidates, got %d", len(pop)-reserveSlots, popCount)
	}
}
