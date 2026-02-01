package math

import "math/rand"

// RNG wraps math/rand for deterministic runs.
type RNG struct {
	r *rand.Rand
}

func New(seed int64) RNG {
	// #nosec G404 -- deterministic RNG for reproducible evaluation runs
	return RNG{r: rand.New(rand.NewSource(seed))}
}

func (r RNG) Float64() float64 {
	return r.r.Float64()
}

func (r RNG) Intn(n int) int {
	return r.r.Intn(n)
}
