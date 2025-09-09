package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestHandlers_IngestSmoke(t *testing.T) {
	client := shared.NewTestClient(t)

	// Clean database to ensure test isolation
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	// items
	client.DoRequestWithStatus(t, http.MethodPost, "/v1/items:upsert", map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{"item_id": "i1", "available": true, "price": 12.5, "tags": []string{"t1"}},
			{"item_id": "i2", "available": true},
		},
	}, http.StatusAccepted)

	// users
	client.DoRequestWithStatus(t, http.MethodPost, "/v1/users:upsert", map[string]any{
		"namespace": "default",
		"users":     []map[string]any{{"user_id": "u1"}, {"user_id": "u2"}},
	}, http.StatusAccepted)

	// events
	client.DoRequestWithStatus(t, http.MethodPost, "/v1/events:batch", map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "u1", "item_id": "i1", "type": 0, "value": 1},
			{"user_id": "u1", "item_id": "i2", "type": 3, "value": 1},
		},
	}, http.StatusAccepted)

	// recommendations (current stub expected; update when real logic is wired)
	body := client.DoRequestWithStatus(t, http.MethodPost, "/v1/recommendations", map[string]any{
		"user_id": "u1", "namespace": "default", "k": 5,
	}, http.StatusOK)
	var resp struct {
		ModelVersion string `json:"model_version"`
		Items        []struct {
			ItemID string  `json:"item_id"`
			Score  float64 `json:"score"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	require.Equal(t, "popularity_v1", resp.ModelVersion)
	require.NotEmpty(t, resp.Items)

	// co-vis for i2 should include i1, given u1 touched both
	simBody := client.DoRequestWithStatus(t, http.MethodGet, "/v1/items/i2/similar?namespace=default&k=5", nil, http.StatusOK)
	var sim []struct {
		ItemID string  `json:"item_id"`
		Score  float64 `json:"score"`
	}
	require.NoError(t, json.Unmarshal(simBody, &sim))
	require.NotEmpty(t, sim)
	found := false
	for _, s := range sim {
		if s.ItemID == "i1" {
			found = true
			break
		}
	}
	require.True(t, found, "expected i1 in similar-to-i2")
}
