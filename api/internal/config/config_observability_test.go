package config

import (
	"context"
	"testing"
)

func TestLoadObservabilityDefaults(t *testing.T) {
	src := baseConfigSource()

	cfg, err := Load(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Observability.MetricsEnabled {
		t.Fatal("expected metrics to be enabled by default")
	}
	if cfg.Observability.MetricsPath != "/metrics" {
		t.Fatalf("unexpected metrics path: %q", cfg.Observability.MetricsPath)
	}
	if cfg.Observability.TracingEnabled {
		t.Fatal("expected tracing disabled by default")
	}
	if cfg.Observability.TraceExporter != "stdout" {
		t.Fatalf("expected trace exporter default stdout, got %q", cfg.Observability.TraceExporter)
	}
}

func TestLoadObservabilityCustomizesMetricsPath(t *testing.T) {
	src := baseConfigSource()
	src["OBSERVABILITY_METRICS_ENABLED"] = "true"
	src["OBSERVABILITY_METRICS_PATH"] = "/custom"

	cfg, err := Load(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Observability.MetricsPath != "/custom" {
		t.Fatalf("expected metrics path override, got %q", cfg.Observability.MetricsPath)
	}
}

func TestLoadObservabilityValidatesExporter(t *testing.T) {
	src := baseConfigSource()
	src["OBSERVABILITY_TRACING_ENABLED"] = "true"
	src["OBSERVABILITY_TRACING_EXPORTER"] = "zipkin"

	if _, err := Load(context.Background(), src); err == nil {
		t.Fatal("expected error for unsupported exporter")
	}
}
