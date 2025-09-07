package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// requestStatusRecorder wraps ResponseWriter to capture status code.
type requestStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *requestStatusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func extractClientIP(r *http.Request) string {
	// Prefer X-Real-IP
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	// Fallback to first X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	// Finally, remote addr (strip port if present)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

// RequestLogger logs request/response metadata with request id and client ip.
func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &requestStatusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(rec, r)

			clientIP := extractClientIP(r)
			reqID := chimw.GetReqID(r.Context())
			logger.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rec.status),
				zap.Duration("latency_ms", time.Since(start)),
				zap.String("req_id", reqID),
				zap.String("client_ip", clientIP),
			)
		})
	}
}
