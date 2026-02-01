package noop

import (
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/metrics"
)

type NoopMetrics struct{}

var _ metrics.Metrics = NoopMetrics{}

func (NoopMetrics) IncCounter(_ string, _ int64, _ map[string]string) {}

func (NoopMetrics) ObserveDuration(_ string, _ time.Duration, _ map[string]string) {}
