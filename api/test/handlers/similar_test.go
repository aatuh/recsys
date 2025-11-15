package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"recsys/specs/endpoints"
	"recsys/test/shared"
)

func TestSimilarItems_UsesGeneratedEmbeddings(t *testing.T) {
	client := shared.NewTestClient(t)

	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	payload := map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{
				"item_id":     "anchor",
				"available":   true,
				"brand":       "Voltify",
				"category":    "Electronics",
				"description": "Smart assistant hub",
			},
			{
				"item_id":     "neighbor_a",
				"available":   true,
				"brand":       "Voltify",
				"category":    "Electronics",
				"description": "Companion device",
			},
			{
				"item_id":     "neighbor_b",
				"available":   true,
				"brand":       "LeafPress",
				"category":    "Books",
				"description": "Reader pick",
			},
		},
	}

	client.DoRequestWithStatus(t, http.MethodPost, endpoints.ItemsUpsert, payload, http.StatusAccepted)

	body := client.DoRequestWithStatus(t, http.MethodGet,
		fmt.Sprintf("%s?namespace=default&k=2", endpoints.ItemsSimilarPath("anchor")),
		nil, http.StatusOK)

	var resp []struct {
		ItemID string `json:"item_id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("response not JSON: %v", err)
	}

	if len(resp) == 0 {
		t.Fatalf("expected at least one similar item when embeddings are synthesized")
	}
}
