package ope

import (
	"math"
	"sort"
)

// Sample represents a single logged action.
type Sample struct {
	Reward          float64
	LoggingProp     float64
	TargetProp      float64
	PositionWeight  float64
	ModelPrediction float64
}

// EstimatorResult captures estimator outputs.
type EstimatorResult struct {
	Value    float64
	Variance float64
}

// IPS computes the inverse propensity score estimate.
func IPS(samples []Sample, clip float64) EstimatorResult {
	if len(samples) == 0 {
		return EstimatorResult{}
	}
	values := make([]float64, len(samples))
	sum := 0.0
	for i, s := range samples {
		w := weight(s, clip)
		val := w * s.Reward * s.PositionWeight
		values[i] = val
		sum += val
	}
	mean := sum / float64(len(samples))
	return EstimatorResult{Value: mean, Variance: variance(values, mean)}
}

// SNIPS computes the self-normalized IPS estimate.
func SNIPS(samples []Sample, clip float64) EstimatorResult {
	if len(samples) == 0 {
		return EstimatorResult{}
	}
	weighted := make([]float64, len(samples))
	weights := make([]float64, len(samples))
	sum := 0.0
	denom := 0.0
	for i, s := range samples {
		w := weight(s, clip)
		weights[i] = w
		val := w * s.Reward * s.PositionWeight
		weighted[i] = val
		sum += val
		denom += w
	}
	if denom == 0 {
		return EstimatorResult{}
	}
	mean := sum / denom
	return EstimatorResult{Value: mean, Variance: variance(weighted, mean)}
}

// DR computes the doubly-robust estimate.
func DR(samples []Sample, clip float64) EstimatorResult {
	if len(samples) == 0 {
		return EstimatorResult{}
	}
	values := make([]float64, len(samples))
	sum := 0.0
	for i, s := range samples {
		w := weight(s, clip)
		val := s.ModelPrediction*s.PositionWeight + w*(s.Reward-s.ModelPrediction)*s.PositionWeight
		values[i] = val
		sum += val
	}
	mean := sum / float64(len(samples))
	return EstimatorResult{Value: mean, Variance: variance(values, mean)}
}

func weight(s Sample, clip float64) float64 {
	if s.LoggingProp <= 0 {
		return 0
	}
	w := s.TargetProp / s.LoggingProp
	if clip > 0 && w > clip {
		return clip
	}
	return w
}

func variance(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	varSum := 0.0
	for _, v := range values {
		d := v - mean
		varSum += d * d
	}
	return varSum / float64(len(values)-1)
}

// EffectiveSampleSize computes the ESS for weights.
func EffectiveSampleSize(weights []float64) float64 {
	if len(weights) == 0 {
		return 0
	}
	sum := 0.0
	sumSq := 0.0
	for _, w := range weights {
		sum += w
		sumSq += w * w
	}
	if sumSq == 0 {
		return 0
	}
	return (sum * sum) / sumSq
}

// Percentile computes the pth percentile for sorted or unsorted values.
func Percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	clamped := p
	if clamped < 0 {
		clamped = 0
	}
	if clamped > 1 {
		clamped = 1
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	idx := int(math.Round(float64(len(sorted)-1) * clamped))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
