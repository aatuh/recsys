package store

import (
	"context"
	"errors"
	"recsys/internal/types"

	_ "embed"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed queries/user_factor_topk.sql
var userFactorTopKSQL string

// CollaborativeTopK returns the top-N items for a user based on ALS factors.
// It excludes items already in the provided exclusion list.
func (s *Store) CollaborativeTopK(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	userID string,
	k int,
	excludeIDs []string,
) ([]types.ScoredItem, error) {
	if k <= 0 {
		return nil, errors.New("k must be positive")
	}

	var out []types.ScoredItem
	err := s.withRetry(ctx, func(ctx context.Context) error {
		rows, err := s.Pool.Query(ctx, userFactorTopKSQL, orgID, ns, userID, k, excludeIDs)
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
