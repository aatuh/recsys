package statistics

import (
	"errors"
	"math"

	"gonum.org/v1/gonum/stat/distuv"
)

// ChiSquarePValue computes the p-value for a goodness-of-fit test.
func ChiSquarePValue(observed map[string]int, expected map[string]float64) (float64, error) {
	if len(observed) == 0 {
		return 1, nil
	}

	total := 0
	for _, v := range observed {
		if v < 0 {
			return 1, errors.New("observed count must be non-negative")
		}
		total += v
	}
	if total == 0 {
		return 1, nil
	}

	sumExpected := 0.0
	for _, v := range expected {
		if v < 0 {
			return 1, errors.New("expected ratios must be non-negative")
		}
		sumExpected += v
	}
	if sumExpected == 0 {
		return 1, errors.New("expected ratios sum to zero")
	}

	chi := 0.0
	k := 0
	for key, obs := range observed {
		ratio, ok := expected[key]
		if !ok {
			return 1, errors.New("expected ratios missing variant")
		}
		exp := ratio / sumExpected * float64(total)
		if exp <= 0 {
			continue
		}
		diff := float64(obs) - exp
		chi += diff * diff / exp
		k++
	}
	if k <= 1 {
		return 1, nil
	}

	df := float64(k - 1)
	p := distuv.ChiSquared{K: df}.Survival(chi)
	if math.IsNaN(p) || math.IsInf(p, 0) {
		return 1, nil
	}
	if p < 0 {
		return 0, nil
	}
	if p > 1 {
		return 1, nil
	}
	return p, nil
}
