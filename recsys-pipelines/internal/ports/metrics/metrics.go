package metrics

import "time"

type Metrics interface {
	IncCounter(name string, delta int64, labels map[string]string)
	ObserveDuration(name string, d time.Duration, labels map[string]string)
}
