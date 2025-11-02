package store

import (
	"context"
	"errors"
	"recsys/internal/types"

	_ "embed"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed queries/content_similarity_topk.sql
var contentSimilarityTopKSQL string

// ContentSimilarityTopK returns items ranked by tag overlap with the supplied tag set.
func (s *Store) ContentSimilarityTopK(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	tags []string,
	k int,
	excludeIDs []string,
) ([]types.ScoredItem, error) {
	if len(tags) == 0 || k <= 0 {
		return []types.ScoredItem{}, nil
	}

	var out []types.ScoredItem
	err := s.withRetry(ctx, func(ctx context.Context) error {
		rows, err := s.Pool.Query(ctx, contentSimilarityTopKSQL, orgID, ns, k, tags, excludeIDs)
		if err != nil {
			return err
		}
		defer rows.Close()

		items := make([]types.ScoredItem, 0, k)
		for rows.Next() {
			var it types.ScoredItem
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
			return []types.ScoredItem{}, nil
		}
		return nil, err
	}
	return out, nil
}
