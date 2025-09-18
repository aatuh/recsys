package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"recsys/specs/endpoints"
	"recsys/specs/types"
	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestAuditDecisions_EndToEnd(t *testing.T) {
	client := shared.NewTestClient(t)
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	do := func(method, path string, body any, want int) []byte {
		return client.DoRequestWithStatus(t, method, path, body, want)
	}

	// Seed catalog and activity
	do(http.MethodPost, endpoints.ItemsUpsert, types.ItemsUpsertRequest{
		Namespace: "default",
		Items: []types.Item{
			{ItemID: "slot-1", Available: true, Tags: []string{"brand:acme", "category:slots"}},
			{ItemID: "slot-2", Available: true},
		},
	}, http.StatusAccepted)
	do(http.MethodPost, endpoints.UsersUpsert, types.UsersUpsertRequest{
		Namespace: "default",
		Users:     []types.User{{UserID: "player-42"}},
	}, http.StatusAccepted)

	now := time.Now().UTC().Format(time.RFC3339)
	do(http.MethodPost, endpoints.EventsBatch, types.EventsBatchRequest{
		Namespace: "default",
		Events: []types.Event{
			{UserID: "player-42", ItemID: "slot-1", Type: 3, Value: 1, TS: now},
			{UserID: "player-42", ItemID: "slot-2", Type: 0, Value: 1, TS: now},
		},
	}, http.StatusAccepted)

	var recResp types.RecommendResponse

	var listResp types.AuditDecisionListResponse

	overallDeadline := time.Now().Add(5 * time.Second)
	var lastListPayload []byte
	for len(listResp.Decisions) == 0 {
		recBody := do(http.MethodPost, endpoints.Recommendations, types.RecommendRequest{
			UserID:         "player-42",
			Namespace:      "default",
			K:              5,
			IncludeReasons: true,
		}, http.StatusOK)
		require.NoError(t, json.Unmarshal(recBody, &recResp))
		require.NotEmpty(t, recResp.Items)

		pollDeadline := time.Now().Add(1 * time.Second)
		for {
			lastListPayload = do(http.MethodGet, endpoints.AuditDecisions+"?namespace=default&limit=5", nil, http.StatusOK)
			require.NoError(t, json.Unmarshal(lastListPayload, &listResp))
			if len(listResp.Decisions) > 0 || time.Now().After(pollDeadline) {
				break
			}
			time.Sleep(75 * time.Millisecond)
		}

		if len(listResp.Decisions) > 0 || time.Now().After(overallDeadline) {
			break
		}
	}

	if len(listResp.Decisions) == 0 {
		t.Fatalf("timed out waiting for decision trace; last payload: %s", string(lastListPayload))
	}

	recent := listResp.Decisions[0]
	require.Equal(t, "default", recent.Namespace)
	require.NotEmpty(t, recent.DecisionID)
	require.NotEmpty(t, recent.FinalItems)
	require.NotEmpty(t, recent.UserHash)

	detailBody := do(http.MethodGet, endpoints.AuditDecisionByIDPath(recent.DecisionID), nil, http.StatusOK)
	var detailResp types.AuditDecisionDetail
	require.NoError(t, json.Unmarshal(detailBody, &detailResp))
	require.Equal(t, recent.DecisionID, detailResp.DecisionID)
	require.Equal(t, recent.Namespace, detailResp.Namespace)
	require.Equal(t, recent.UserHash, detailResp.UserHash)
	require.Len(t, detailResp.FinalItems, len(recResp.Items))
	require.Equal(t, recResp.Items[0].ItemID, detailResp.FinalItems[0].ItemID)
	require.NotNil(t, detailResp.Config.Alpha)

	searchBody := do(http.MethodPost, endpoints.AuditSearch, types.AuditDecisionsSearchRequest{
		Namespace: "default",
		UserHash:  detailResp.UserHash,
		Limit:     1,
	}, http.StatusOK)
	var searchResp types.AuditDecisionListResponse
	require.NoError(t, json.Unmarshal(searchBody, &searchResp))
	require.Len(t, searchResp.Decisions, 1)
	require.Equal(t, detailResp.DecisionID, searchResp.Decisions[0].DecisionID)
}
