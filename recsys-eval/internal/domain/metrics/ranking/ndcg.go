package ranking

import (
	"fmt"
	"math"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"
)

type ndcgAtK struct{ k int }

func (m ndcgAtK) Name() string { return fmt.Sprintf("ndcg@%d", m.k) }

func (m ndcgAtK) Compute(c metrics.EvalCase) float64 {
	c = makeCase(c)
	items := topK(c.Recommended, m.k)
	if len(items) == 0 {
		return 0
	}

	dcg := 0.0
	for i, item := range items {
		if _, ok := c.Relevant[item]; ok {
			dcg += 1.0 / math.Log2(float64(i+2))
		}
	}

	rel := totalRelevant(c.Relevant)
	if rel == 0 {
		return 0
	}
	ideal := 0.0
	limit := rel
	if limit > m.k {
		limit = m.k
	}
	for i := 0; i < limit; i++ {
		ideal += 1.0 / math.Log2(float64(i+2))
	}
	if ideal == 0 {
		return 0
	}
	return dcg / ideal
}

func NewNDCGAtK(spec metrics.MetricSpec) (metrics.Metric, error) {
	k := spec.K
	if k <= 0 {
		return nil, fmt.Errorf("ndcg@K requires k > 0")
	}
	return ndcgAtK{k: k}, nil
}
