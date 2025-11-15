package shared

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"recsys/shared/util"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// MustPool creates a new pool connection and returns it.
func MustPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(getDSN())
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		if isTransientNetworkErr(err) {
			t.Skipf("skipping integration test: database unavailable (%v)", err)
		}
		require.NoError(t, err)
	}

	ctxPing, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if err := pool.Ping(ctxPing); err != nil {
		if isTransientNetworkErr(err) {
			t.Skipf("skipping integration test: database unavailable (%v)", err)
		}
		require.NoError(t, err)
	}

	t.Cleanup(func() { pool.Close() })
	return pool
}

// MustOrgID returns the org ID from the environment or a default value.
func MustOrgID(t *testing.T) uuid.UUID {
	t.Helper()
	idStr := util.MustGetEnv("ORG_ID")
	id, err := uuid.Parse(idStr)
	require.NoError(t, err)
	return id
}

// CleanTables truncates mutable tables between tests.
func CleanTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	dbLock.Lock()
	defer dbLock.Unlock()
	_, err := pool.Exec(context.Background(), `
TRUNCATE TABLE rec_decisions;
TRUNCATE TABLE events, items, users RESTART IDENTITY;
TRUNCATE TABLE event_type_config;`)
	require.NoError(t, err)
}

// WithDatabaseLock provides exclusive access to the shared test database for the duration of fn.
func WithDatabaseLock(t *testing.T, fn func()) {
	t.Helper()
	dbLock.Lock()
	defer dbLock.Unlock()
	fn()
}

// MustHaveEventTypeDefaults guarantees global defaults exist.
func MustHaveEventTypeDefaults(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	// Insert default event types if they don't exist
	_, err := pool.Exec(context.Background(), `
    INSERT INTO event_type_defaults(type, name, weight, half_life_days) VALUES
      (0,'view',0.1,NULL),(1,'click',0.3,NULL),(2,'add',0.7,NULL),(3,'purchase',1.0,NULL),(4,'custom',0.2,NULL)
    ON CONFLICT (type) DO UPDATE
      SET name=EXCLUDED.name,
          weight=EXCLUDED.weight,
          half_life_days=EXCLUDED.half_life_days;
  `)
	require.NoError(t, err)
}

// TestClient wraps an HTTP client for integration testing against a real server.
type TestClient struct {
	Client  *http.Client
	BaseURL string
}

var (
	loadEnvOnce sync.Once
	dbLock      sync.Mutex
)

func init() {
	loadEnvOnce.Do(func() {
		_ = loadTestEnv()
		ensureDefaultEnv()
	})
}

func loadTestEnv() error {
	bases := []string{
		".env.test",
		".env.test.example",
		".env",
		".env.example",
	}
	prefixes := []string{".", "..", "../.."}
	for _, prefix := range prefixes {
		for _, base := range bases {
			candidate := filepath.Join(prefix, base)
			if err := applyEnvFile(candidate); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("unable to load env from %v", bases)
}

func applyEnvFile(filename string) error {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	f, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

func ensureDefaultEnv() {
	if _, err := uuid.Parse(os.Getenv("ORG_ID")); err != nil {
		_ = os.Setenv("ORG_ID", "00000000-0000-0000-0000-000000000000")
	}
	if os.Getenv("API_BASE_URL") == "" {
		_ = os.Setenv("API_BASE_URL", "http://localhost:8000")
	}
	if os.Getenv("DATABASE_URL") == "" {
		_ = os.Setenv("DATABASE_URL", "postgres://localhost:5432/recsys?sslmode=disable")
	}
}

func isTransientNetworkErr(err error) bool {
	if err == nil {
		return false
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") {
		return true
	}
	return false
}

// NewTestClient creates a new test client that connects to a running server.
func NewTestClient(t *testing.T) *TestClient {
	t.Helper()

	baseURL := util.MustGetEnv("API_BASE_URL")
	u, err := url.Parse(baseURL)
	require.NoError(t, err)
	host := u.Host
	if !strings.Contains(host, ":") {
		if u.Scheme == "https" {
			host += ":443"
		} else {
			host += ":80"
		}
	}
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.DialContext(context.Background(), "tcp", host)
	if err != nil {
		if isTransientNetworkErr(err) {
			t.Skipf("skipping integration test: api unavailable (%v)", err)
		}
		require.NoError(t, err)
	}
	if conn != nil {
		_ = conn.Close()
	}

	return &TestClient{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL: baseURL,
	}
}

// DoRequest makes an HTTP request to the running server.
func (tc *TestClient) DoRequest(t *testing.T, method, path string, body any) (*http.Response, []byte) {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewReader(jsonBody)
	}

	url := fmt.Sprintf("%s%s", tc.BaseURL, path)
	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Org-ID", MustOrgID(t).String())

	resp, err := tc.Client.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

// DoRequestWithStatus makes an HTTP request and asserts the expected status code.
func (tc *TestClient) DoRequestWithStatus(t *testing.T, method, path string, body any, expectedStatus int) []byte {
	t.Helper()

	resp, respBody := tc.DoRequest(t, method, path, body)
	require.Equal(t, expectedStatus, resp.StatusCode)

	return respBody
}

// DoRawRequest makes an HTTP request with a raw body (for malformed JSON tests).
func (tc *TestClient) DoRawRequest(t *testing.T, method, path string, body io.Reader, expectedStatus int) []byte {
	t.Helper()

	url := fmt.Sprintf("%s%s", tc.BaseURL, path)
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Org-ID", MustOrgID(t).String())

	resp, err := tc.Client.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, expectedStatus, resp.StatusCode)

	return respBody
}

// MustJSON marshals the given value to JSON and returns a bytes.Reader.
func MustJSON(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

// getDSN reads DATABASE_URL
func getDSN() string {
	return util.MustGetEnv("DATABASE_URL")
}
