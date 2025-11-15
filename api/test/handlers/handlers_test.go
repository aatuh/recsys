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

func TestHandlers_IngestSmoke(t *testing.T) {
	client := shared.NewTestClient(t)

	// Clean database to ensure test isolation
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	// items
	client.DoRequestWithStatus(t, http.MethodPost, endpoints.ItemsUpsert, types.ItemsUpsertRequest{
		Namespace: "default",
		Items: []types.Item{
			{ItemID: "i1", Available: true, Price: &[]float64{12.5}[0], Tags: []string{"t1"}},
			{ItemID: "i2", Available: true},
		},
	}, http.StatusAccepted)

	// users
	client.DoRequestWithStatus(t, http.MethodPost, endpoints.UsersUpsert, types.UsersUpsertRequest{
		Namespace: "default",
		Users:     []types.User{{UserID: "u1"}, {UserID: "u2"}},
	}, http.StatusAccepted)

	// events
	client.DoRequestWithStatus(t, http.MethodPost, endpoints.EventsBatch, types.EventsBatchRequest{
		Namespace: "default",
		Events: []types.Event{
			{UserID: "u1", ItemID: "i1", Type: 0, Value: 1},
			{UserID: "u1", ItemID: "i2", Type: 3, Value: 1},
		},
	}, http.StatusAccepted)

	// recommendations (current stub expected; update when real logic is wired)
	body := client.DoRequestWithStatus(t, http.MethodPost, endpoints.Recommendations, types.RecommendRequest{
		UserID:    "u1",
		Namespace: "default",
		K:         5,
	}, http.StatusOK)
	var resp types.RecommendResponse
	require.NoError(t, json.Unmarshal(body, &resp))
	require.Equal(t, "popularity_v1", resp.ModelVersion)
	require.NotEmpty(t, resp.Items)

	// co-vis for i2 should include i1, given u1 touched both
	simBody := client.DoRequestWithStatus(t, http.MethodGet, endpoints.ItemsSimilarPath("i2")+"?namespace=default&k=5", nil, http.StatusOK)
	var sim []types.ScoredItem
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

func TestVersionEndpoint(t *testing.T) {
	client := shared.NewTestClient(t)

	body := client.DoRequestWithStatus(t, http.MethodGet, endpoints.Version, nil, http.StatusOK)
	var resp struct {
		GitCommit   string `json:"git_commit"`
		BuildTime   string `json:"build_time"`
		ModelVersion string `json:"model_version"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.GitCommit)
	require.NotEmpty(t, resp.BuildTime)
	require.NotEmpty(t, resp.ModelVersion)
}
