package ranking

import "github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"

func topK(items []string, k int) []string {
	if k <= 0 || k >= len(items) {
		return items
	}
	return items[:k]
}

func relevantCount(items []string, relevant map[string]struct{}) int {
	count := 0
	for _, item := range items {
		if _, ok := relevant[item]; ok {
			count++
		}
	}
	return count
}

func totalRelevant(relevant map[string]struct{}) int {
	return len(relevant)
}

// makeCase ensures metrics handle empty data safely.
func makeCase(c metrics.EvalCase) metrics.EvalCase {
	if c.Relevant == nil {
		c.Relevant = map[string]struct{}{}
	}
	return c
}
