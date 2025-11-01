package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrorMetrics tracks error statistics
type ErrorMetrics struct {
	mu sync.RWMutex

	// Counters
	TotalRequests int64
	Error4xx      int64
	Error5xx      int64
	ErrorByCode   map[int]int64
	ErrorByPath   map[string]int64

	// Timing
	ErrorLatency      time.Duration
	TotalErrorLatency time.Duration

	// Last reset time
	LastReset time.Time
}

// NewErrorMetrics creates a new error metrics tracker
func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		ErrorByCode: make(map[int]int64),
		ErrorByPath: make(map[string]int64),
		LastReset:   time.Now(),
	}
}

// RecordError records an error occurrence
func (em *ErrorMetrics) RecordError(status int, path string, latency time.Duration) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.TotalRequests++
	em.TotalErrorLatency += latency

	if status >= 400 && status < 500 {
		em.Error4xx++
	} else if status >= 500 {
		em.Error5xx++
	}

	em.ErrorByCode[status]++
	em.ErrorByPath[path]++
}

// GetStats returns current error statistics
func (em *ErrorMetrics) GetStats() map[string]interface{} {
	em.mu.RLock()
	defer em.mu.RUnlock()

	stats := map[string]interface{}{
		"total_requests": em.TotalRequests,
		"error_4xx":      em.Error4xx,
		"error_5xx":      em.Error5xx,
		"error_by_code":  em.ErrorByCode,
		"error_by_path":  em.ErrorByPath,
		"last_reset":     em.LastReset,
	}

	if em.TotalRequests > 0 {
		stats["error_rate_4xx"] = float64(em.Error4xx) / float64(em.TotalRequests)
		stats["error_rate_5xx"] = float64(em.Error5xx) / float64(em.TotalRequests)
		stats["avg_error_latency_ms"] = em.TotalErrorLatency.Milliseconds() / em.TotalRequests
	}

	return stats
}

// Reset clears all metrics
func (em *ErrorMetrics) Reset() {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.TotalRequests = 0
	em.Error4xx = 0
	em.Error5xx = 0
	em.ErrorByCode = make(map[int]int64)
	em.ErrorByPath = make(map[string]int64)
	em.TotalErrorLatency = 0
	em.LastReset = time.Now()
}

// ErrorMetricsMiddleware tracks error metrics
func ErrorMetricsMiddleware(metrics *ErrorMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &metricsRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			// Record error metrics for 4xx and 5xx responses
			if rec.status >= http.StatusBadRequest {
				metrics.RecordError(rec.status, r.URL.Path, time.Since(start))
			}
		})
	}
}

// LogErrorMetrics logs error metrics periodically until the context is done.
func LogErrorMetrics(ctx context.Context, logger *zap.Logger, metrics *ErrorMetrics, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := metrics.GetStats()
			logger.Info("error metrics", zap.Any("stats", stats))
		}
	}
}

type metricsRecorder struct {
	http.ResponseWriter
	status int
}

func (r *metricsRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
