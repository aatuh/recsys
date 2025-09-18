package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"recsys/specs/endpoints"
	"recsys/specs/types"
	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestEventTypes_UpsertAndList(t *testing.T) {
	client := shared.NewTestClient(t)

	// Upsert tenant override: boost view (type=0) weight to 0.6
	client.DoRequestWithStatus(t, http.MethodPost, endpoints.EventTypesUpsert,
		types.EventTypeConfigUpsertRequest{
			Namespace: "default",
			Types: []types.EventTypeConfig{
				{Type: 0, Name: &[]string{"view"}[0], Weight: 0.6, IsActive: &[]bool{true}[0]},
			},
		}, http.StatusAccepted)

	// List effective config
	body := client.DoRequestWithStatus(t, http.MethodGet, endpoints.EventTypes+"?namespace=default", nil, http.StatusOK)

	var rows []types.EventTypeConfigUpsertResponse
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

	client.DoRequestWithStatus(t, http.MethodGet, endpoints.EventTypes, nil, http.StatusBadRequest) // no namespace
}
