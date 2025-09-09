package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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
	rows, err := s.Pool.Query(ctx, `
SELECT item_id
FROM (
  SELECT item_id, MAX(ts) AS last_ts
  FROM events
  WHERE org_id = $1 AND namespace = $2
    AND user_id = $3
    AND ts >= $4
  GROUP BY item_id
) u
ORDER BY last_ts DESC
LIMIT $5
`, orgID, ns, userID, since, limit)
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
