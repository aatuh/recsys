package metrics

import (
	"fmt"
	"strings"
)

// EvalCase is a single evaluation case.
type EvalCase struct {
	Recommended []string
	Relevant    map[string]struct{}
}

// Metric computes a score for a case.
type Metric interface {
	Name() string
	Compute(c EvalCase) float64
}

// Registry maps metric names to factories.
type Registry struct {
	factories map[string]func(spec MetricSpec) (Metric, error)
}

func NewRegistry() *Registry {
	return &Registry{factories: map[string]func(spec MetricSpec) (Metric, error){}}
}

func (r *Registry) Register(name string, factory func(spec MetricSpec) (Metric, error)) {
	r.factories[strings.ToLower(name)] = factory
}

func (r *Registry) Build(spec MetricSpec) (Metric, error) {
	factory, ok := r.factories[strings.ToLower(spec.Name)]
	if !ok {
		return nil, fmt.Errorf("unknown metric: %s", spec.Name)
	}
	return factory(spec)
}
