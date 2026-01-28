package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"recsys/internal/auth"
	"recsys/internal/config"
)

func TestRequireTenantClaimBlocksMissingClaim(t *testing.T) {
	mw := NewRequireTenantClaim(config.AuthConfig{RequireTenantClaim: true})
	req := httptest.NewRequest(http.MethodPost, "/v1/recommend", nil)
	rec := httptest.NewRecorder()

	mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireTenantClaimAllowsClaim(t *testing.T) {
	mw := NewRequireTenantClaim(config.AuthConfig{RequireTenantClaim: true})
	req := httptest.NewRequest(http.MethodPost, "/v1/recommend", nil)
	ctx := auth.WithInfo(req.Context(), auth.Info{
		TenantID:     "tenant-a",
		TenantSource: auth.TenantSourceClaim,
	})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestTenantMiddlewareMismatch(t *testing.T) {
	mw, err := NewTenantMiddleware(config.AuthConfig{TenantHeader: "X-Org-Id"})
	if err != nil {
		t.Fatalf("new tenant middleware: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/recommend", nil)
	req.Header.Set("X-Org-Id", "tenant-b")
	ctx := auth.WithInfo(req.Context(), auth.Info{
		TenantID:     "tenant-a",
		TenantSource: auth.TenantSourceClaim,
	})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestTenantMiddlewareMatch(t *testing.T) {
	mw, err := NewTenantMiddleware(config.AuthConfig{TenantHeader: "X-Org-Id"})
	if err != nil {
		t.Fatalf("new tenant middleware: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/recommend", nil)
	req.Header.Set("X-Org-Id", "tenant-a")
	ctx := auth.WithInfo(req.Context(), auth.Info{
		TenantID:     "tenant-a",
		TenantSource: auth.TenantSourceClaim,
	})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAdminRoleRequiresRole(t *testing.T) {
	mw := NewAdminRoleMiddleware(config.AuthConfig{AdminRole: "admin"})
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/tenants/tenant-a/config", nil)
	rec := httptest.NewRecorder()

	mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAdminRoleAllowsRole(t *testing.T) {
	mw := NewAdminRoleMiddleware(config.AuthConfig{AdminRole: "admin"})
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/tenants/tenant-a/config", nil)
	ctx := auth.WithInfo(req.Context(), auth.Info{Roles: []string{"admin"}})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
