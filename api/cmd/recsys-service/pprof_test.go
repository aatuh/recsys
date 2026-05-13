package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	toolkitconfig "github.com/aatuh/api-toolkit/contrib/v2/config"
	"github.com/aatuh/recsys-suite/api/internal/config"
)

func TestPprofHandlerDisabledByDefault(t *testing.T) {
	cfg := config.Config{
		Config: toolkitconfig.Config{Addr: "127.0.0.1:8000", Env: "development"},
	}

	if got := systemPprofHandler(cfg); got != nil {
		t.Fatalf("pprofHandler() = %T, want nil", got)
	}
}

func TestPprofHandlerEnabledAfterSafeConfigValidation(t *testing.T) {
	cfg := config.Config{
		Config:      toolkitconfig.Config{Addr: "127.0.0.1:8000", Env: "development"},
		Performance: config.PerformanceConfig{PprofEnabled: true},
	}

	if err := config.Validate(cfg); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	got := systemPprofHandler(cfg)
	if got == nil {
		t.Fatalf("pprofHandler() = nil")
	}
	if got == http.DefaultServeMux {
		t.Fatalf("pprofHandler() used http.DefaultServeMux")
	}
	rec := httptest.NewRecorder()
	got.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("pprof index status = %d, want %d", rec.Code, http.StatusOK)
	}
}
