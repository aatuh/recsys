package store

import (
	"context"

	"github.com/google/uuid"
)

// ListItemsAvailability returns whether each item ID is currently available.
func (s *Store) ListItemsAvailability(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	itemIDs []string,
) (map[string]bool, error) {
	if len(itemIDs) == 0 {
		return map[string]bool{}, nil
	}

	rows, err := s.Pool.Query(ctx, `
		SELECT item_id, available
		FROM items
		WHERE org_id = $1
		  AND namespace = $2
		  AND item_id = ANY($3::text[])
	`, orgID, ns, itemIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]bool, len(itemIDs))
	for rows.Next() {
		var (
			id        string
			available bool
		)
		if err := rows.Scan(&id, &available); err != nil {
			return nil, err
		}
		out[id] = available
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
