package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestHTTPMetricsMiddlewareRecordsRequests(t *testing.T) {
	t.Parallel()

	metrics := NewHTTPMetrics()

	router := chi.NewRouter()
	router.Use(metrics.Middleware)
	router.Get("/items", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	total := testutil.ToFloat64(metrics.requests.WithLabelValues(http.MethodGet, "/items", "201"))
	if total != 1 {
		t.Fatalf("expected request counter to be 1, got %v", total)
	}

	if inflight := testutil.ToFloat64(metrics.inflight); inflight != 0 {
		t.Fatalf("expected in-flight gauge to be 0, got %v", inflight)
	}

	if count := testutil.CollectAndCount(metrics.durations); count == 0 {
		t.Fatal("expected histogram to record observations")
	}
}

func TestHTTPMetricsMiddlewareSkipsMetricsEndpoint(t *testing.T) {
	t.Parallel()

	metrics := NewHTTPMetrics()
	metrics.SetMetricsPath("/metrics")

	router := chi.NewRouter()
	router.Use(metrics.Middleware)
	router.Get("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if count := testutil.CollectAndCount(metrics.requests); count != 0 {
		t.Fatalf("expected no metrics for metrics endpoint, got %d collectors", count)
	}
}
