package store

import (
	"context"
	"errors"
	"time"

	_ "embed"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed queries/catalog_candidates.sql
var catalogCandidatesSQL string

// CatalogQueryOptions controls selection of catalog rows for metadata refresh.
type CatalogQueryOptions struct {
	MissingOnly     bool
	UpdatedSince    *time.Time
	CursorUpdatedAt *time.Time
	CursorItemID    string
	Limit           int
}

// CatalogItem represents a row from the items table with metadata payloads.
type CatalogItem struct {
	ItemID          string
	Available       bool
	Price           *float64
	Tags            []string
	Props           []byte
	Embedding       []float64
	Brand           *string
	Category        *string
	CategoryPath    []string
	Description     *string
	ImageURL        *string
	MetadataVersion *string
	UpdatedAt       time.Time
}

// CatalogItems returns items requiring metadata backfill or refresh.
func (s *Store) CatalogItems(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	opts CatalogQueryOptions,
) ([]CatalogItem, error) {
	if s == nil {
		return nil, errors.New("store is nil")
	}
	if ns == "" {
		ns = "default"
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}

	cursorID := interface{}(nil)
	if opts.CursorItemID != "" {
		cursorID = opts.CursorItemID
	}

	var items []CatalogItem
	err := s.withRetry(ctx, func(ctx context.Context) error {
		rows, err := s.Pool.Query(
			ctx,
			catalogCandidatesSQL,
			orgID,
			ns,
			opts.MissingOnly,
			opts.UpdatedSince,
			opts.CursorUpdatedAt,
			cursorID,
			limit,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		result := make([]CatalogItem, 0, limit)
		for rows.Next() {
			var item CatalogItem
			if err := rows.Scan(
				&item.ItemID,
				&item.Available,
				&item.Price,
				&item.Tags,
				&item.Props,
				&item.Embedding,
				&item.Brand,
				&item.Category,
				&item.CategoryPath,
				&item.Description,
				&item.ImageURL,
				&item.MetadataVersion,
				&item.UpdatedAt,
			); err != nil {
				return err
			}
			result = append(result, item)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		items = result
		return nil
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return []CatalogItem{}, nil
		}
		return nil, err
	}
	return items, nil
}
