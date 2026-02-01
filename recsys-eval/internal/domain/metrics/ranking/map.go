package ranking

import (
	"fmt"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"
)

type mapAtK struct{ k int }

func (m mapAtK) Name() string { return fmt.Sprintf("map@%d", m.k) }

func (m mapAtK) Compute(c metrics.EvalCase) float64 {
	c = makeCase(c)
	items := topK(c.Recommended, m.k)
	if len(items) == 0 {
		return 0
	}
	relevantTotal := totalRelevant(c.Relevant)
	if relevantTotal == 0 {
		return 0
	}

	sumPrec := 0.0
	relSoFar := 0
	for i, item := range items {
		if _, ok := c.Relevant[item]; ok {
			relSoFar++
			prec := float64(relSoFar) / float64(i+1)
			sumPrec += prec
		}
	}
	denom := relevantTotal
	if denom > m.k {
		denom = m.k
	}
	return sumPrec / float64(denom)
}

func NewMAPAtK(spec metrics.MetricSpec) (metrics.Metric, error) {
	k := spec.K
	if k <= 0 {
		return nil, fmt.Errorf("map@K requires k > 0")
	}
	return mapAtK{k: k}, nil
}
