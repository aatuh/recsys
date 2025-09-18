package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"recsys/shared/util"
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
	require.NoError(t, err)

	ctxPing, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	require.NoError(t, pool.Ping(ctxPing))

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
	_, err := pool.Exec(context.Background(), `
TRUNCATE TABLE rec_decisions;
TRUNCATE TABLE events, items, users RESTART IDENTITY;
TRUNCATE TABLE event_type_config;`)
	require.NoError(t, err)
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

// NewTestClient creates a new test client that connects to a running server.
func NewTestClient(t *testing.T) *TestClient {
	t.Helper()

	baseURL := util.MustGetEnv("API_BASE_URL")

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
