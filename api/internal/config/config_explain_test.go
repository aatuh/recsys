package config

import (
	"context"
	"testing"
	"time"
)

func TestLoadExplainDefaults(t *testing.T) {
	src := baseConfigSource()

	cfg, err := Load(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Explain.CircuitBreaker.Enabled {
		t.Fatal("expected circuit breaker disabled by default")
	}
	if cfg.Explain.CircuitBreaker.FailureThreshold != 3 {
		t.Fatalf("unexpected failure threshold: %d", cfg.Explain.CircuitBreaker.FailureThreshold)
	}
	if cfg.Explain.CircuitBreaker.ResetAfter != time.Minute {
		t.Fatalf("unexpected reset interval: %v", cfg.Explain.CircuitBreaker.ResetAfter)
	}
	if cfg.Explain.CircuitBreaker.HalfOpenSuccesses != 1 {
		t.Fatalf("unexpected half-open successes: %d", cfg.Explain.CircuitBreaker.HalfOpenSuccesses)
	}
}

func TestLoadExplainBreakerOverrides(t *testing.T) {
	src := baseConfigSource()
	src["LLM_BREAKER_ENABLED"] = "true"
	src["LLM_BREAKER_FAILURES"] = "5"
	src["LLM_BREAKER_RESET"] = "45s"
	src["LLM_BREAKER_HALF_OPEN_SUCCESS"] = "2"

	cfg, err := Load(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Explain.CircuitBreaker.Enabled {
		t.Fatal("expected circuit breaker enabled")
	}
	if cfg.Explain.CircuitBreaker.FailureThreshold != 5 {
		t.Fatalf("expected failure threshold 5, got %d", cfg.Explain.CircuitBreaker.FailureThreshold)
	}
	if cfg.Explain.CircuitBreaker.ResetAfter != 45*time.Second {
		t.Fatalf("expected reset after 45s, got %v", cfg.Explain.CircuitBreaker.ResetAfter)
	}
	if cfg.Explain.CircuitBreaker.HalfOpenSuccesses != 2 {
		t.Fatalf("expected half-open successes 2, got %d", cfg.Explain.CircuitBreaker.HalfOpenSuccesses)
	}
}
