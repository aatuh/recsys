package ranking

import (
	"fmt"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"
)

type precisionAtK struct{ k int }

type recallAtK struct{ k int }

func (m precisionAtK) Name() string { return fmt.Sprintf("precision@%d", m.k) }
func (m recallAtK) Name() string    { return fmt.Sprintf("recall@%d", m.k) }

func (m precisionAtK) Compute(c metrics.EvalCase) float64 {
	c = makeCase(c)
	items := topK(c.Recommended, m.k)
	if len(items) == 0 {
		return 0
	}
	return float64(relevantCount(items, c.Relevant)) / float64(len(items))
}

func (m recallAtK) Compute(c metrics.EvalCase) float64 {
	c = makeCase(c)
	total := totalRelevant(c.Relevant)
	if total == 0 {
		return 0
	}
	items := topK(c.Recommended, m.k)
	return float64(relevantCount(items, c.Relevant)) / float64(total)
}

func NewPrecisionAtK(spec metrics.MetricSpec) (metrics.Metric, error) {
	k := spec.K
	if k <= 0 {
		return nil, fmt.Errorf("precision@K requires k > 0")
	}
	return precisionAtK{k: k}, nil
}

func NewRecallAtK(spec metrics.MetricSpec) (metrics.Metric, error) {
	k := spec.K
	if k <= 0 {
		return nil, fmt.Errorf("recall@K requires k > 0")
	}
	return recallAtK{k: k}, nil
}
