package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestEventTypes_UpsertAndList(t *testing.T) {
	client := shared.NewTestClient(t)

	// Upsert tenant override: boost view (type=0) weight to 0.6
	client.DoRequestWithStatus(t, http.MethodPost, "/v1/event-types:upsert",
		map[string]any{
			"namespace": "default",
			"types": []map[string]any{
				{"type": 0, "name": "view", "weight": 0.6, "is_active": true},
			},
		}, http.StatusAccepted)

	// List effective config
	body := client.DoRequestWithStatus(t, http.MethodGet, "/v1/event-types?namespace=default", nil, http.StatusOK)

	var rows []struct {
		Type         int16    `json:"type"`
		Name         *string  `json:"name"`
		Weight       float64  `json:"weight"`
		HalfLifeDays *float64 `json:"half_life_days"`
		IsActive     bool     `json:"is_active"`
		Source       string   `json:"source"`
	}
	require.NoError(t, json.Unmarshal(body, &rows))
	require.NotEmpty(t, rows)
	found := false
	for _, r := range rows {
		if r.Type == 0 {
			found = true
			require.Equal(t, "tenant", r.Source)
			require.InDelta(t, 0.6, r.Weight, 1e-9)
		}
	}
	require.True(t, found, "expected type 0 in effective config")
}

func TestEventTypes_MissingNamespace_400(t *testing.T) {
	client := shared.NewTestClient(t)

	client.DoRequestWithStatus(t, http.MethodGet, "/v1/event-types", nil, http.StatusBadRequest) // no namespace
}
