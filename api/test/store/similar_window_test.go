package store

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T) *shared.TestClient {
	t.Helper()
	return shared.NewTestClient(t)
}

// TestSimilar_RespectsCoVisWindow validates that the "similar" endpoint
// filters co-vis pairs outside the configured window.
//
// Setup:
//   - Old co-vis: X with Y at now-40d (outside 30-day window)
//   - New co-vis: X with Z at now-1h (within 30-day window)
//
// Expectation:
//   - Results for X contain Z but not Y (server uses 30-day window)
func TestSimilar_RespectsCoVisWindow(t *testing.T) {
	client := newTestClient(t)

	// Clean database to ensure test isolation
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	// Ingest items and user.
	do := func(method, path string, body any, want int) {
		client.DoRequestWithStatus(t, method, path, body, want)
	}

	do("POST", "/v1/items:upsert", map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{"item_id": "X", "available": true},
			{"item_id": "Y", "available": true},
			{"item_id": "Z", "available": true},
		},
	}, http.StatusAccepted)
	do("POST", "/v1/users:upsert", map[string]any{
		"namespace": "default",
		"users":     []map[string]any{{"user_id": "u1"}},
	}, http.StatusAccepted)

	// Old co-vis pair (X,Y) at now-40d (outside 30-day window).
	oldTS := time.Now().UTC().Add(-40 * 24 * time.Hour).Format(time.RFC3339)
	do("POST", "/v1/events:batch", map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "u1", "item_id": "X", "type": 0, "value": 1, "ts": oldTS},
			{"user_id": "u1", "item_id": "Y", "type": 0, "value": 1, "ts": oldTS},
		},
	}, http.StatusAccepted)

	// New co-vis pair (X,Z) at now-1h (within 30-day window).
	newTS := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	do("POST", "/v1/events:batch", map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "u1", "item_id": "X", "type": 0, "value": 1, "ts": newTS},
			{"user_id": "u1", "item_id": "Z", "type": 0, "value": 1, "ts": newTS},
		},
	}, http.StatusAccepted)

	// Similar for X should include Z but not Y with 30-day window.
	body := client.DoRequestWithStatus(t, http.MethodGet, "/v1/items/X/similar?namespace=default&k=10", nil, http.StatusOK)

	var sim []struct {
		ItemID string  `json:"item_id"`
		Score  float64 `json:"score"`
	}
	require.NoError(t, json.Unmarshal(body, &sim))
	require.NotEmpty(t, sim)

	seen := map[string]bool{}
	for _, s := range sim {
		seen[s.ItemID] = true
	}
	require.True(t, seen["Z"], "expected Z within window")
	require.False(t, seen["Y"], "did not expect Y outside window")

}
