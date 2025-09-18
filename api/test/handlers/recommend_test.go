package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

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

	do("POST", "/v1/items:upsert", map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{"item_id": "A", "available": true},
			{"item_id": "B", "available": true},
			{"item_id": "C", "available": true},
		},
	}, http.StatusAccepted)

	do("POST", "/v1/users:upsert", map[string]any{
		"namespace": "default",
		"users":     []map[string]any{{"user_id": "u1"}},
	}, http.StatusAccepted)

	now := time.Now().UTC().Format(time.RFC3339)
	do("POST", "/v1/events:batch", map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "u1", "item_id": "A", "type": 0, "value": 1, "ts": now},
			{"user_id": "u1", "item_id": "B", "type": 3, "value": 1, "ts": now},
		},
	}, http.StatusAccepted)

	// include reasons + exclude B
	body := do("POST", "/v1/recommendations", map[string]any{
		"user_id": "u1", "namespace": "default", "k": 5,
		"include_reasons": true,
		"constraints":     map[string]any{"exclude_item_ids": []string{"B"}},
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

	do("POST", "/v1/items:upsert", map[string]any{
		"namespace": "default",
		"items":     []map[string]any{{"item_id": "A", "available": true}, {"item_id": "B", "available": true}},
	}, http.StatusAccepted)
	do("POST", "/v1/users:upsert", map[string]any{"namespace": "default", "users": []map[string]any{{"user_id": "u1"}}}, http.StatusAccepted)

	now := time.Now().UTC().Format(time.RFC3339)
	do("POST", "/v1/events:batch", map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "u1", "item_id": "A", "type": 0, "value": 1, "ts": now}, // view
			{"user_id": "u1", "item_id": "B", "type": 3, "value": 1, "ts": now}, // purchase
		},
	}, http.StatusAccepted)

	// Initial recs: defaults mean B (purchase=1.0) should beat A (view=0.1)
	body1 := client.DoRequestWithStatus(t, http.MethodPost, "/v1/recommendations", map[string]any{
		"user_id": "u1", "namespace": "default", "k": 2,
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
	do("POST", "/v1/event-types:upsert", map[string]any{
		"namespace": "default",
		"types": []map[string]any{
			{"type": 0, "weight": 1.2},
			{"type": 3, "weight": 0.4},
		},
	}, http.StatusAccepted)

	// Recs after override: A should now top B
	body2 := client.DoRequestWithStatus(t, http.MethodPost, "/v1/recommendations", map[string]any{
		"user_id": "u1", "namespace": "default", "k": 2,
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

	do("POST", "/v1/items:upsert", map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{"item_id": "A", "available": true, "tags": []string{"brand:acme", "category:slots"}},
			{"item_id": "B", "available": true, "tags": []string{"brand:acme", "category:slots"}},
		},
	}, http.StatusAccepted)

	do("POST", "/v1/users:upsert", map[string]any{
		"namespace": "default",
		"users":     []map[string]any{{"user_id": "u1"}},
	}, http.StatusAccepted)

	now := time.Now().UTC().Format(time.RFC3339)
	do("POST", "/v1/events:batch", map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "u1", "item_id": "A", "type": 0, "value": 1, "ts": now},
			{"user_id": "u1", "item_id": "B", "type": 0, "value": 0.5, "ts": now},
		},
	}, http.StatusAccepted)

	bodyNumeric := do("POST", "/v1/recommendations", map[string]any{
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

	bodyFull := do("POST", "/v1/recommendations", map[string]any{
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
}
