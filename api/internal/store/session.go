package store

import (
	"context"
	"errors"

	_ "embed"

	recmodel "github.com/aatuh/recsys-algo/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed queries/session_sequence_topk.sql
var sessionSequenceTopKSQL string

// SessionSequenceTopK returns items co-occurring after the most recent user events within a lookahead window.
func (s *Store) SessionSequenceTopK(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	userID string,
	lookback int,
	horizonMinutes float64,
	excludeIDs []string,
	k int,
) ([]recmodel.ScoredItem, error) {
	if userID == "" || k <= 0 || lookback <= 0 {
		return []recmodel.ScoredItem{}, nil
	}

	rows, err := s.Pool.Query(ctx, sessionSequenceTopKSQL, orgID, ns, userID, lookback, horizonMinutes, excludeIDs, k)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil, recmodel.ErrFeatureUnavailable
		}
		return nil, err
	}
	defer rows.Close()

	items := make([]recmodel.ScoredItem, 0, k)
	for rows.Next() {
		var it recmodel.ScoredItem
		if err := rows.Scan(&it.ItemID, &it.Score); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
