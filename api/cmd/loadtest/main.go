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
	)
	flag.Parse()

	if *requests <= 0 || *concurrency <= 0 {
		fmt.Fprintln(os.Stderr, "n and c must be positive")
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
	printReport(*requests, errCount, durations, statusCounts, total)
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

func printReport(total int, errCount int, durations []time.Duration, statusCounts map[int]int, elapsed time.Duration) {
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p50 := percentile(durations, 0.50)
	p95 := percentile(durations, 0.95)
	p99 := percentile(durations, 0.99)

	success := 0
	for code, count := range statusCounts {
		if code >= 200 && code < 300 {
			success += count
		}
	}

	rps := 0.0
	if elapsed > 0 {
		rps = float64(total) / elapsed.Seconds()
	}

	fmt.Printf("requests: %d  success: %d  errors: %d\n", total, success, errCount)
	fmt.Printf("elapsed: %s  rps: %.2f\n", elapsed.Truncate(time.Millisecond), rps)
	fmt.Printf("latency: p50=%s p95=%s p99=%s\n", p50, p95, p99)
	fmt.Println("status codes:")
	var codes []int
	for code := range statusCounts {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	for _, code := range codes {
		fmt.Printf("  %d: %d\n", code, statusCounts[code])
	}
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
