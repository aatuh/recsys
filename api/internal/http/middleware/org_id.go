package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"recsys/internal/http/common"
)

type ctxKeyOrgID struct{}

// RequireOrgID ensures requests include a valid X-Org-ID header for tenant isolation.
func RequireOrgID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			header := r.Header.Get("X-Org-ID")
			if header == "" {
				common.BadRequest(w, r, "missing_org_id", "X-Org-ID header is required", nil)
				return
			}
			id, err := uuid.Parse(header)
			if err != nil {
				common.BadRequest(w, r, "invalid_org_id", "X-Org-ID must be a valid UUID", map[string]any{"value": header})
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyOrgID{}, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OrgIDFromContext retrieves the tenant UUID injected by RequireOrgID.
func OrgIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if v, ok := ctx.Value(ctxKeyOrgID{}).(uuid.UUID); ok {
		return v, true
	}
	return uuid.UUID{}, false
}
