package middleware

import (
	"net/http"

	"github.com/aatuh/recsys-suite/api/internal/auth"
	"github.com/aatuh/recsys-suite/api/internal/config"
	"github.com/aatuh/recsys-suite/api/internal/http/problem"
)

// RequireTenantClaim enforces that a tenant claim is present when configured.
type RequireTenantClaim struct {
	cfg config.AuthConfig
}

// NewRequireTenantClaim constructs the middleware.
func NewRequireTenantClaim(cfg config.AuthConfig) *RequireTenantClaim {
	return &RequireTenantClaim{cfg: cfg}
}

// Handler wraps the next handler with tenant-claim enforcement.
func (m *RequireTenantClaim) Handler(next http.Handler) http.Handler {
	if m == nil {
		return next
	}
	if next == nil {
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isProtectedPath(r) || !m.cfg.RequireTenantClaim {
			next.ServeHTTP(w, r)
			return
		}
		info, ok := auth.FromContext(r.Context())
		if !ok || info.TenantID == "" || info.TenantSource != auth.TenantSourceClaim {
			problem.Write(w, r, http.StatusForbidden, "RECSYS_TENANT_SCOPE_REQUIRED", "tenant scope required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
