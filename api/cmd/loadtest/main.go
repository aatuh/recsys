package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type report struct {
	GeneratedAt         string        `json:"generated_at"`
	URL                 string        `json:"url"`
	Endpoint            string        `json:"endpoint"`
	Surface             string        `json:"surface"`
	Tenant              string        `json:"tenant"`
	Requests            int           `json:"requests"`
	Concurrency         int           `json:"concurrency"`
	UserCardinality     int           `json:"user_cardinality"`
	Success             int           `json:"success"`
	Errors              int           `json:"errors"`
	ElapsedMS           int64         `json:"elapsed_ms"`
	RPS                 float64       `json:"rps"`
	LatencyMS           latencyReport `json:"latency_ms"`
	StatusCodes         map[int]int   `json:"status_codes"`
	CatalogSize         int           `json:"catalog_size,omitempty"`
	ArtifactSizeBytes   int64         `json:"artifact_size_bytes,omitempty"`
	CPUNotes            string        `json:"cpu_notes,omitempty"`
	MemoryNotes         string        `json:"memory_notes,omitempty"`
	DegradationBehavior string        `json:"degradation_behavior,omitempty"`
}

type latencyReport struct {
	P50 float64 `json:"p50"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

type result struct {
	status int
	dur    time.Duration
	err    error
}

func main() {
	var (
		baseURL         = flag.String("url", "http://localhost:8000", "base URL")
		endpoint        = flag.String("endpoint", "/v1/recommend", "endpoint path")
		surface         = flag.String("surface", "home", "surface name")
		segment         = flag.String("segment", "", "segment (optional)")
		itemID          = flag.String("item-id", "item_1", "item id for /v1/similar")
		k               = flag.Int("k", 20, "number of items to request")
		userPrefix      = flag.String("user-prefix", "user", "user id prefix")
		userCardinality = flag.Int("user-cardinality", 1000, "unique users to cycle through")
		tenantID        = flag.String("tenant", "demo", "tenant id/header value")
		tenantHeader    = flag.String("tenant-header", "X-Org-Id", "tenant header name")
		devHeaders      = flag.Bool("dev", true, "send dev auth headers")
		devTenantHeader = flag.String("dev-tenant-header", "X-Dev-Org-Id", "dev tenant header name")
		devUserHeader   = flag.String("dev-user-header", "X-Dev-User-Id", "dev user header name")
		bearerToken     = flag.String("bearer", "", "bearer token (optional)")
		apiKey          = flag.String("api-key", "", "API key (optional)")
		apiKeyHeader    = flag.String("api-key-header", "X-API-Key", "API key header name")
		requests        = flag.Int("n", 200, "number of requests")
		concurrency     = flag.Int("c", 10, "concurrency")
		timeout         = flag.Duration("timeout", 10*time.Second, "request timeout")
		reportJSON      = flag.String("report-json", "", "write a JSON benchmark report to this path")
		reportMarkdown  = flag.String("report-markdown", "", "write a Markdown benchmark report to this path")
		catalogSize     = flag.Int("catalog-size", 0, "catalog item count to include in reports")
		artifactSize    = flag.Int64("artifact-size-bytes", 0, "artifact size in bytes to include in reports")
		cpuNotes        = flag.String("cpu-notes", "", "CPU/environment notes to include in reports")
		memoryNotes     = flag.String("memory-notes", "", "memory/environment notes to include in reports")
		degradation     = flag.String("degradation", "", "observed degradation behavior to include in reports")
	)
	flag.Parse()

	if *requests <= 0 || *concurrency <= 0 {
		fmt.Fprintln(os.Stderr, "n and c must be positive")
		os.Exit(1)
	}
	if err := validateReportInputs(*catalogSize, *artifactSize); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	url := strings.TrimRight(*baseURL, "/") + "/" + strings.TrimLeft(*endpoint, "/")
	client := &http.Client{Timeout: *timeout}

	jobs := make(chan int)
	results := make(chan result, *requests)

	var wg sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for idx := range jobs {
				body, err := buildPayload(*endpoint, *surface, *segment, *itemID, *k, *userPrefix, *userCardinality, idx)
				if err != nil {
					results <- result{err: err}
					continue
				}
				req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
				if err != nil {
					results <- result{err: err}
					continue
				}
				req.Header.Set("Content-Type", "application/json")
				if strings.TrimSpace(*tenantHeader) != "" && strings.TrimSpace(*tenantID) != "" {
					req.Header.Set(*tenantHeader, *tenantID)
				}
				if *devHeaders {
					if strings.TrimSpace(*devTenantHeader) != "" && strings.TrimSpace(*tenantID) != "" {
						req.Header.Set(*devTenantHeader, *tenantID)
					}
					if strings.TrimSpace(*devUserHeader) != "" {
						req.Header.Set(*devUserHeader, fmt.Sprintf("%s-%d", *userPrefix, idx))
					}
				}
				if strings.TrimSpace(*bearerToken) != "" {
					req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(*bearerToken))
				}
				if strings.TrimSpace(*apiKey) != "" && strings.TrimSpace(*apiKeyHeader) != "" {
					req.Header.Set(*apiKeyHeader, strings.TrimSpace(*apiKey))
				}

				start := time.Now()
				resp, err := client.Do(req)
				if err != nil {
					results <- result{err: err, dur: time.Since(start)}
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				results <- result{status: resp.StatusCode, dur: time.Since(start)}
			}
		}(i)
	}

	start := time.Now()
	go func() {
		for i := 0; i < *requests; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	var durations []time.Duration
	statusCounts := map[int]int{}
	var errCount int
	for i := 0; i < *requests; i++ {
		res := <-results
		if res.err != nil {
			errCount++
			continue
		}
		statusCounts[res.status]++
		durations = append(durations, res.dur)
	}
	wg.Wait()

	total := time.Since(start)
	rep := buildReport(reportInput{
		GeneratedAt:         time.Now().UTC(),
		URL:                 url,
		Endpoint:            *endpoint,
		Surface:             *surface,
		Tenant:              *tenantID,
		Requests:            *requests,
		Concurrency:         *concurrency,
		UserCardinality:     *userCardinality,
		ErrCount:            errCount,
		Durations:           durations,
		StatusCounts:        statusCounts,
		Elapsed:             total,
		CatalogSize:         *catalogSize,
		ArtifactSizeBytes:   *artifactSize,
		CPUNotes:            *cpuNotes,
		MemoryNotes:         *memoryNotes,
		DegradationBehavior: *degradation,
	})
	printReport(rep)
	if err := writeReports(rep, *reportJSON, *reportMarkdown); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildPayload(endpoint, surface, segment, itemID string, k int, userPrefix string, userCardinality, idx int) ([]byte, error) {
	if strings.Contains(endpoint, "similar") {
		payload := map[string]any{
			"surface": surface,
			"item_id": itemID,
			"k":       k,
		}
		if segment != "" {
			payload["segment"] = segment
		}
		return json.Marshal(payload)
	}
	userID := fmt.Sprintf("%s-%d", userPrefix, idx%max(1, userCardinality))
	payload := map[string]any{
		"surface": surface,
		"k":       k,
		"user": map[string]any{
			"user_id": userID,
		},
	}
	if segment != "" {
		payload["segment"] = segment
	}
	return json.Marshal(payload)
}

type reportInput struct {
	GeneratedAt         time.Time
	URL                 string
	Endpoint            string
	Surface             string
	Tenant              string
	Requests            int
	Concurrency         int
	UserCardinality     int
	ErrCount            int
	Durations           []time.Duration
	StatusCounts        map[int]int
	Elapsed             time.Duration
	CatalogSize         int
	ArtifactSizeBytes   int64
	CPUNotes            string
	MemoryNotes         string
	DegradationBehavior string
}

func buildReport(in reportInput) report {
	durations := append([]time.Duration(nil), in.Durations...)
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	success := 0
	statusCounts := map[int]int{}
	for code, count := range in.StatusCounts {
		statusCounts[code] = count
		if code >= 200 && code < 300 {
			success += count
		}
	}
	rps := 0.0
	if in.Elapsed > 0 {
		rps = float64(in.Requests) / in.Elapsed.Seconds()
	}
	return report{
		GeneratedAt:     in.GeneratedAt.UTC().Format(time.RFC3339),
		URL:             in.URL,
		Endpoint:        in.Endpoint,
		Surface:         in.Surface,
		Tenant:          in.Tenant,
		Requests:        in.Requests,
		Concurrency:     in.Concurrency,
		UserCardinality: in.UserCardinality,
		Success:         success,
		Errors:          in.ErrCount,
		ElapsedMS:       in.Elapsed.Milliseconds(),
		RPS:             rps,
		LatencyMS: latencyReport{
			P50: durationMillis(percentile(durations, 0.50)),
			P95: durationMillis(percentile(durations, 0.95)),
			P99: durationMillis(percentile(durations, 0.99)),
		},
		StatusCodes:         statusCounts,
		CatalogSize:         in.CatalogSize,
		ArtifactSizeBytes:   in.ArtifactSizeBytes,
		CPUNotes:            strings.TrimSpace(in.CPUNotes),
		MemoryNotes:         strings.TrimSpace(in.MemoryNotes),
		DegradationBehavior: strings.TrimSpace(in.DegradationBehavior),
	}
}

func printReport(rep report) {
	fmt.Printf("requests: %d  success: %d  errors: %d\n", rep.Requests, rep.Success, rep.Errors)
	fmt.Printf("elapsed: %dms  rps: %.2f\n", rep.ElapsedMS, rep.RPS)
	fmt.Printf("latency: p50=%.2fms p95=%.2fms p99=%.2fms\n", rep.LatencyMS.P50, rep.LatencyMS.P95, rep.LatencyMS.P99)
	fmt.Println("status codes:")
	var codes []int
	for code := range rep.StatusCodes {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	for _, code := range codes {
		fmt.Printf("  %d: %d\n", code, rep.StatusCodes[code])
	}
}

func writeReports(rep report, jsonPath, markdownPath string) error {
	if strings.TrimSpace(jsonPath) != "" {
		b, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(jsonPath, append(b, '\n'), 0o600); err != nil {
			return fmt.Errorf("write json report: %w", err)
		}
	}
	if strings.TrimSpace(markdownPath) != "" {
		if err := os.WriteFile(markdownPath, []byte(markdownReport(rep)), 0o600); err != nil {
			return fmt.Errorf("write markdown report: %w", err)
		}
	}
	return nil
}

func markdownReport(rep report) string {
	var b strings.Builder
	b.WriteString("# RecSys Load Test Report\n\n")
	fmt.Fprintf(&b, "- Generated at: `%s`\n", rep.GeneratedAt)
	fmt.Fprintf(&b, "- Target: `%s`\n", rep.URL)
	fmt.Fprintf(&b, "- Tenant/surface: `%s` / `%s`\n", rep.Tenant, rep.Surface)
	fmt.Fprintf(&b, "- Requests/concurrency: `%d` / `%d`\n", rep.Requests, rep.Concurrency)
	fmt.Fprintf(&b, "- User cardinality: `%d`\n", rep.UserCardinality)
	if rep.CatalogSize > 0 {
		fmt.Fprintf(&b, "- Catalog size: `%d` items\n", rep.CatalogSize)
	}
	if rep.ArtifactSizeBytes > 0 {
		fmt.Fprintf(&b, "- Artifact size: `%d` bytes\n", rep.ArtifactSizeBytes)
	}
	if rep.CPUNotes != "" {
		fmt.Fprintf(&b, "- CPU notes: %s\n", rep.CPUNotes)
	}
	if rep.MemoryNotes != "" {
		fmt.Fprintf(&b, "- Memory notes: %s\n", rep.MemoryNotes)
	}
	if rep.DegradationBehavior != "" {
		fmt.Fprintf(&b, "- Degradation behavior: %s\n", rep.DegradationBehavior)
	}
	b.WriteString("\n| Metric | Value |\n| --- | ---: |\n")
	fmt.Fprintf(&b, "| Success | %d |\n", rep.Success)
	fmt.Fprintf(&b, "| Errors | %d |\n", rep.Errors)
	fmt.Fprintf(&b, "| Elapsed | %d ms |\n", rep.ElapsedMS)
	fmt.Fprintf(&b, "| RPS | %.2f |\n", rep.RPS)
	fmt.Fprintf(&b, "| p50 | %.2f ms |\n", rep.LatencyMS.P50)
	fmt.Fprintf(&b, "| p95 | %.2f ms |\n", rep.LatencyMS.P95)
	fmt.Fprintf(&b, "| p99 | %.2f ms |\n", rep.LatencyMS.P99)
	b.WriteString("\n## Status Codes\n\n")
	var codes []int
	for code := range rep.StatusCodes {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	for _, code := range codes {
		fmt.Fprintf(&b, "- `%d`: %d\n", code, rep.StatusCodes[code])
	}
	return b.String()
}

func validateReportInputs(catalogSize int, artifactSizeBytes int64) error {
	if catalogSize < 0 {
		return fmt.Errorf("catalog-size must be non-negative")
	}
	if artifactSizeBytes < 0 {
		return fmt.Errorf("artifact-size-bytes must be non-negative")
	}
	return nil
}

func percentile(durations []time.Duration, p float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	if p <= 0 {
		return durations[0]
	}
	if p >= 1 {
		return durations[len(durations)-1]
	}
	pos := int(float64(len(durations)-1) * p)
	return durations[pos]
}

func durationMillis(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
