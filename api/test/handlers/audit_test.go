package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

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
	do(http.MethodPost, "/v1/items:upsert", map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{"item_id": "slot-1", "available": true, "tags": []string{"brand:acme", "category:slots"}},
			{"item_id": "slot-2", "available": true},
		},
	}, http.StatusAccepted)
	do(http.MethodPost, "/v1/users:upsert", map[string]any{
		"namespace": "default",
		"users":     []map[string]any{{"user_id": "player-42"}},
	}, http.StatusAccepted)

	now := time.Now().UTC().Format(time.RFC3339)
	do(http.MethodPost, "/v1/events:batch", map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "player-42", "item_id": "slot-1", "type": 3, "value": 1, "ts": now},
			{"user_id": "player-42", "item_id": "slot-2", "type": 0, "value": 1, "ts": now},
		},
	}, http.StatusAccepted)

	var recResp struct {
		Items []struct {
			ItemID  string   `json:"item_id"`
			Reasons []string `json:"reasons"`
		}
	}

	var listResp struct {
		Decisions []struct {
			DecisionID string `json:"decision_id"`
			Namespace  string `json:"namespace"`
			FinalItems []struct {
				ItemID string `json:"item_id"`
			} `json:"final_items"`
			UserHash string `json:"user_hash"`
		}
	}

	overallDeadline := time.Now().Add(5 * time.Second)
	var lastListPayload []byte
	for len(listResp.Decisions) == 0 {
		recBody := do(http.MethodPost, "/v1/recommendations", map[string]any{
			"user_id":         "player-42",
			"namespace":       "default",
			"k":               5,
			"include_reasons": true,
		}, http.StatusOK)
		require.NoError(t, json.Unmarshal(recBody, &recResp))
		require.NotEmpty(t, recResp.Items)

		pollDeadline := time.Now().Add(1 * time.Second)
		for {
			lastListPayload = do(http.MethodGet, "/v1/audit/decisions?namespace=default&limit=5", nil, http.StatusOK)
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

	detailBody := do(http.MethodGet, fmt.Sprintf("/v1/audit/decisions/%s", recent.DecisionID), nil, http.StatusOK)
	var detailResp struct {
		DecisionID string `json:"decision_id"`
		Namespace  string `json:"namespace"`
		UserHash   string `json:"user_hash"`
		FinalItems []struct {
			ItemID  string   `json:"item_id"`
			Reasons []string `json:"reasons"`
		} `json:"final_items"`
		EffectiveConfig struct {
			Alpha *float64 `json:"alpha"`
		} `json:"effective_config"`
	}
	require.NoError(t, json.Unmarshal(detailBody, &detailResp))
	require.Equal(t, recent.DecisionID, detailResp.DecisionID)
	require.Equal(t, recent.Namespace, detailResp.Namespace)
	require.Equal(t, recent.UserHash, detailResp.UserHash)
	require.Len(t, detailResp.FinalItems, len(recResp.Items))
	require.Equal(t, recResp.Items[0].ItemID, detailResp.FinalItems[0].ItemID)
	require.NotNil(t, detailResp.EffectiveConfig.Alpha)

	searchBody := do(http.MethodPost, "/v1/audit/search", map[string]any{
		"namespace": "default",
		"user_hash": detailResp.UserHash,
		"limit":     1,
	}, http.StatusOK)
	var searchResp struct {
		Decisions []struct {
			DecisionID string `json:"decision_id"`
		} `json:"decisions"`
	}
	require.NoError(t, json.Unmarshal(searchBody, &searchResp))
	require.Len(t, searchResp.Decisions, 1)
	require.Equal(t, detailResp.DecisionID, searchResp.Decisions[0].DecisionID)
}
