package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"recsys/internal/http/common"

	"go.uber.org/zap"
)

// JSONRecoverer converts panics into a JSON error response using common.HttpError.
func JSONRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// include stack trace for better diagnostics (surfaced only in debug env)
				err := fmt.Errorf("panic: %v\n%s", rec, debug.Stack())
				common.HttpError(
					w,
					r,
					err,
					http.StatusInternalServerError,
				)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// JSONRecovererWithLogger recovers from panics, logs details, and returns JSON error.
func JSONRecovererWithLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := string(debug.Stack())
					logger.Error("panic recovered",
						zap.Any("panic", rec),
						zap.String("stack", stack),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)
					common.HttpError(
						w,
						r,
						fmt.Errorf("panic: %v\n%s", rec, stack),
						http.StatusInternalServerError,
					)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
