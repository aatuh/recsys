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
	recommendationRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "recsys",
		Name:      "recommendation_requests_total",
		Help:      "Recommendation API requests by endpoint, outcome, and whether the successful response was empty.",
	}, []string{"endpoint", "outcome", "empty"})
	recommendationLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "recsys",
		Name:      "recommendation_latency_seconds",
		Help:      "Recommendation API latency by endpoint and outcome.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"endpoint", "outcome"})
	recommendationReturnedItems = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "recsys",
		Name:      "recommendation_returned_items",
		Help:      "Returned recommendation item count by endpoint.",
		Buckets:   []float64{0, 1, 5, 10, 20, 50, 100, 200},
	}, []string{"endpoint"})
	recommendationWarnings = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "recsys",
		Name:      "recommendation_warnings",
		Help:      "Non-fatal warning count by endpoint.",
		Buckets:   []float64{0, 1, 2, 5, 10, 20},
	}, []string{"endpoint"})
	artifactLoadFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "recsys",
		Name:      "artifact_load_failures_total",
		Help:      "Artifact or manifest load failures by artifact kind.",
	}, []string{"kind"})
	artifactManifestAge = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "recsys",
		Name:      "artifact_manifest_age_seconds",
		Help:      "Age in seconds of the most recently loaded artifact manifest.",
	})
	artifactManifestArtifacts = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "recsys",
		Name:      "artifact_manifest_artifacts",
		Help:      "Artifact pointer count in the most recently loaded artifact manifest.",
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

// RecordRecommendation tracks recommendation endpoint outcomes.
func RecordRecommendation(endpoint string, outcome string, returned int, warnings int, durationSeconds float64) {
	if endpoint == "" {
		endpoint = "unknown"
	}
	if outcome == "" {
		outcome = "unknown"
	}
	empty := "false"
	if outcome == "success" && returned == 0 {
		empty = "true"
	}
	recommendationRequests.WithLabelValues(endpoint, outcome, empty).Inc()
	recommendationLatency.WithLabelValues(endpoint, outcome).Observe(durationSeconds)
	if outcome == "success" {
		recommendationReturnedItems.WithLabelValues(endpoint).Observe(float64(returned))
		recommendationWarnings.WithLabelValues(endpoint).Observe(float64(warnings))
	}
}

// RecordArtifactLoadFailure tracks artifact loader failures without exposing object-store URIs.
func RecordArtifactLoadFailure(kind string) {
	if kind == "" {
		kind = "unknown"
	}
	artifactLoadFailures.WithLabelValues(kind).Inc()
}

// RecordArtifactManifest tracks freshness for the latest successfully loaded manifest.
func RecordArtifactManifest(ageSeconds float64, artifactCount int) {
	if ageSeconds >= 0 {
		artifactManifestAge.Set(ageSeconds)
	}
	if artifactCount >= 0 {
		artifactManifestArtifacts.Set(float64(artifactCount))
	}
}
