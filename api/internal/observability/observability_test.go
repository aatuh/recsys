package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"recsys/internal/config"
)

func preserveOtelGlobals(t *testing.T) {
	t.Helper()

	prevProvider := otel.GetTracerProvider()
	prevPropagator := otel.GetTextMapPropagator()

	t.Cleanup(func() {
		otel.SetTracerProvider(prevProvider)
		otel.SetTextMapPropagator(prevPropagator)
	})
}

func TestSetupWithMetricsEnabled(t *testing.T) {
	preserveOtelGlobals(t)

	cfg := config.ObservabilityConfig{
		MetricsEnabled: true,
		MetricsPath:    "/metricsz",
		TracingEnabled: false,
	}

	prov, err := Setup(context.Background(), cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("Setup returned error: %v", err)
	}
	if prov == nil {
		t.Fatal("expected provider, got nil")
	}
	if prov.MetricsMiddleware == nil {
		t.Fatal("expected metrics middleware to be configured")
	}
	if prov.MetricsHandler == nil {
		t.Fatal("expected metrics handler when metrics enabled")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, cfg.MetricsPath, nil)
	prov.MetricsHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected metrics handler to return 200, got %d", rec.Code)
	}
}

func TestSetupWithTracingEnabled(t *testing.T) {
	preserveOtelGlobals(t)

	cfg := config.ObservabilityConfig{
		MetricsEnabled: false,
		TracingEnabled: true,
		TraceExporter:  "stdout",
	}

	prov, err := Setup(context.Background(), cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("Setup returned error: %v", err)
	}
	if prov.TraceMiddleware == nil {
		t.Fatal("expected trace middleware when tracing enabled")
	}
	if prov.Shutdown == nil {
		t.Fatal("expected shutdown func when tracing enabled")
	}
	if err := prov.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}

func TestSetupWithTracingDisabled(t *testing.T) {
	preserveOtelGlobals(t)

	cfg := config.ObservabilityConfig{
		MetricsEnabled: false,
		TracingEnabled: false,
	}

	prov, err := Setup(context.Background(), cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("Setup returned error: %v", err)
	}
	if prov.TraceMiddleware == nil {
		t.Fatal("expected trace middleware to default to noop provider")
	}
	if prov.Shutdown != nil {
		t.Fatal("expected no shutdown func when tracing disabled")
	}
}
