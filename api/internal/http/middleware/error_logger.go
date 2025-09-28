package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type statusRecorder struct {
	http.ResponseWriter
	status      int
	body        *bytes.Buffer
	contentType string
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.contentType = r.Header().Get("Content-Type")
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.body == nil {
		r.body = &bytes.Buffer{}
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// ErrorLogger logs all error responses (4xx/5xx) with comprehensive context.
func ErrorLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(rec, r)

			// Log all error responses (4xx and 5xx)
			if rec.status >= 400 {
				fields := []zap.Field{
					zap.Int("status", rec.status),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("query", r.URL.RawQuery),
					zap.String("req_id", chimw.GetReqID(r.Context())),
					zap.Duration("latency_ms", time.Since(start)),
					zap.String("user_agent", r.UserAgent()),
					zap.String("remote_addr", r.RemoteAddr),
				}

				// Add request headers for debugging
				if orgID := r.Header.Get("X-Org-ID"); orgID != "" {
					fields = append(fields, zap.String("org_id", orgID))
				}
				if contentType := r.Header.Get("Content-Type"); contentType != "" {
					fields = append(fields, zap.String("content_type", contentType))
				}

				// Add response body for error context (if JSON)
				if rec.body != nil && rec.body.Len() > 0 {
					if strings.Contains(rec.contentType, "application/json") {
						var errorResp map[string]interface{}
						if err := json.Unmarshal(rec.body.Bytes(), &errorResp); err == nil {
							fields = append(fields, zap.Any("error_response", errorResp))
						}
					}
				}

				// Add request body for debugging (limited size)
				if r.Body != nil && r.ContentLength > 0 && r.ContentLength < 1024 {
					bodyBytes, _ := io.ReadAll(r.Body)
					if len(bodyBytes) > 0 {
						var reqBody interface{}
						if err := json.Unmarshal(bodyBytes, &reqBody); err == nil {
							fields = append(fields, zap.Any("request_body", reqBody))
						}
						// Restore body for potential retry
						r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
					}
				}

				// Use appropriate log level
				if rec.status >= 500 {
					logger.Error("http server error", fields...)
				} else {
					// Use Info level for 4xx errors to avoid stack traces in development mode
					logger.Info("http client error", fields...)
				}
			}
		})
	}
}
