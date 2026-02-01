package statistics

import (
	"math"
	"math/rand"
	"sort"
)

// BootstrapResult stores confidence interval bounds.
type BootstrapResult struct {
	Lower float64
	Upper float64
	Level float64
}

// BootstrapMean estimates a confidence interval for the mean using bootstrap resampling.
func BootstrapMean(samples []float64, iterations int, seed int64, level float64) *BootstrapResult {
	if len(samples) == 0 || iterations <= 0 {
		return nil
	}
	if level <= 0 || level >= 1 {
		level = 0.95
	}

	// #nosec G404 -- deterministic bootstrap resampling
	rng := rand.New(rand.NewSource(seed))
	n := len(samples)
	means := make([]float64, iterations)
	for i := 0; i < iterations; i++ {
		sum := 0.0
		for j := 0; j < n; j++ {
			idx := rng.Intn(n)
			sum += samples[idx]
		}
		means[i] = sum / float64(n)
	}

	sort.Float64s(means)
	alpha := (1 - level) / 2
	lowerIdx := int(math.Floor(alpha * float64(iterations-1)))
	upperIdx := int(math.Ceil((1 - alpha) * float64(iterations-1)))
	if lowerIdx < 0 {
		lowerIdx = 0
	}
	if upperIdx >= iterations {
		upperIdx = iterations - 1
	}

	return &BootstrapResult{Lower: means[lowerIdx], Upper: means[upperIdx], Level: level}
}
