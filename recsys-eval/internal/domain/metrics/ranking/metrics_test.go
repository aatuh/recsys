package ranking

import (
	"math"
	"testing"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"
)

func TestRankingMetrics(t *testing.T) {
	c := metrics.EvalCase{
		Recommended: []string{"i1", "i2", "i3"},
		Relevant: map[string]struct{}{
			"i1": {},
			"i3": {},
		},
	}

	precision, _ := NewPrecisionAtK(metrics.MetricSpec{Name: "precision", K: 2})
	recall, _ := NewRecallAtK(metrics.MetricSpec{Name: "recall", K: 2})
	mapk, _ := NewMAPAtK(metrics.MetricSpec{Name: "map", K: 3})
	ndcg, _ := NewNDCGAtK(metrics.MetricSpec{Name: "ndcg", K: 3})
	hitrate, _ := NewHitRateAtK(metrics.MetricSpec{Name: "hitrate", K: 1})

	if got := precision.Compute(c); got != 0.5 {
		t.Fatalf("precision@2 expected 0.5 got %.4f", got)
	}
	if got := recall.Compute(c); got != 0.5 {
		t.Fatalf("recall@2 expected 0.5 got %.4f", got)
	}
	if got := mapk.Compute(c); math.Abs(got-0.8333) > 0.0005 {
		t.Fatalf("map@3 expected ~0.8333 got %.4f", got)
	}
	if got := ndcg.Compute(c); math.Abs(got-0.9197) > 0.001 {
		t.Fatalf("ndcg@3 expected ~0.9197 got %.4f", got)
	}
	if got := hitrate.Compute(c); got != 1 {
		t.Fatalf("hitrate@1 expected 1 got %.4f", got)
	}
}
