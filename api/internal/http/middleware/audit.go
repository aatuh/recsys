package middleware

import (
	"net/http"
	"strings"

	"github.com/aatuh/recsys-suite/api/internal/audit"
	"github.com/aatuh/recsys-suite/api/internal/auth"
	"github.com/aatuh/recsys-suite/api/internal/config"
	"github.com/aatuh/recsys-suite/api/internal/http/problem"

	"github.com/aatuh/api-toolkit/authorization"
	"github.com/aatuh/api-toolkit/response_writer"
)

// AuditMiddleware logs admin actions to an audit sink.
type AuditMiddleware struct {
	logger audit.Logger
}

// NewAuditMiddleware constructs the middleware.
func NewAuditMiddleware(cfg config.AuditConfig, logger audit.Logger) *AuditMiddleware {
	if !cfg.Enabled || logger == nil {
		return &AuditMiddleware{}
	}
	return &AuditMiddleware{logger: logger}
}

// Handler wraps the next handler with audit logging.
func (m *AuditMiddleware) Handler(next http.Handler) http.Handler {
	if m == nil || m.logger == nil {
		return next
	}
	if next == nil {
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAdminPath(r) || !isAuditMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}
		ww := response_writer.Wrap(w)
		next.ServeHTTP(ww, r)

		info, _ := auth.FromContext(r.Context())
		tenant, _ := authorization.TenantIDFromContext(r.Context())
		event := audit.Event{
			RequestID: problem.RequestID(r),
			ActorID:   info.UserID,
			TenantID:  tenant,
			Method:    r.Method,
			Path:      r.URL.Path,
			Action:    actionName(r),
			Status:    ww.Status(),
		}
		_ = m.logger.Log(r.Context(), event)
	})
}

func isAuditMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func actionName(r *http.Request) string {
	if r == nil || r.URL == nil {
		return "admin_action"
	}
	return r.Method + " " + r.URL.Path
}
