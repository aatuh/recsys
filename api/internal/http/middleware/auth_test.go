package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestAPIKeyAuthorizer(t *testing.T) {
	allowedOrg := uuid.New()
	otherOrg := uuid.New()

	authorizer := NewAPIKeyAuthorizer(map[string]APIKeyAccess{
		"key-allow-specific": {
			OrgIDs: map[uuid.UUID]struct{}{allowedOrg: {}},
		},
		"key-allow-all": {
			AllowAll: true,
		},
	}, zap.NewNop())

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := authorizer.Middleware(baseHandler)
	handler = RequireOrgID()(handler)

	t.Run("missing key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Org-ID", allowedOrg.String())
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("invalid key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Org-ID", allowedOrg.String())
		req.Header.Set("X-API-Key", "nope")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("forbidden tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Org-ID", otherOrg.String())
		req.Header.Set("X-API-Key", "key-allow-specific")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("allowed specific", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Org-ID", allowedOrg.String())
		req.Header.Set("X-API-Key", "key-allow-specific")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("allowed wildcard", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Org-ID", uuid.New().String())
		req.Header.Set("X-API-Key", "key-allow-all")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})
}

func TestAPIKeyFromContext(t *testing.T) {
	ctx := ContextWithAPIKey(nil, "test-key")
	if key, ok := APIKeyFromContext(ctx); !ok || key != "test-key" {
		t.Fatalf("expected to retrieve api key, got ok=%v key=%q", ok, key)
	}
}
