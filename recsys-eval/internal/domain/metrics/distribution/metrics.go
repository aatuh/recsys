package distribution

import (
	"fmt"
	"math"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"
)

// RegisterDefaults registers distribution metrics.
func RegisterDefaults(reg *metrics.Registry) {
	reg.Register("coverage", NewCoverageAtK)
	reg.Register("coverage@k", NewCoverageAtK)
	reg.Register("novelty", NewNoveltyAtK)
	reg.Register("novelty@k", NewNoveltyAtK)
	reg.Register("diversity", NewDiversityAtK)
	reg.Register("diversity@k", NewDiversityAtK)
}

type coverageAtK struct{ k int }

func (m coverageAtK) Name() string { return fmt.Sprintf("coverage@%d", m.k) }
func (m coverageAtK) Compute(metrics.EvalCase) float64 {
	return 0
}
func (m coverageAtK) ComputeDataset(cases []metrics.EvalCase) float64 {
	if m.k <= 0 {
		return 0
	}
	catalog := map[string]struct{}{}
	covered := map[string]struct{}{}
	for _, c := range cases {
		for _, id := range c.Recommended {
			if id != "" {
				catalog[id] = struct{}{}
			}
		}
		for _, id := range topK(c.Recommended, m.k) {
			if id != "" {
				covered[id] = struct{}{}
			}
		}
		for id := range c.Relevant {
			if id != "" {
				catalog[id] = struct{}{}
			}
		}
	}
	if len(catalog) == 0 {
		return 0
	}
	return float64(len(covered)) / float64(len(catalog))
}

type noveltyAtK struct{ k int }

func (m noveltyAtK) Name() string { return fmt.Sprintf("novelty@%d", m.k) }
func (m noveltyAtK) Compute(metrics.EvalCase) float64 {
	return 0
}
func (m noveltyAtK) ComputeDataset(cases []metrics.EvalCase) float64 {
	if m.k <= 0 {
		return 0
	}
	popularity := map[string]int{}
	total := 0
	for _, c := range cases {
		for _, id := range c.Recommended {
			if id == "" {
				continue
			}
			popularity[id]++
			total++
		}
	}
	if total == 0 {
		return 0
	}
	sum := 0.0
	count := 0
	for _, c := range cases {
		for _, id := range topK(c.Recommended, m.k) {
			pop := popularity[id]
			if pop == 0 {
				continue
			}
			p := float64(pop) / float64(total)
			sum += -math.Log2(p)
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

type diversityAtK struct{ k int }

func (m diversityAtK) Name() string { return fmt.Sprintf("diversity@%d", m.k) }
func (m diversityAtK) Compute(metrics.EvalCase) float64 {
	return 0
}
func (m diversityAtK) ComputeDataset(cases []metrics.EvalCase) float64 {
	if m.k <= 0 {
		return 0
	}
	counts := map[string]int{}
	total := 0
	for _, c := range cases {
		for _, id := range topK(c.Recommended, m.k) {
			if id == "" {
				continue
			}
			counts[id]++
			total++
		}
	}
	if total == 0 {
		return 0
	}
	if len(counts) <= 1 {
		return 0
	}
	entropy := 0.0
	for _, count := range counts {
		p := float64(count) / float64(total)
		if p > 0 {
			entropy += -p * math.Log2(p)
		}
	}
	maxEntropy := math.Log2(float64(len(counts)))
	if maxEntropy == 0 {
		return 0
	}
	return entropy / maxEntropy
}

func NewCoverageAtK(spec metrics.MetricSpec) (metrics.Metric, error) {
	if spec.K <= 0 {
		return nil, fmt.Errorf("coverage@K requires k > 0")
	}
	return coverageAtK{k: spec.K}, nil
}

func NewNoveltyAtK(spec metrics.MetricSpec) (metrics.Metric, error) {
	if spec.K <= 0 {
		return nil, fmt.Errorf("novelty@K requires k > 0")
	}
	return noveltyAtK{k: spec.K}, nil
}

func NewDiversityAtK(spec metrics.MetricSpec) (metrics.Metric, error) {
	if spec.K <= 0 {
		return nil, fmt.Errorf("diversity@K requires k > 0")
	}
	return diversityAtK{k: spec.K}, nil
}

func topK(items []string, k int) []string {
	if k <= 0 || len(items) <= k {
		return items
	}
	return items[:k]
}
