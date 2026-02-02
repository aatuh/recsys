package middleware

import (
	"net/http"
	"strings"

	"github.com/aatuh/recsys-suite/api/internal/auth"
	"github.com/aatuh/recsys-suite/api/internal/config"

	jwtint "github.com/aatuh/api-toolkit/contrib/v2/integrations/auth/jwt"
	"github.com/aatuh/api-toolkit/v2/authorization"
	"github.com/aatuh/api-toolkit/v2/ports"
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

		info, _ := auth.FromContext(r.Context())
		if subj, ok := jwtint.SubjectFromContext(r.Context()); ok {
			if info.UserID == "" {
				info.UserID = subj.UserID
			}
			if len(subj.Claims) > 0 {
				if info.TenantID == "" {
					if tenant := auth.ExtractTenant(subj.Claims, m.cfg.TenantClaimKeys); tenant != "" {
						info.TenantID = tenant
						info.TenantSource = auth.TenantSourceClaim
					}
				}
				info.Roles = mergeRoles(info.Roles, auth.ExtractRoles(subj.Claims, m.cfg.RoleClaimKeys))
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

func mergeRoles(base, extra []string) []string {
	if len(extra) == 0 {
		return base
	}
	if len(base) == 0 {
		out := make([]string, len(extra))
		copy(out, extra)
		return out
	}
	seen := make(map[string]struct{}, len(base)+len(extra))
	out := make([]string, 0, len(base)+len(extra))
	for _, role := range base {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	for _, role := range extra {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	return out
}
