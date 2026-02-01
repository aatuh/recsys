package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/recsys-suite/api/internal/auth"
	"github.com/aatuh/recsys-suite/api/internal/config"
)

type stubAPIKeyStore struct {
	key      auth.APIKey
	err      error
	lastHash string
}

func (s *stubAPIKeyStore) Lookup(ctx context.Context, hash string) (auth.APIKey, error) {
	s.lastHash = hash
	if s.err != nil {
		return auth.APIKey{}, s.err
	}
	return s.key, nil
}

func TestAPIKeyMiddlewareSetsTenant(t *testing.T) {
	store := &stubAPIKeyStore{key: auth.APIKey{
		ID:               "key-1",
		TenantID:         "tenant-uuid",
		TenantExternalID: "tenant-a",
		Roles:            []string{"admin"},
	}}
	cfg := config.AuthConfig{
		TenantHeader: "X-Org-Id",
		APIKeys: config.APIKeyConfig{
			Enabled:    true,
			Header:     "X-API-Key",
			HashSecret: "pepper",
		},
	}
	mw, err := NewAPIKeyMiddleware(cfg, store, nil)
	if err != nil {
		t.Fatalf("new api key middleware: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/recommend", nil)
	req.Header.Set("X-API-Key", "secret")
	rec := httptest.NewRecorder()

	mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info, ok := auth.FromContext(r.Context())
		if !ok {
			t.Fatalf("missing auth info")
		}
		if info.TenantID != "tenant-a" {
			t.Fatalf("expected tenant-a, got %q", info.TenantID)
		}
		if info.APIKeyID != "key-1" {
			t.Fatalf("expected api key id, got %q", info.APIKeyID)
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if store.lastHash == "" {
		t.Fatalf("expected hash to be computed")
	}
	expectedHash := auth.HashAPIKey("secret", "pepper")
	if store.lastHash != expectedHash {
		t.Fatalf("expected hash %q, got %q", expectedHash, store.lastHash)
	}
}

func TestAPIKeyMiddlewareRejectsMismatchedTenant(t *testing.T) {
	store := &stubAPIKeyStore{key: auth.APIKey{
		ID:               "key-1",
		TenantID:         "tenant-uuid",
		TenantExternalID: "tenant-a",
	}}
	cfg := config.AuthConfig{
		TenantHeader: "X-Org-Id",
		APIKeys: config.APIKeyConfig{
			Enabled: true,
			Header:  "X-API-Key",
		},
	}
	mw, err := NewAPIKeyMiddleware(cfg, store, nil)
	if err != nil {
		t.Fatalf("new api key middleware: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/recommend", nil)
	req.Header.Set("X-API-Key", "secret")
	req.Header.Set("X-Org-Id", "other-tenant")
	rec := httptest.NewRecorder()

	mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
