package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildReportComputesLatencyAndStatusCounts(t *testing.T) {
	rep := buildReport(reportInput{
		GeneratedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		URL:         "http://localhost:8000/v1/recommend",
		Endpoint:    "/v1/recommend",
		Surface:     "home",
		Tenant:      "demo",
		Requests:    4,
		Concurrency: 2,
		ErrCount:    1,
		Durations: []time.Duration{
			10 * time.Millisecond,
			20 * time.Millisecond,
			30 * time.Millisecond,
		},
		StatusCounts:      map[int]int{200: 2, 503: 1},
		Elapsed:           100 * time.Millisecond,
		CatalogSize:       1000,
		ArtifactSizeBytes: 2048,
	})

	if rep.Success != 2 || rep.Errors != 1 {
		t.Fatalf("success/errors = %d/%d, want 2/1", rep.Success, rep.Errors)
	}
	if rep.RPS != 40 {
		t.Fatalf("RPS = %v, want 40", rep.RPS)
	}
	if rep.LatencyMS.P50 != 20 || rep.LatencyMS.P95 != 20 || rep.LatencyMS.P99 != 20 {
		t.Fatalf("LatencyMS = %+v, want p50/p95/p99 20ms", rep.LatencyMS)
	}
	if rep.StatusCodes[503] != 1 {
		t.Fatalf("StatusCodes = %+v, want 503 count", rep.StatusCodes)
	}
}

func TestWriteReportsWritesJSONAndMarkdown(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "report.json")
	mdPath := filepath.Join(dir, "report.md")
	rep := report{
		GeneratedAt: "2026-01-01T00:00:00Z",
		URL:         "http://localhost:8000/v1/recommend",
		Endpoint:    "/v1/recommend",
		Surface:     "home",
		Tenant:      "demo",
		Requests:    1,
		Concurrency: 1,
		Success:     1,
		StatusCodes: map[int]int{200: 1},
	}

	if err := writeReports(rep, jsonPath, mdPath); err != nil {
		t.Fatalf("writeReports() error = %v", err)
	}
	jsonBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json report: %v", err)
	}
	if !strings.Contains(string(jsonBytes), `"requests": 1`) {
		t.Fatalf("json report = %s, want request count", string(jsonBytes))
	}
	mdBytes, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	if !strings.Contains(string(mdBytes), "RecSys Load Test Report") {
		t.Fatalf("markdown report = %s, want title", string(mdBytes))
	}
}

func TestValidateReportInputsRejectsNegativeValues(t *testing.T) {
	if err := validateReportInputs(-1, 0); err == nil {
		t.Fatalf("validateReportInputs() error = nil for negative catalog size")
	}
	if err := validateReportInputs(0, -1); err == nil {
		t.Fatalf("validateReportInputs() error = nil for negative artifact size")
	}
}
