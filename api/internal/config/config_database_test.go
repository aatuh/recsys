package config

import (
	"context"
	"testing"
	"time"
)

func TestLoadDatabaseDefaults(t *testing.T) {
	src := baseConfigSource()

	cfg, err := Load(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	db := cfg.Database
	if db.MaxConnIdle != 90*time.Second {
		t.Fatalf("expected max conn idle 90s, got %v", db.MaxConnIdle)
	}
	if db.MaxConnLifetime != 0 {
		t.Fatalf("expected max conn lifetime 0, got %v", db.MaxConnLifetime)
	}
	if db.HealthCheckPeriod != 30*time.Second {
		t.Fatalf("expected health check period 30s, got %v", db.HealthCheckPeriod)
	}
	if db.AcquireTimeout != 5*time.Second {
		t.Fatalf("expected acquire timeout 5s, got %v", db.AcquireTimeout)
	}
	if db.MinConns != 0 {
		t.Fatalf("expected min conns 0, got %d", db.MinConns)
	}
	if db.MaxConns != 10 {
		t.Fatalf("expected max conns 10, got %d", db.MaxConns)
	}
	if db.QueryTimeout != 5*time.Second {
		t.Fatalf("expected query timeout 5s, got %v", db.QueryTimeout)
	}
	if db.RetryAttempts != 3 {
		t.Fatalf("expected retry attempts 3, got %d", db.RetryAttempts)
	}
	if db.RetryInitialBackoff != 50*time.Millisecond {
		t.Fatalf("expected retry backoff 50ms, got %v", db.RetryInitialBackoff)
	}
	if db.RetryMaxBackoff != 500*time.Millisecond {
		t.Fatalf("expected retry max backoff 500ms, got %v", db.RetryMaxBackoff)
	}
}

func TestLoadDatabaseOverrides(t *testing.T) {
	src := baseConfigSource()
	src["DATABASE_MAX_CONN_IDLE"] = "120s"
	src["DATABASE_MAX_CONN_LIFETIME"] = "1h"
	src["DATABASE_HEALTH_CHECK_PERIOD"] = "45s"
	src["DATABASE_ACQUIRE_TIMEOUT"] = "3s"
	src["DATABASE_MIN_CONNS"] = "5"
	src["DATABASE_MAX_CONNS"] = "25"
	src["DATABASE_QUERY_TIMEOUT"] = "2s"
	src["DATABASE_RETRY_ATTEMPTS"] = "4"
	src["DATABASE_RETRY_BACKOFF"] = "75ms"
	src["DATABASE_RETRY_MAX_BACKOFF"] = "500ms"

	cfg, err := Load(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	db := cfg.Database
	if db.MaxConnIdle != 120*time.Second {
		t.Fatalf("unexpected max conn idle: %v", db.MaxConnIdle)
	}
	if db.MaxConnLifetime != time.Hour {
		t.Fatalf("unexpected max conn lifetime: %v", db.MaxConnLifetime)
	}
	if db.HealthCheckPeriod != 45*time.Second {
		t.Fatalf("unexpected health check period: %v", db.HealthCheckPeriod)
	}
	if db.AcquireTimeout != 3*time.Second {
		t.Fatalf("unexpected acquire timeout: %v", db.AcquireTimeout)
	}
	if db.MinConns != 5 {
		t.Fatalf("unexpected min conns: %d", db.MinConns)
	}
	if db.MaxConns != 25 {
		t.Fatalf("unexpected max conns: %d", db.MaxConns)
	}
	if db.QueryTimeout != 2*time.Second {
		t.Fatalf("unexpected query timeout: %v", db.QueryTimeout)
	}
	if db.RetryAttempts != 4 {
		t.Fatalf("unexpected retry attempts: %d", db.RetryAttempts)
	}
	if db.RetryInitialBackoff != 75*time.Millisecond {
		t.Fatalf("unexpected retry backoff: %v", db.RetryInitialBackoff)
	}
	if db.RetryMaxBackoff != 500*time.Millisecond {
		t.Fatalf("unexpected retry max backoff: %v", db.RetryMaxBackoff)
	}
}
