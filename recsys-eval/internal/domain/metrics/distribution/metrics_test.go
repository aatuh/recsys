package distribution

import (
	"math"
	"testing"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"
)

func TestCoverageNoveltyDiversity(t *testing.T) {
	cases := []metrics.EvalCase{
		{Recommended: []string{"a", "b", "c"}},
		{Recommended: []string{"a", "c", "d"}},
	}

	coverage := coverageAtK{k: 2}.ComputeDataset(cases)
	if math.Abs(coverage-0.75) > 1e-6 {
		t.Fatalf("coverage@2 expected 0.75 got %.6f", coverage)
	}

	novelty := noveltyAtK{k: 2}.ComputeDataset(cases)
	if math.Abs(novelty-1.835) > 0.01 {
		t.Fatalf("novelty@2 expected ~1.835 got %.6f", novelty)
	}

	diversity := diversityAtK{k: 2}.ComputeDataset(cases)
	if math.Abs(diversity-0.946) > 0.01 {
		t.Fatalf("diversity@2 expected ~0.946 got %.6f", diversity)
	}
}
