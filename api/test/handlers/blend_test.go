package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"recsys/internal/store"
	"recsys/specs/endpoints"
	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestRecommend_Blend_EmbeddingTiltsRanking(t *testing.T) {
	client := shared.NewTestClient(t)

	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	s := store.New(pool)
	org := shared.MustOrgID(t)
	ns := "default"

	// Items: P is the anchor for u1 (unavailable so it is not a candidate).
	// A shares embedding with P; B is dissimilar.
	vec := func(idx int) *[]float64 {
		v := make([]float64, store.EmbeddingDims)
		if idx >= 0 && idx < len(v) {
			v[idx] = 1.0
		}
		return &v
	}
	items := []store.ItemUpsert{
		{ItemID: "P", Available: false, Embedding: vec(0)},
		{ItemID: "A", Available: true, Embedding: vec(0)},
		{ItemID: "B", Available: true, Embedding: vec(1)},
	}
	require.NoError(t, s.UpsertItems(context.Background(), org, ns, items))

	// Users via API.
	client.DoRequestWithStatus(t, http.MethodPost, endpoints.UsersUpsert,
		map[string]any{
			"namespace": ns,
			"users": []map[string]any{
				{"user_id": "u1"},
				{"user_id": "u_pop"},
			},
		}, http.StatusAccepted)

	// Equalize base popularity for A and B.
	now := time.Now().UTC().Format(time.RFC3339)
	client.DoRequestWithStatus(t, http.MethodPost, endpoints.EventsBatch,
		map[string]any{
			"namespace": ns,
			"events": []map[string]any{
				{"user_id": "u_pop", "item_id": "A",
					"type": 0, "value": 1, "ts": now},
				{"user_id": "u_pop", "item_id": "B",
					"type": 0, "value": 1, "ts": now},
				// u1 recently interacted with P (the anchor).
				{"user_id": "u1", "item_id": "P",
					"type": 0, "value": 1, "ts": now},
			},
		}, http.StatusAccepted)

	// Ask for blended recs with ALS=1 only.
	body := client.DoRequestWithStatus(t, http.MethodPost,
		endpoints.Recommendations,
		map[string]any{
			"user_id":         "u1",
			"namespace":       ns,
			"k":               2,
			"include_reasons": true,
			"blend": map[string]any{
				"pop":  0.0,
				"cooc": 0.0,
				"als":  1.0,
			},
		}, http.StatusOK)

	var resp struct {
		ModelVersion string `json:"model_version"`
		Items        []struct {
			ItemID  string   `json:"item_id"`
			Score   float64  `json:"score"`
			Reasons []string `json:"reasons"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))

	require.GreaterOrEqual(t, len(resp.Items), 2)
	require.Equal(t, "A", resp.Items[0].ItemID,
		"embedding similarity to anchor should tilt A above B")
	require.Equal(t, "blend_v1", resp.ModelVersion)

	// Top item should include an embedding-related reason.
	hasEmb := false
	for _, r := range resp.Items[0].Reasons {
		if strings.Contains(strings.ToLower(r), "embedding") {
			hasEmb = true
			break
		}
	}
	require.True(t, hasEmb,
		`expected an "embedding" reason on the top item`)
}

func TestRecommend_Blend_CoVisTiltsRanking(t *testing.T) {
	client := shared.NewTestClient(t)

	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	s := store.New(pool)
	org := shared.MustOrgID(t)
	ns := "default"

	// Items: P is anchor; C and D are candidates.
	require.NoError(t, s.UpsertItems(context.Background(), org, ns,
		[]store.ItemUpsert{
			{ItemID: "P", Available: false},
			{ItemID: "C", Available: true},
			{ItemID: "D", Available: true},
		}))

	// Users.
	client.DoRequestWithStatus(t, http.MethodPost, endpoints.UsersUpsert,
		map[string]any{
			"namespace": ns,
			"users": []map[string]any{
				{"user_id": "u1"},
				{"user_id": "u2"},
				{"user_id": "u3"},
				{"user_id": "u_pop"},
			},
		}, http.StatusAccepted)

	// Make C co-visited with P by multiple users; D not.
	now := time.Now().UTC().Format(time.RFC3339)
	events := []map[string]any{
		// Global popularity tie for C and D.
		{"user_id": "u_pop", "item_id": "C", "type": 0, "value": 1, "ts": now},
		{"user_id": "u_pop", "item_id": "D", "type": 0, "value": 1, "ts": now},
		// Co-vis patterns: u2 and u3 do P then C.
		{"user_id": "u2", "item_id": "P", "type": 0, "value": 1, "ts": now},
		{"user_id": "u2", "item_id": "C", "type": 0, "value": 1, "ts": now},
		{"user_id": "u3", "item_id": "P", "type": 0, "value": 1, "ts": now},
		{"user_id": "u3", "item_id": "C", "type": 0, "value": 1, "ts": now},
		// u1 touched P recently to seed anchors.
		{"user_id": "u1", "item_id": "P", "type": 0, "value": 1, "ts": now},
	}
	client.DoRequestWithStatus(t, http.MethodPost, endpoints.EventsBatch,
		map[string]any{"namespace": ns, "events": events},
		http.StatusAccepted)

	body := client.DoRequestWithStatus(t, http.MethodPost,
		endpoints.Recommendations,
		map[string]any{
			"user_id":         "u1",
			"namespace":       ns,
			"k":               2,
			"include_reasons": true,
			"blend": map[string]any{
				"pop":  0.0,
				"cooc": 1.0,
				"als":  0.0,
			},
		}, http.StatusOK)

	var resp struct {
		ModelVersion string `json:"model_version"`
		Items        []struct {
			ItemID  string   `json:"item_id"`
			Score   float64  `json:"score"`
			Reasons []string `json:"reasons"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))

	require.GreaterOrEqual(t, len(resp.Items), 2)
	require.Equal(t, "C", resp.Items[0].ItemID,
		"co-vis with anchor should tilt C above D")
	require.Equal(t, "blend_v1", resp.ModelVersion)

	// Top item should include a co-vis related reason.
	hasCoVis := false
	for _, r := range resp.Items[0].Reasons {
		if strings.Contains(strings.ToLower(r), "co_vis") ||
			strings.Contains(strings.ToLower(r), "co-vis") {
			hasCoVis = true
			break
		}
	}
	require.True(t, hasCoVis,
		`expected a "co_visitation" reason on the top item`)
}
