package middleware

import (
	"net/http"
	"strings"

	"recsys/internal/auth"
	"recsys/internal/config"

	jwtint "github.com/aatuh/api-toolkit-contrib/integrations/auth/jwt"
	"github.com/aatuh/api-toolkit/authorization"
	"github.com/aatuh/api-toolkit/ports"
)

// ClaimsMiddleware enriches context with tenant/user/roles extracted from JWTs or dev headers.
type ClaimsMiddleware struct {
	cfg config.AuthConfig
	log ports.Logger
}

// NewClaimsMiddleware constructs a claims extractor.
func NewClaimsMiddleware(cfg config.AuthConfig, log ports.Logger) *ClaimsMiddleware {
	if log == nil {
		log = ports.NopLogger{}
	}
	return &ClaimsMiddleware{cfg: cfg, log: log}
}

// Handler wraps the next handler with claims extraction.
func (m *ClaimsMiddleware) Handler(next http.Handler) http.Handler {
	if m == nil {
		return next
	}
	if next == nil {
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isProtectedPath(r) {
			next.ServeHTTP(w, r)
			return
		}

		info := auth.Info{}
		if subj, ok := jwtint.SubjectFromContext(r.Context()); ok {
			info.UserID = subj.UserID
			if len(subj.Claims) > 0 {
				if tenant := auth.ExtractTenant(subj.Claims, m.cfg.TenantClaimKeys); tenant != "" {
					info.TenantID = tenant
					info.TenantSource = auth.TenantSourceClaim
				}
				info.Roles = auth.ExtractRoles(subj.Claims, m.cfg.RoleClaimKeys)
			}
		}

		if info.TenantID == "" && m.cfg.DevHeaders.Enabled {
			if header := strings.TrimSpace(m.cfg.DevTenantHeader); header != "" {
				if tenant := strings.TrimSpace(r.Header.Get(header)); tenant != "" {
					info.TenantID = tenant
					info.TenantSource = auth.TenantSourceDev
				}
			}
		}

		ctx := auth.WithInfo(r.Context(), info)
		if info.UserID != "" {
			scope, _ := authorization.ScopeFromContext(ctx)
			scope.UserID = info.UserID
			ctx = authorization.WithScope(ctx, scope)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
