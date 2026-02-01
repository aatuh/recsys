package ranking

import (
	"fmt"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"
)

type hitRateAtK struct{ k int }

func (m hitRateAtK) Name() string { return fmt.Sprintf("hitrate@%d", m.k) }

func (m hitRateAtK) Compute(c metrics.EvalCase) float64 {
	c = makeCase(c)
	items := topK(c.Recommended, m.k)
	if len(items) == 0 {
		return 0
	}
	for _, item := range items {
		if _, ok := c.Relevant[item]; ok {
			return 1
		}
	}
	return 0
}

func NewHitRateAtK(spec metrics.MetricSpec) (metrics.Metric, error) {
	k := spec.K
	if k <= 0 {
		return nil, fmt.Errorf("hitrate@K requires k > 0")
	}
	return hitRateAtK{k: k}, nil
}
