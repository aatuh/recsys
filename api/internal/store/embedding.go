package store

import (
	"context"
	"recsys/internal/types"

	_ "embed"

	"github.com/google/uuid"
)

//go:embed queries/embedding_similarity.sql
var embeddingSimilaritySQL string

//go:embed queries/items_embeddings.sql
var itemsEmbeddingsSQL string

// SimilarByEmbeddingTopK returns k nearest neighbors by cosine distance
// from the anchor item's embedding. Requires both anchor and neighbor
// embeddings to be present. Orders by smallest distance; score is
// converted to similarity in [0,1] as (1 - distance).
func (s *Store) SimilarByEmbeddingTopK(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	itemID string,
	k int,
) ([]types.ScoredItem, error) {
	if k <= 0 {
		k = 20
	}
	rows, err := s.Pool.Query(ctx, embeddingSimilaritySQL, orgID, ns, itemID, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]types.ScoredItem, 0, k)
	for rows.Next() {
		var it types.ScoredItem
		if err := rows.Scan(&it.ItemID, &it.Score); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// ListItemsEmbeddings returns item_id -> embedding vector for given IDs.
// Missing or NULL embeddings are omitted from the map.
func (s *Store) ListItemsEmbeddings(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	ids []string,
) (map[string][]float64, error) {
	if len(ids) == 0 {
		return map[string][]float64{}, nil
	}

	rows, err := s.Pool.Query(ctx, itemsEmbeddingsSQL, orgID, ns, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string][]float64, len(ids))
	for rows.Next() {
		var id string
		var emb []float64
		if err := rows.Scan(&id, &emb); err != nil {
			return nil, err
		}
		if len(emb) > 0 {
			// Copy to avoid aliasing the driver buffer.
			cp := make([]float64, len(emb))
			copy(cp, emb)
			out[id] = cp
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
