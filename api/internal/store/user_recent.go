package store

import (
	"context"
	"time"

	_ "embed"

	"github.com/google/uuid"
)

//go:embed queries/user_recent_items.sql
var userRecentItemsSQL string

// ListUserRecentItemIDs returns distinct recent items for a user since "since".
// Items are ordered by most recent interaction time and capped by limit.
func (s *Store) ListUserRecentItemIDs(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	userID string,
	since time.Time,
	limit int,
) ([]string, error) {
	if userID == "" || limit <= 0 {
		return []string{}, nil
	}
	rows, err := s.Pool.Query(ctx, userRecentItemsSQL, orgID, ns, userID, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
