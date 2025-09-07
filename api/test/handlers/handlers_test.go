package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	handlers "recsys/internal/http/handlers"
	mymw "recsys/internal/http/middleware"
	"recsys/internal/http/store"
	"recsys/test/shared"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestRouter(t *testing.T, org uuid.UUID) http.Handler {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)

	r := chi.NewRouter()
	r.Use(chimw.RequestID, chimw.RealIP)
	r.Use(mymw.RequestLogger(logger), mymw.JSONRecovererWithLogger(logger), mymw.ErrorLogger(logger))

	h := &handlers.Handler{Store: store.New(pool), DefaultOrg: org}
	r.Post("/v1/items:upsert", h.ItemsUpsert)
	r.Post("/v1/users:upsert", h.UsersUpsert)
	r.Post("/v1/events:batch", h.EventsBatch)
	r.Post("/v1/recommendations", h.Recommend)
	r.Get("/v1/items/{item_id}/similar", h.ItemSimilar)
	return r
}

func mustJSON(t *testing.T, v any) *bytes.Reader {
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

func TestHandlers_IngestSmoke(t *testing.T) {
	org := shared.MustOrgID(t)
	srv := newTestRouter(t, org)

	// items
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/items:upsert", mustJSON(t, map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{"item_id": "i1", "available": true, "price": 12.5, "tags": []string{"t1"}},
			{"item_id": "i2", "available": true},
		},
	}))
	srv.ServeHTTP(w, req)
	require.Equal(t, http.StatusAccepted, w.Code)

	// users
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/users:upsert", mustJSON(t, map[string]any{
		"namespace": "default",
		"users":     []map[string]any{{"user_id": "u1"}, {"user_id": "u2"}},
	}))
	srv.ServeHTTP(w, req)
	require.Equal(t, http.StatusAccepted, w.Code)

	// events
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/events:batch", mustJSON(t, map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "u1", "item_id": "i1", "type": 0, "value": 1},
			{"user_id": "u1", "item_id": "i2", "type": 3, "value": 1},
		},
	}))
	srv.ServeHTTP(w, req)
	require.Equal(t, http.StatusAccepted, w.Code)

	// recommendations (current stub expected; update when real logic is wired)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/recommendations", mustJSON(t, map[string]any{
		"user_id": "u1", "namespace": "default", "k": 5,
	}))
	srv.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		ModelVersion string `json:"model_version"`
		Items        []any  `json:"items"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	// When you switch to store-backed recommendations, change these expectations:
	require.Equal(t, "pop_0", resp.ModelVersion)
	require.Empty(t, resp.Items)
}
