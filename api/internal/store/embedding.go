package store

import (
	"context"
	"errors"

	_ "embed"

	recmodel "github.com/aatuh/recsys-algo/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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
) ([]recmodel.ScoredItem, error) {
	if k <= 0 {
		k = 20
	}
	var out []recmodel.ScoredItem
	err := s.withRetry(ctx, func(ctx context.Context) error {
		rows, err := s.Pool.Query(ctx, embeddingSimilaritySQL, orgID, ns, itemID, k)
		if err != nil {
			return err
		}
		defer rows.Close()

		items := make([]recmodel.ScoredItem, 0, k)
		for rows.Next() {
			var it recmodel.ScoredItem
			if err := rows.Scan(&it.ItemID, &it.Score); err != nil {
				return err
			}
			items = append(items, it)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		out = items
		return nil
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil, recmodel.ErrFeatureUnavailable
		}
		return nil, err
	}
	return out, nil
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

	var out map[string][]float64
	err := s.withRetry(ctx, func(ctx context.Context) error {
		rows, err := s.Pool.Query(ctx, itemsEmbeddingsSQL, orgID, ns, ids)
		if err != nil {
			return err
		}
		defer rows.Close()

		res := make(map[string][]float64, len(ids))
		for rows.Next() {
			var id string
			var emb []float64
			if err := rows.Scan(&id, &emb); err != nil {
				return err
			}
			if len(emb) > 0 {
				cp := make([]float64, len(emb))
				copy(cp, emb)
				res[id] = cp
			}
		}
		if err := rows.Err(); err != nil {
			return err
		}
		out = res
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
