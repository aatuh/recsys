package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"recsys/specs/endpoints"
	"recsys/specs/types"
	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestExplainLLMFallbackContract(t *testing.T) {
	client := shared.NewTestClient(t)
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	do := func(method, path string, body any, want int) []byte {
		return client.DoRequestWithStatus(t, method, path, body, want)
	}

	// Prepare catalog, users, and interactions.
	do(http.MethodPost, endpoints.ItemsUpsert, types.ItemsUpsertRequest{
		Namespace: "default",
		Items: []types.Item{
			{ItemID: "item-1", Available: true, Tags: []string{"brand:a", "category:b"}},
			{ItemID: "item-2", Available: true, Tags: []string{"brand:b"}},
		},
	}, http.StatusAccepted)
	do(http.MethodPost, endpoints.UsersUpsert, types.UsersUpsertRequest{
		Namespace: "default",
		Users:     []types.User{{UserID: "user-42"}},
	}, http.StatusAccepted)

	now := time.Now().UTC()
	do(http.MethodPost, endpoints.EventsBatch, types.EventsBatchRequest{
		Namespace: "default",
		Events: []types.Event{
			{UserID: "user-42", ItemID: "item-1", Type: 3, Value: 1, TS: now.Format(time.RFC3339)},
			{UserID: "user-42", ItemID: "item-2", Type: 0, Value: 1, TS: now.Format(time.RFC3339)},
		},
	}, http.StatusAccepted)

	// Ensure at least one decision trace exists to populate explain facts.
	var recResp types.RecommendResponse
	deadline := time.Now().Add(5 * time.Second)
	for {
		body := do(http.MethodPost, endpoints.Recommendations, types.RecommendRequest{
			UserID:         "user-42",
			Namespace:      "default",
			K:              5,
			IncludeReasons: true,
		}, http.StatusOK)
		require.NoError(t, json.Unmarshal(body, &recResp))
		require.NotEmpty(t, recResp.Items)

		// Poll audit list to ensure recorder processed the decision.
		listBody := do(http.MethodGet, endpoints.AuditDecisions+"?namespace=default&limit=1", nil, http.StatusOK)
		var list types.AuditDecisionListResponse
		require.NoError(t, json.Unmarshal(listBody, &list))
		if len(list.Decisions) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for decision trace")
		}
		time.Sleep(100 * time.Millisecond)
	}

	explainReq := types.ExplainLLMRequest{
		TargetType: "item",
		TargetID:   "item-1",
		Namespace:  "default",
		Surface:    "home",
		From:       now.Add(-1 * time.Minute).Format(time.RFC3339),
		To:         time.Now().Add(10 * time.Second).Format(time.RFC3339),
		Question:   "Why is item-1 underperforming?",
	}

	resp, body := client.DoRequest(t, http.MethodPost, endpoints.ExplainLLM, explainReq)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "miss", resp.Header.Get("X-Explain-Cache"))

	var explainResp types.ExplainLLMResponse
	require.NoError(t, json.Unmarshal(body, &explainResp))

	require.Contains(t, explainResp.Markdown, "## Summary")
	require.Equal(t, "fallback", explainResp.Model)
	require.Equal(t, "miss", explainResp.Cache)
	require.NotEmpty(t, explainResp.Facts["metrics"])

	hasDisabledWarning := false
	for _, w := range explainResp.Warnings {
		if strings.Contains(w, "llm_disabled") {
			hasDisabledWarning = true
			break
		}
	}
	require.True(t, hasDisabledWarning, "expected llm_disabled warning")
}
