package middleware

import (
	"net/http"
	"strings"

	"recsys/internal/auth"
	"recsys/internal/config"
	"recsys/internal/http/problem"

	"github.com/aatuh/api-toolkit/middleware/auth/tenant"
	"github.com/go-chi/chi/v5"
)

// TenantMiddleware enforces tenant scoping and header/claim consistency.
type TenantMiddleware struct {
	mw *tenant.Middleware
}

// NewTenantMiddleware constructs a tenant scoping middleware.
func NewTenantMiddleware(cfg config.AuthConfig) (*TenantMiddleware, error) {
	mw, err := tenant.New(tenant.Options{
		HeaderName:        cfg.TenantHeader,
		URLParam:          "tenant_id",
		URLParamExtractor: urlParamExtractor{},
		TenantFromContext: auth.TenantIDFromContext,
		ErrorHandler:      tenantErrorHandler,
	})
	if err != nil {
		return nil, err
	}
	return &TenantMiddleware{mw: mw}, nil
}

// Handler wraps the next handler with tenant checks.
func (m *TenantMiddleware) Handler(next http.Handler) http.Handler {
	if m == nil || m.mw == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isProtectedPath(r) {
			next.ServeHTTP(w, r)
			return
		}
		m.mw.Handler(next).ServeHTTP(w, r)
	})
}

type urlParamExtractor struct{}

func (urlParamExtractor) URLParam(r *http.Request, key string) string {
	if r == nil {
		return ""
	}
	return chi.URLParam(r, key)
}

const (
	tenantMissingMsg       = "tenant scope missing"
	tenantMismatchMsg      = "tenant scope mismatch"
	tenantInvalidHeaderMsg = "tenant header is invalid"
	tenantMisconfigMsg     = "tenant sources not configured"
)

func tenantErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	code := "RECSYS_TENANT_SCOPE_INVALID"
	detail := "tenant scope invalid"
	status := http.StatusForbidden
	if err != nil {
		switch {
		case strings.Contains(err.Error(), tenantMissingMsg):
			code = "RECSYS_TENANT_SCOPE_REQUIRED"
			detail = "tenant scope required"
		case strings.Contains(err.Error(), tenantMismatchMsg):
			code = "RECSYS_TENANT_SCOPE_MISMATCH"
			detail = "tenant scope mismatch"
		case strings.Contains(err.Error(), tenantInvalidHeaderMsg):
			code = "RECSYS_TENANT_SCOPE_INVALID"
			detail = "tenant scope invalid"
		case strings.Contains(err.Error(), tenantMisconfigMsg):
			code = "RECSYS_TENANT_SCOPE_MISCONFIGURED"
			detail = "tenant scope misconfigured"
			status = http.StatusInternalServerError
		}
	}
	problem.Write(w, r, status, code, detail)
}
