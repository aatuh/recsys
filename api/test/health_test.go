package test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aatuh/api-toolkit/v2/specs"
)

type healthResponse struct {
	Status string `json:"status"`
}

func TestHealthEndpoint(t *testing.T) {
	cfg := MustLoadConfig(t)

	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(cfg.APIHost, "/")+specs.HealthEndpoints.Health, nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to call /health endpoint: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 OK, got %s, body: %s", resp.Status, string(body))
	}

	var payload healthResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if payload.Status != "healthy" {
		t.Fatalf("expected health status 'healthy', got '%s'", payload.Status)
	}
}
