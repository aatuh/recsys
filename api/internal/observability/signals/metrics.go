package signals

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/aatuh/recsys-algo/algorithm"
)

// Metrics exposes counters for signal availability outcomes.
type Metrics struct {
	outcomes *prometheus.CounterVec
}

// NewMetrics registers signal metrics with the provided Prometheus registerer.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	if reg == nil {
		return nil
	}

	outcomes := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_signal_outcomes_total",
		Help: "Outcomes for recommendation signal retrievals.",
	}, []string{"signal", "outcome"})

	reg.MustRegister(outcomes)
	return &Metrics{outcomes: outcomes}
}

// RecordSignal records the outcome of a signal retrieval attempt.
func (m *Metrics) RecordSignal(signal algorithm.Signal, outcome algorithm.SignalOutcome) {
	if m == nil || m.outcomes == nil {
		return
	}
	m.outcomes.WithLabelValues(string(signal), strings.ToLower(string(outcome))).Inc()
}
