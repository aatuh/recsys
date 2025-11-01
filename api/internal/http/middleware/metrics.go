package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTPMetrics exposes Prometheus instruments for HTTP handlers.
type HTTPMetrics struct {
	registry    *prometheus.Registry
	requests    *prometheus.CounterVec
	durations   *prometheus.HistogramVec
	inflight    prometheus.Gauge
	metricsPath string
}

// NewHTTPMetrics constructs metrics collectors.
func NewHTTPMetrics() *HTTPMetrics {
	registry := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total count of HTTP requests",
	}, []string{"method", "route", "status"})
	durations := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Latency of HTTP requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
	inflight := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_inflight_requests",
		Help: "Current number of in-flight HTTP requests",
	})

	registry.MustRegister(requests, durations, inflight)

	return &HTTPMetrics{
		registry:    registry,
		requests:    requests,
		durations:   durations,
		inflight:    inflight,
		metricsPath: "/metrics",
	}
}

// Middleware wraps the next handler with metrics instrumentation.
func (m *HTTPMetrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := routePattern(r)
		if route == m.metricsPath {
			next.ServeHTTP(w, r)
			return
		}

		m.inflight.Inc()
		defer m.inflight.Dec()

		start := time.Now()
		rw := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)

		status := strconv.Itoa(rw.status())
		m.requests.WithLabelValues(r.Method, route, status).Inc()
		m.durations.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
	})
}

// Handler exposes the Prometheus metrics handler.
func (m *HTTPMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// SetMetricsPath configures the metrics endpoint to avoid double counting.
func (m *HTTPMetrics) SetMetricsPath(path string) {
	if path == "" {
		return
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	m.metricsPath = path
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) status() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}

func routePattern(r *http.Request) string {
	if r == nil {
		return "unknown"
	}
	if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
		if pattern := routeContext.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	if path := r.URL.Path; path != "" {
		return path
	}
	return "unknown"
}
