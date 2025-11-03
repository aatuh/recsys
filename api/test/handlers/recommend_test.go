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

func TestRecommend_ReasonsAndExcludes(t *testing.T) {
	client := shared.NewTestClient(t)

	// Clean database to ensure test isolation
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	// upsert items/users/events
	do := func(method, path string, body any, want int) []byte {
		return client.DoRequestWithStatus(t, method, path, body, want)
	}

	do("POST", endpoints.ItemsUpsert, types.ItemsUpsertRequest{
		Namespace: "default",
		Items: []types.Item{
			{ItemID: "A", Available: true},
			{ItemID: "B", Available: true},
			{ItemID: "C", Available: true},
		},
	}, http.StatusAccepted)

	do("POST", endpoints.UsersUpsert, types.UsersUpsertRequest{
		Namespace: "default",
		Users:     []types.User{{UserID: "u1"}},
	}, http.StatusAccepted)

	now := time.Now().UTC().Format(time.RFC3339)
	do("POST", endpoints.EventsBatch, types.EventsBatchRequest{
		Namespace: "default",
		Events: []types.Event{
			{UserID: "u1", ItemID: "A", Type: 0, Value: 1, TS: now},
			{UserID: "u1", ItemID: "B", Type: 3, Value: 1, TS: now},
		},
	}, http.StatusAccepted)

	// include reasons + exclude B
	body := do("POST", endpoints.Recommendations, types.RecommendRequest{
		UserID:         "u1",
		Namespace:      "default",
		K:              5,
		IncludeReasons: true,
		Constraints:    &types.RecommendConstraints{ExcludeItemIDs: []string{"B"}},
	}, http.StatusOK)

	var resp struct {
		ModelVersion string `json:"model_version"`
		Items        []struct {
			ItemID  string    `json:"item_id"`
			Score   float64   `json:"score"`
			Reasons []string  `json:"reasons"`
			Explain *struct{} `json:"explain"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	require.Equal(t, "popularity_v1", resp.ModelVersion)
	require.NotEmpty(t, resp.Items)
	for _, it := range resp.Items {
		require.NotEqual(t, "B", it.ItemID) // excluded
		require.NotEmpty(t, it.Reasons)     // requested
		require.Nil(t, it.Explain)          // default explain level omits block
	}
}

func TestRecommend_TenantOverrideAffectsRanking(t *testing.T) {
	client := shared.NewTestClient(t)

	// Clean database to ensure test isolation
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	// baseline data: A has only views (type 0), B has only purchases (type 3)
	do := func(method, path string, body any, want int) {
		client.DoRequestWithStatus(t, method, path, body, want)
	}

	do("POST", endpoints.ItemsUpsert, map[string]any{
		"namespace": "default",
		"items":     []map[string]any{{"item_id": "A", "available": true}, {"item_id": "B", "available": true}},
	}, http.StatusAccepted)
	do("POST", endpoints.UsersUpsert, map[string]any{"namespace": "default", "users": []map[string]any{{"user_id": "u1"}}}, http.StatusAccepted)

	now := time.Now().UTC().Format(time.RFC3339)
	do("POST", endpoints.EventsBatch, map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "u1", "item_id": "A", "type": 0, "value": 1, "ts": now}, // view
			{"user_id": "u1", "item_id": "B", "type": 3, "value": 1, "ts": now}, // purchase
		},
	}, http.StatusAccepted)

	// Initial recs: defaults mean B (purchase=1.0) should beat A (view=0.1)
	body1 := client.DoRequestWithStatus(t, http.MethodPost, endpoints.Recommendations, map[string]any{
		"user_id": "u1", "namespace": "default", "k": 2,
		"overrides": map[string]any{
			"rule_exclude_events": false,
		},
	}, http.StatusOK)
	var r1 struct {
		Items []struct {
			ItemID string `json:"item_id"`
		}
	}
	require.NoError(t, json.Unmarshal(body1, &r1))
	require.GreaterOrEqual(t, len(r1.Items), 2)
	require.Equal(t, "B", r1.Items[0].ItemID)

	// Override: boost view weight above purchase
	do("POST", endpoints.EventTypesUpsert, map[string]any{
		"namespace": "default",
		"types": []map[string]any{
			{"type": 0, "weight": 1.2},
			{"type": 3, "weight": 0.4},
		},
	}, http.StatusAccepted)

	// Recs after override: A should now top B
	body2 := client.DoRequestWithStatus(t, http.MethodPost, endpoints.Recommendations, map[string]any{
		"user_id": "u1", "namespace": "default", "k": 2,
		"overrides": map[string]any{
			"rule_exclude_events": false,
		},
	}, http.StatusOK)
	var r2 struct {
		Items []struct {
			ItemID string `json:"item_id"`
		}
	}
	require.NoError(t, json.Unmarshal(body2, &r2))
	require.GreaterOrEqual(t, len(r2.Items), 2)
	require.Equal(t, "A", r2.Items[0].ItemID)
}

func TestRecommend_ExplainLevels(t *testing.T) {
	client := shared.NewTestClient(t)

	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	do := func(method, path string, body any, want int) []byte {
		return client.DoRequestWithStatus(t, method, path, body, want)
	}

	do("POST", endpoints.ItemsUpsert, map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{"item_id": "A", "available": true, "tags": []string{"brand:acme", "category:slots"}},
			{"item_id": "B", "available": true, "tags": []string{"brand:acme", "category:slots"}},
		},
	}, http.StatusAccepted)

	do("POST", endpoints.UsersUpsert, types.UsersUpsertRequest{
		Namespace: "default",
		Users:     []types.User{{UserID: "u1"}},
	}, http.StatusAccepted)

	now := time.Now().UTC().Format(time.RFC3339)
	do("POST", endpoints.EventsBatch, map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "u1", "item_id": "A", "type": 0, "value": 1, "ts": now},
			{"user_id": "u1", "item_id": "B", "type": 0, "value": 0.5, "ts": now},
		},
	}, http.StatusAccepted)

	bodyNumeric := do("POST", endpoints.Recommendations, map[string]any{
		"user_id":   "u1",
		"namespace": "default",
		"k":         1,
		"constraints": map[string]any{
			"exclude_item_ids": []string{},
		},
		"include_reasons": true,
		"explain_level":   "numeric",
	}, http.StatusOK)

	var respNumeric struct {
		Items []struct {
			ItemID  string  `json:"item_id"`
			Score   float64 `json:"score"`
			Explain struct {
				Blend struct {
					Alpha         float64 `json:"alpha"`
					Beta          float64 `json:"beta"`
					Gamma         float64 `json:"gamma"`
					PopNorm       float64 `json:"pop_norm"`
					CoocNorm      float64 `json:"cooc_norm"`
					EmbNorm       float64 `json:"emb_norm"`
					Contributions struct {
						Pop  float64 `json:"pop"`
						Cooc float64 `json:"cooc"`
						Emb  float64 `json:"emb"`
					} `json:"contrib"`
				} `json:"blend"`
				Anchors []string `json:"anchors"`
			} `json:"explain"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(bodyNumeric, &respNumeric))
	require.NotEmpty(t, respNumeric.Items)
	numericExplain := respNumeric.Items[0].Explain
	require.NotNil(t, numericExplain)
	require.NotEmpty(t, numericExplain.Anchors)
	expected := numericExplain.Blend.Alpha*numericExplain.Blend.PopNorm +
		numericExplain.Blend.Beta*numericExplain.Blend.CoocNorm +
		numericExplain.Blend.Gamma*numericExplain.Blend.EmbNorm
	require.InDelta(t,
		expected,
		numericExplain.Blend.Contributions.Pop+
			numericExplain.Blend.Contributions.Cooc+
			numericExplain.Blend.Contributions.Emb,
		1e-6,
	)

	bodyFull := do("POST", endpoints.Recommendations, map[string]any{
		"user_id":         "u1",
		"namespace":       "default",
		"k":               1,
		"include_reasons": true,
		"explain_level":   "full",
	}, http.StatusOK)

	var respFull struct {
		Items []struct {
			Explain struct {
				Blend struct {
					Raw *struct {
						Pop float64 `json:"pop"`
					} `json:"raw"`
				} `json:"blend"`
			} `json:"explain"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(bodyFull, &respFull))
	require.NotEmpty(t, respFull.Items)
	require.NotNil(t, respFull.Items[0].Explain.Blend.Raw)
	require.Greater(t, respFull.Items[0].Explain.Blend.Raw.Pop, 0.0)

	bodyTrace := do("POST", endpoints.Recommendations, map[string]any{
		"user_id":         "u1",
		"namespace":       "default",
		"k":               1,
		"include_reasons": true,
		"context": map[string]any{
			"surface": "home",
		},
	}, http.StatusOK)

	var respTrace struct {
		Trace struct {
			Extras struct {
				Sources map[string]struct {
					Count int     `json:"count"`
					Ms    float64 `json:"duration_ms"`
				} `json:"candidate_sources"`
			} `json:"extras"`
		} `json:"trace"`
	}
	require.NoError(t, json.Unmarshal(bodyTrace, &respTrace))
	require.NotNil(t, respTrace.Trace.Extras.Sources)
	require.NotZero(t, len(respTrace.Trace.Extras.Sources))

	for src, metric := range respTrace.Trace.Extras.Sources {
		switch src {
		case "popularity", "collaborative", "content", "session", "merged", "post_exclusion":
			// expected sources
		default:
			t.Fatalf("unexpected source metric %s", src)
		}
		require.GreaterOrEqual(t, metric.Count, 0)
		require.GreaterOrEqual(t, metric.Ms, 0.0)
	}
}

func TestManualOverrideBoostSurfacedItem(t *testing.T) {
	client := shared.NewTestClient(t)

	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	do := func(method, path string, body any, want int) []byte {
		return client.DoRequestWithStatus(t, method, path, body, want)
	}

	do(http.MethodPost, endpoints.ItemsUpsert, map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{"item_id": "shoe_a", "available": true},
			{"item_id": "shoe_b", "available": true},
			{"item_id": "shoe_c", "available": true},
			{"item_id": "watch_gps", "available": true},
		},
	}, http.StatusAccepted)

	do(http.MethodPost, endpoints.UsersUpsert, map[string]any{
		"namespace": "default",
		"users":     []map[string]any{{"user_id": "runner"}},
	}, http.StatusAccepted)

	now := time.Now().UTC().Format(time.RFC3339)
	do(http.MethodPost, endpoints.EventsBatch, map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "runner", "item_id": "shoe_a", "type": 3, "value": 1, "ts": now},
			{"user_id": "runner", "item_id": "shoe_b", "type": 3, "value": 1, "ts": now},
			{"user_id": "runner", "item_id": "shoe_c", "type": 3, "value": 1, "ts": now},
			{"user_id": "runner", "item_id": "watch_gps", "type": 0, "value": 1, "ts": now},
		},
	}, http.StatusAccepted)

	type recItem struct {
		ItemID string  `json:"item_id"`
		Score  float64 `json:"score"`
	}
	parseItems := func(b []byte) []recItem {
		var resp struct {
			Items []recItem `json:"items"`
		}
		require.NoError(t, json.Unmarshal(b, &resp))
		return resp.Items
	}

	req := map[string]any{
		"user_id":   "runner",
		"namespace": "default",
		"k":         3,
		"context":   map[string]any{"surface": "home"},
		"overrides": map[string]any{
			"rule_exclude_events": false,
		},
	}
	initial := do(http.MethodPost, endpoints.Recommendations, req, http.StatusOK)
	itemsBefore := parseItems(initial)
	require.GreaterOrEqual(t, len(itemsBefore), 3)
	initialTopScore := itemsBefore[0].Score
	for _, it := range itemsBefore {
		require.NotEqual(t, "watch_gps", it.ItemID)
	}

	boostValue := 5.0
	do(http.MethodPost, endpoints.ManualOverrides, types.ManualOverrideRequest{
		Namespace:  "default",
		Surface:    "home",
		ItemID:     "watch_gps",
		Action:     "boost",
		BoostValue: &boostValue,
		CreatedBy:  "test-suite",
	}, http.StatusCreated)

	time.Sleep(3 * time.Second)

	after := do(http.MethodPost, endpoints.Recommendations, req, http.StatusOK)
	itemsAfter := parseItems(after)
	require.GreaterOrEqual(t, len(itemsAfter), 3)

	var boosted *recItem
	for _, it := range itemsAfter {
		if it.ItemID == "watch_gps" {
			boosted = &it
			break
		}
	}
	require.NotNil(t, boosted, "boosted item should appear in results")
	require.Greater(t, boosted.Score, initialTopScore, "boosted item should outrank previous best score")
}
