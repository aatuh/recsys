package middleware

import (
	"net/http"

	"github.com/aatuh/recsys-suite/api/internal/auth"
	"github.com/aatuh/recsys-suite/api/internal/config"
	"github.com/aatuh/recsys-suite/api/internal/http/problem"
)

// RequireAuth enforces that some authentication context is present.
type RequireAuth struct {
	required bool
}

// NewRequireAuth constructs the middleware.
func NewRequireAuth(cfg config.AuthConfig) *RequireAuth {
	return &RequireAuth{required: cfg.Required}
}

// Handler wraps the next handler with auth enforcement.
func (m *RequireAuth) Handler(next http.Handler) http.Handler {
	if m == nil {
		return next
	}
	if next == nil {
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.required || !isProtectedPath(r) {
			next.ServeHTTP(w, r)
			return
		}
		info, ok := auth.FromContext(r.Context())
		if !ok || !hasAuthContext(info) {
			problem.Write(w, r, http.StatusUnauthorized, "RECSYS_AUTH_REQUIRED", "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func hasAuthContext(info auth.Info) bool {
	if info.UserID != "" {
		return true
	}
	if info.APIKeyID != "" {
		return true
	}
	if info.TenantID != "" {
		return true
	}
	if len(info.Roles) > 0 {
		return true
	}
	return false
}
