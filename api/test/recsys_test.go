package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"recsys/src/specs/endpoints"
	"recsys/src/specs/types"
)

func TestRecommendValidate(t *testing.T) {
	cfg := MustLoadConfig(t)

	payload := map[string]any{
		"surface": "home",
		"user": map[string]any{
			"user_id": randID("user"),
		},
	}

	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 3 * time.Second}
	url := strings.TrimRight(cfg.APIHost, "/") + endpoints.RecommendValidate

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	reqID := randID("req")
	req.Header.Set("X-Request-Id", reqID)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %s", resp.Status)
	}
	if got := resp.Header.Get("X-Request-Id"); got != reqID {
		t.Fatalf("expected X-Request-Id echo, got %q", got)
	}

	var out types.ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if out.NormalizedRequest.Segment != "default" {
		t.Fatalf("expected default segment, got %q", out.NormalizedRequest.Segment)
	}
	if out.NormalizedRequest.K != 20 {
		t.Fatalf("expected default k=20, got %d", out.NormalizedRequest.K)
	}
}

func TestRecommendBadRequest(t *testing.T) {
	cfg := MustLoadConfig(t)

	payload := map[string]any{"user": map[string]any{"user_id": randID("user")}}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 3 * time.Second}
	url := strings.TrimRight(cfg.APIHost, "/") + endpoints.Recommend

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %s", resp.Status)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/problem+json") {
		t.Fatalf("expected problem+json, got %q", ct)
	}
}

func TestSimilarEndpoint(t *testing.T) {
	cfg := MustLoadConfig(t)

	payload := map[string]any{
		"surface": "pdp",
		"item_id": randID("item"),
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 3 * time.Second}
	url := strings.TrimRight(cfg.APIHost, "/") + endpoints.Similar

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %s", resp.Status)
	}
}

func randID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
