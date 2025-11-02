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

func TestItemsUpsert_MetadataFieldsPersist(t *testing.T) {
	client := shared.NewTestClient(t)

	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	brand := "Acme"
	category := "Shoes"
	categoryPath := []string{"Footwear", "Running"}
	description := "Lightweight training shoe"
	imageURL := "https://cdn.example.com/sku-1.jpg"
	metadataVersion := "2025-10-01"

	req := types.ItemsUpsertRequest{
		Namespace: "default",
		Items: []types.Item{
			{
				ItemID:          "sku-1",
				Available:       true,
				Tags:            []string{"brand:acme", "category:shoes"},
				Brand:           &brand,
				Category:        &category,
				CategoryPath:    &categoryPath,
				Description:     &description,
				ImageURL:        &imageURL,
				MetadataVersion: &metadataVersion,
			},
		},
	}

	client.DoRequestWithStatus(t, http.MethodPost, endpoints.ItemsUpsert, req, http.StatusAccepted)

	body := client.DoRequestWithStatus(t, http.MethodGet, endpoints.ItemsList+"?namespace=default&limit=10", nil, http.StatusOK)
	var resp struct {
		Items []map[string]any `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.Items)

	item := resp.Items[0]
	require.Equal(t, "sku-1", item["item_id"])
	require.Equal(t, brand, item["brand"])
	require.Equal(t, category, item["category"])
	require.Equal(t, metadataVersion, item["metadata_version"])

	rawPath, ok := item["category_path"].([]any)
	require.True(t, ok)
	require.Len(t, rawPath, len(categoryPath))
	for i, v := range rawPath {
		require.Equal(t, categoryPath[i], v.(string))
	}
	require.Equal(t, description, item["description"])
	require.Equal(t, imageURL, item["image_url"])
}
