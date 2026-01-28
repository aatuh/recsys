package algorithm

import (
	"math"
	"reflect"
	"strings"
	"testing"

	recmodel "github.com/aatuh/recsys-algo/model"
)

// Tags:
// - brand:acme / brand:bravo
// - category:phone / category:laptop
// - content tags: t:android, t:ios, t:gaming, t:work
func metaFixture() map[string]recmodel.ItemTags {
	return map[string]recmodel.ItemTags{
		"A": {ItemID: "A", Tags: []string{
			"brand:acme", "category:phone", "t:android", "t:gaming",
		}},
		"B": {ItemID: "B", Tags: []string{
			"brand:acme", "category:phone", "t:android", "t:work",
		}},
		"C": {ItemID: "C", Tags: []string{
			"brand:acme", "category:laptop", "t:work",
		}},
		"D": {ItemID: "D", Tags: []string{
			"brand:bravo", "category:laptop", "t:work",
		}},
		"E": {ItemID: "E", Tags: []string{
			"brand:bravo", "category:phone", "t:ios",
		}},
		"F": {ItemID: "F", Tags: []string{
			"brand:bravo", "category:phone", "t:ios", "t:gaming",
		}},
	}
}

func candidatesFixture() []recmodel.ScoredItem {
	// Higher score is better. Order here is the "pre" order.
	return []recmodel.ScoredItem{
		{ItemID: "A", Score: 1.00},
		{ItemID: "B", Score: 0.95},
		{ItemID: "C", Score: 0.90},
		{ItemID: "D", Score: 0.80},
		{ItemID: "E", Score: 0.70},
		{ItemID: "F", Score: 0.60},
	}
}

func ids(xs []recmodel.ScoredItem) []string {
	out := make([]string, 0, len(xs))
	for _, it := range xs {
		out = append(out, it.ItemID)
	}
	return out
}

func TestMMR_Lambda1KeepsRelevanceOrder(t *testing.T) {
	meta := metaFixture()
	cands := candidatesFixture()

	got := MMRReRank(cands, meta, 5, 1.0, 0, 0)
	want := []string{"A", "B", "C", "D", "E"}
	if !reflect.DeepEqual(ids(got), want) {
		t.Fatalf("lambda=1.0 should keep order.\n got=%v\nwant=%v",
			ids(got), want)
	}
}

func TestMMR_Lambda0SpreadsDiversity(t *testing.T) {
	meta := metaFixture()
	cands := candidatesFixture()

	got := MMRReRank(cands, meta, 4, 0.0, 0, 0)
	// With lambda=0, objective is pure diversity. We expect it to pick
	// A (first), then a different tag profile before B.
	if len(got) < 2 {
		t.Fatalf("need at least 2 picks, got=%v", ids(got))
	}
	if got[0].ItemID != "A" {
		t.Fatalf("first pick stable by order; got %s", got[0].ItemID)
	}
	// Accept either C, D, or E/F as second (all diversify vs A).
	if got[1].ItemID == "B" {
		t.Fatalf("expected diverse second pick, got %v", ids(got))
	}
}

func TestCaps_Enforced(t *testing.T) {
	meta := metaFixture()
	cands := candidatesFixture()

	// brandCap=1, categoryCap=1 forces mix.
	got := MMRReRank(cands, meta, 6, 0.6, 1, 1)

	brandCount := map[string]int{}
	catCount := map[string]int{}
	for _, it := range got {
		m := meta[it.ItemID]
		for _, tg := range m.Tags {
			if len(tg) >= 6 && tg[:6] == "brand:" {
				brandCount[tg]++
			}
			if len(tg) >= 9 && tg[:9] == "category:" {
				catCount[tg]++
			}
		}
	}
	for b, n := range brandCount {
		if n > 1 {
			t.Fatalf("brand cap violated: %s=%d", b, n)
		}
	}
	for c, n := range catCount {
		if n > 1 {
			t.Fatalf("category cap violated: %s=%d", c, n)
		}
	}
}

func TestCaps_CustomPrefixes(t *testing.T) {
	meta := map[string]recmodel.ItemTags{
		"X": {ItemID: "X", Tags: []string{"maker:acme", "genre:rpg"}},
		"Y": {ItemID: "Y", Tags: []string{"maker:acme", "genre:action"}},
		"Z": {ItemID: "Z", Tags: []string{"maker:bravo", "genre:rpg"}},
	}
	cands := []recmodel.ScoredItem{
		{ItemID: "X", Score: 1.0},
		{ItemID: "Y", Score: 0.9},
		{ItemID: "Z", Score: 0.8},
	}

	got, _, _ := mmrReRankInternal(
		cands,
		meta,
		3,
		0.5,
		1,
		0,
		[]string{"maker"},
		defaultCategoryTagPrefixes,
	)
	acmeCount := 0
	for _, it := range got {
		for _, tag := range meta[it.ItemID].Tags {
			if strings.HasPrefix(tag, "maker:acme") {
				acmeCount++
			}
		}
	}
	if acmeCount > 1 {
		t.Fatalf("maker prefix cap not enforced: got %v", ids(got))
	}
}

func TestJaccard_Basics(t *testing.T) {
	a := map[string]struct{}{"x": {}, "y": {}}
	b := map[string]struct{}{"y": {}, "z": {}}
	s := jaccard(a, b)
	want := 1.0 / 3.0
	eps := 1e-9
	if math.Abs(s-want) > eps {
		t.Fatalf("jaccard wrong: got=%f want≈%f", s, want)
	}
}
