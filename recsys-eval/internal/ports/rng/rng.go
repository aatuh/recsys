package rng

// RNG provides randomness for deterministic runs.
type RNG interface {
	Float64() float64
	Intn(n int) int
}
