package middleware

import (
	"net/http"
	"strings"

	"github.com/aatuh/recsys-suite/api/internal/auth"
	"github.com/aatuh/recsys-suite/api/internal/config"
	"github.com/aatuh/recsys-suite/api/internal/http/problem"
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAdminPath(r) {
			next.ServeHTTP(w, r)
			return
		}
		if !m.rbacEnabled() {
			next.ServeHTTP(w, r)
			return
		}
		allowed := m.allowedRoles(r.Method)
		if hasAnyRole(auth.RolesFromContext(r.Context()), allowed) {
			next.ServeHTTP(w, r)
			return
		}
		problem.Write(w, r, http.StatusForbidden, "RECSYS_FORBIDDEN", "insufficient scope")
	})
}

func (m *AdminRoleMiddleware) rbacEnabled() bool {
	if m == nil {
		return false
	}
	return strings.TrimSpace(m.cfg.AdminRole) != "" ||
		strings.TrimSpace(m.cfg.OperatorRole) != "" ||
		strings.TrimSpace(m.cfg.ViewerRole) != ""
}

func (m *AdminRoleMiddleware) allowedRoles(method string) []string {
	if m == nil {
		return nil
	}
	viewer := strings.TrimSpace(m.cfg.ViewerRole)
	operator := strings.TrimSpace(m.cfg.OperatorRole)
	admin := strings.TrimSpace(m.cfg.AdminRole)

	allowed := []string{}
	if isReadMethod(method) {
		allowed = appendRole(allowed, viewer)
	}
	allowed = appendRole(allowed, operator)
	allowed = appendRole(allowed, admin)
	return allowed
}

func appendRole(dst []string, role string) []string {
	if role == "" {
		return dst
	}
	for _, existing := range dst {
		if strings.EqualFold(existing, role) {
			return dst
		}
	}
	return append(dst, role)
}

func isReadMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func hasAnyRole(roles []string, allowed []string) bool {
	if len(roles) == 0 || len(allowed) == 0 {
		return false
	}
	for _, role := range roles {
		normalized := strings.TrimSpace(role)
		if normalized == "" {
			continue
		}
		for _, required := range allowed {
			if strings.EqualFold(normalized, required) {
				return true
			}
		}
	}
	return false
}
