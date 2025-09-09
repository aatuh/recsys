package store

import (
	"context"

	"github.com/google/uuid"
)

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
) ([]ScoredItem, error) {
	if k <= 0 {
		k = 20
	}
	rows, err := s.Pool.Query(ctx, `
WITH anchor AS (
  SELECT embedding
  FROM items
  WHERE org_id=$1 AND namespace=$2 AND item_id=$3 AND embedding IS NOT NULL
)
SELECT i.item_id,
       (1.0 - (a.embedding <=> i.embedding)) AS score
FROM anchor a
JOIN items i
  ON i.org_id=$1 AND i.namespace=$2
WHERE i.item_id <> $3
  AND i.available = true
  AND i.embedding IS NOT NULL
ORDER BY a.embedding <=> i.embedding ASC
LIMIT $4
`, orgID, ns, itemID, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ScoredItem, 0, k)
	for rows.Next() {
		var it ScoredItem
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

	rows, err := s.Pool.Query(ctx, `
SELECT item_id, embedding
FROM items
WHERE org_id = $1 AND namespace = $2
  AND item_id = ANY($3)
  AND embedding IS NOT NULL
`, orgID, ns, ids)
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
