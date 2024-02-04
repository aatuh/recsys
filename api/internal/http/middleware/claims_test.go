package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	jwtint "github.com/aatuh/api-toolkit/contrib/v2/integrations/auth/jwt"
	"github.com/aatuh/api-toolkit/contrib/v2/middleware/auth/devheaders"
	"github.com/aatuh/api-toolkit/v2/authorization"
	"github.com/aatuh/api-toolkit/v2/ports"
	"github.com/aatuh/recsys-suite/api/internal/config"
)

func TestClaimsMiddlewareSetsTenantScopeFromDevHeader(t *testing.T) {
	cfg := config.AuthConfig{
		DevHeaders:      devheaders.Config{Enabled: true, UserIDHeader: "X-Dev-User-Id"},
		DevTenantHeader: "X-Dev-Org-Id",
	}
	mw := NewClaimsMiddleware(cfg, ports.NopLogger{})

	req := httptest.NewRequest(http.MethodGet, "/v1/recommend", nil)
	req.Header.Set("X-Dev-Org-Id", "demo")
	req = req.WithContext(jwtint.WithSubject(req.Context(), jwtint.Subject{UserID: "dev-user-1"}))

	var gotTenant, gotUser string
	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope, _ := authorization.ScopeFromContext(r.Context())
		gotTenant = scope.TenantID
		gotUser = scope.UserID
	}))

	handler.ServeHTTP(httptest.NewRecorder(), req)

	if gotTenant != "demo" {
		t.Fatalf("expected tenant scope to be set from dev header, got %q", gotTenant)
	}
	if gotUser != "dev-user-1" {
		t.Fatalf("expected user scope to be set from dev subject, got %q", gotUser)
	}
}
