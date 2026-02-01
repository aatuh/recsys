package middleware

import (
	"net/http"

	"github.com/aatuh/recsys-suite/api/internal/http/problem"
)

// EnsureRequestIDHeader echoes the request id header when present.
func EnsureRequestIDHeader(next http.Handler) http.Handler {
	if next == nil {
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		problem.SetRequestIDHeader(w, r)
		next.ServeHTTP(w, r)
	})
}
