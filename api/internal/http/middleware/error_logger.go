package middleware

import (
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// ErrorLogger logs 5xx responses with request metadata and correlation id.
func ErrorLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(rec, r)

			if rec.status >= 500 {
				logger.Error("http 5xx",
					zap.Int("status", rec.status),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("req_id", chimw.GetReqID(r.Context())),
					zap.Duration("latency_ms", time.Since(start)),
				)
			}
		})
	}
}
