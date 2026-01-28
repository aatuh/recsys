package middleware

import (
	"net/http"
	"strings"

	"recsys/internal/auth"
	"recsys/internal/config"
	"recsys/internal/http/problem"
)

// AdminRoleMiddleware enforces admin role on control-plane routes.
type AdminRoleMiddleware struct {
	cfg config.AuthConfig
}

// NewAdminRoleMiddleware constructs the middleware.
func NewAdminRoleMiddleware(cfg config.AuthConfig) *AdminRoleMiddleware {
	return &AdminRoleMiddleware{cfg: cfg}
}

// Handler wraps the next handler with admin role checks.
func (m *AdminRoleMiddleware) Handler(next http.Handler) http.Handler {
	if m == nil {
		return next
	}
	if next == nil {
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}
	required := strings.TrimSpace(m.cfg.AdminRole)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAdminPath(r) {
			next.ServeHTTP(w, r)
			return
		}
		if required == "" {
			next.ServeHTTP(w, r)
			return
		}
		roles := auth.RolesFromContext(r.Context())
		for _, role := range roles {
			if strings.EqualFold(strings.TrimSpace(role), required) {
				next.ServeHTTP(w, r)
				return
			}
		}
		problem.Write(w, r, http.StatusForbidden, "RECSYS_FORBIDDEN", "insufficient scope")
	})
}
