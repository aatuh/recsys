package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestSimilar_KLimitAndPresence(t *testing.T) {
	client := shared.NewTestClient(t)

	// Clean database to ensure test isolation
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	// ingest co-visitation: u1 touched X,Y,Z; u2 touched X,Z
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

	do("POST", "/v1/users:upsert", map[string]any{"namespace": "default", "users": []map[string]any{{"user_id": "u1"}, {"user_id": "u2"}}}, http.StatusAccepted)
	now := time.Now().UTC().Format(time.RFC3339)
	do("POST", "/v1/events:batch", map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "u1", "item_id": "X", "type": 0, "value": 1, "ts": now},
			{"user_id": "u1", "item_id": "Y", "type": 0, "value": 1, "ts": now},
			{"user_id": "u1", "item_id": "Z", "type": 0, "value": 1, "ts": now},
			{"user_id": "u2", "item_id": "X", "type": 0, "value": 1, "ts": now},
			{"user_id": "u2", "item_id": "Z", "type": 0, "value": 1, "ts": now},
		},
	}, http.StatusAccepted)

	// ask similar to X with k=1
	body := client.DoRequestWithStatus(t, http.MethodGet, "/v1/items/X/similar?namespace=default&k=1", nil, http.StatusOK)

	var sim []struct {
		ItemID string  `json:"item_id"`
		Score  float64 `json:"score"`
	}
	require.NoError(t, json.Unmarshal(body, &sim))
	require.LessOrEqual(t, len(sim), 1)
	require.NotEmpty(t, sim) // should have either Y or Z
}
