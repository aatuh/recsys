package catalog

import (
	"encoding/json"
	"testing"
	"time"

	"recsys/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuildUpsert_DerivesMetadataAndEmbedding(t *testing.T) {
	price := 129.99
	now := time.Date(2025, time.May, 1, 15, 4, 5, 0, time.UTC)
	props := map[string]any{
		"name":        "Velocity Runner",
		"brand":       "Acme",
		"category":    "Footwear>Running",
		"currency":    "USD",
		"description": "Lightweight shoe for daily training.",
		"image_url":   "https://cdn.example.com/runner.jpg",
	}
	rawProps, err := json.Marshal(props)
	require.NoError(t, err)

	row := store.CatalogItem{
		ItemID:    "sku-123",
		Available: true,
		Price:     &price,
		Tags:      []string{"brand:Acme", "category:Footwear"},
		Props:     rawProps,
		UpdatedAt: now,
	}

	res, err := BuildUpsert(row, Options{GenerateEmbedding: true})
	require.NoError(t, err)
	require.True(t, res.Changed)

	require.NotNil(t, res.Upsert.Brand)
	require.Equal(t, "Acme", *res.Upsert.Brand)

	require.NotNil(t, res.Upsert.Category)
	require.Equal(t, "Footwear>Running", *res.Upsert.Category)

	require.NotNil(t, res.Upsert.CategoryPath)
	require.Equal(t, []string{"Footwear", "Running"}, *res.Upsert.CategoryPath)

	require.NotNil(t, res.Upsert.Description)
	require.Contains(t, *res.Upsert.Description, "Lightweight shoe")

	require.NotNil(t, res.Upsert.ImageURL)
	require.Equal(t, "https://cdn.example.com/runner.jpg", *res.Upsert.ImageURL)

	require.NotNil(t, res.Upsert.MetadataVersion)
	require.Len(t, *res.Upsert.MetadataVersion, 16)

	require.NotNil(t, res.Upsert.Embedding)
	require.Len(t, *res.Upsert.Embedding, EmbeddingDims())

	// Subsequent run with existing metadata should be a no-op.
	row.Brand = res.Upsert.Brand
	row.Category = res.Upsert.Category
	row.CategoryPath = append([]string(nil), (*res.Upsert.CategoryPath)...)
	row.Description = res.Upsert.Description
	row.ImageURL = res.Upsert.ImageURL
	row.MetadataVersion = res.Upsert.MetadataVersion
	row.Embedding = append([]float64(nil), (*res.Upsert.Embedding)...)
	propsAfter, err := json.Marshal(res.Upsert.Props)
	require.NoError(t, err)
	row.Props = propsAfter

	res2, err := BuildUpsert(row, Options{GenerateEmbedding: true})
	require.NoError(t, err)
	require.False(t, res2.Changed)
	require.Nil(t, res2.Upsert.Brand)
	require.Nil(t, res2.Upsert.Embedding)
}

func TestComputeMetadataVersionStable(t *testing.T) {
	price := 10.0
	now := time.Now().UTC().Truncate(time.Second)
	version := computeMetadataVersion(metadataVersionInput{
		ItemID:      "Item-1",
		Name:        "Widget",
		Brand:       "Brand",
		Category:    "Category",
		Description: "Desc",
		ImageURL:    "https://example.com/widget.jpg",
		Price:       &price,
		Currency:    "USD",
		UpdatedAt:   now,
	})

	require.Len(t, version, 16)

	version2 := computeMetadataVersion(metadataVersionInput{
		ItemID:      "Item-1",
		Name:        "Widget",
		Brand:       "Brand",
		Category:    "Category",
		Description: "Desc",
		ImageURL:    "https://example.com/widget.jpg",
		Price:       &price,
		Currency:    "USD",
		UpdatedAt:   now,
	})
	require.Equal(t, version, version2)
}

func BenchmarkDeterministicEmbedding(b *testing.B) {
	existing := make([]float64, EmbeddingDims())
	for i := range existing {
		existing[i] = float64(i%7) / 10.0
	}
	for n := 0; n < b.N; n++ {
		vec, _ := buildEmbedding(
			existing,
			"Acme",
			"Footwear",
			"Lightweight shoe for training "+uuid.NewString(),
		)
		if len(vec) == EmbeddingDims() {
			copy(existing, vec)
		}
	}
}
