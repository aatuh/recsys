package store

import (
	"context"

	_ "embed"

	"recsys/internal/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

//go:embed queries/user_get.sql
var userGetSQL string

// GetUser returns a user record if it exists.
func (s *Store) GetUser(
	ctx context.Context,
	orgID uuid.UUID,
	ns, userID string,
) (*types.UserRecord, error) {
	row := s.Pool.QueryRow(ctx, userGetSQL, orgID, ns, userID)
	var rec types.UserRecord
	var traits map[string]any
	if err := row.Scan(&rec.UserID, &traits, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	rec.Traits = traits
	return &rec, nil
}
