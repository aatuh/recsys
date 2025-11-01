package config

import (
	"context"
	"testing"
)

func TestLoadProfileDevelopmentDefaults(t *testing.T) {
	src := baseConfigSource()

	cfg, err := Load(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Debug.Environment != "development" {
		t.Fatalf("expected environment development, got %s", cfg.Debug.Environment)
	}
	if cfg.Explain.Enabled {
		t.Fatal("expected explain disabled in development profile")
	}
	if cfg.Audit.DecisionTrace.Enabled {
		t.Fatal("expected decision trace disabled in development profile")
	}
	if cfg.Features.DecisionTrace {
		t.Fatal("expected decision trace feature flag disabled")
	}
	if cfg.Features.Explain {
		t.Fatal("expected explain feature flag disabled")
	}
}

func TestLoadProfileTestDefaults(t *testing.T) {
	src := baseConfigSource()
	src["ENV"] = "test"

	cfg, err := Load(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Explain.Enabled {
		t.Fatal("expected explain disabled in test profile")
	}
	if cfg.Audit.DecisionTrace.Enabled {
		t.Fatal("expected decision trace disabled in test profile")
	}
	if cfg.Observability.MetricsEnabled {
		t.Fatal("expected metrics disabled in test profile")
	}
	if cfg.Observability.TracingEnabled {
		t.Fatal("expected tracing disabled in test profile")
	}
}

func TestLoadProfileProductionDefaults(t *testing.T) {
	src := baseConfigSource()
	src["ENV"] = "production"

	cfg, err := Load(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Explain.Enabled {
		t.Fatal("expected explain enabled in production profile")
	}
	if !cfg.Rules.Enabled {
		t.Fatal("expected rules enabled in production profile")
	}
	if !cfg.Audit.DecisionTrace.Enabled {
		t.Fatal("expected decision trace enabled in production profile")
	}
	if !cfg.Observability.TracingEnabled {
		t.Fatal("expected tracing enabled in production profile")
	}
	if !cfg.Features.Explain || !cfg.Features.Rules || !cfg.Features.DecisionTrace {
		t.Fatalf("expected feature flags to reflect enabled services, got %+v", cfg.Features)
	}
}
