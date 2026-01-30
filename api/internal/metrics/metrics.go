package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	cacheRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "recsys",
		Name:      "cache_requests_total",
		Help:      "Cache lookups by cache name and result.",
	}, []string{"cache", "result"})
	backpressureRejections = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "recsys",
		Name:      "backpressure_rejections_total",
		Help:      "Requests rejected by backpressure limits.",
	})
)

// RecordCacheResult tracks cache hits and misses.
func RecordCacheResult(cacheName string, hit bool) {
	if cacheName == "" {
		cacheName = "unknown"
	}
	result := "miss"
	if hit {
		result = "hit"
	}
	cacheRequests.WithLabelValues(cacheName, result).Inc()
}

// RecordBackpressureRejection increments the overload rejection counter.
func RecordBackpressureRejection() {
	backpressureRejections.Inc()
}
