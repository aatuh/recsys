package types

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PopConstraints struct {
	IncludeTagsAny     []string
	MinPrice, MaxPrice *float64
	CreatedAfter       *time.Time
	ExcludeItemIDs     []string
}

type ScoredItem struct {
	ItemID string
	Score  float64
}

// ItemMeta holds lightweight metadata required for diversity and caps.
type ItemMeta struct {
	ItemID string
	Tags   []string
}

// Store is the minimal interface the algorithm needs from persistence.
// This keeps the algorithm independent of the concrete DB layer.
type Store interface {
	// Popularity
	PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string,
		halfLifeDays float64, k int, c *PopConstraints,
	) ([]ScoredItem, error)

	// Metadata, filters and lookups
	ListItemsMeta(ctx context.Context, orgID uuid.UUID, ns string,
		itemIDs []string,
	) (map[string]ItemMeta, error)
	ListUserPurchasedSince(ctx context.Context, orgID uuid.UUID, ns string,
		userID string, since time.Time,
	) ([]string, error)

	// User history
	ListUserRecentItemIDs(ctx context.Context, orgID uuid.UUID, ns string,
		userID string, since time.Time, limit int,
	) ([]string, error)

	// Similarities
	CooccurrenceTopKWithin(ctx context.Context, orgID uuid.UUID, ns string,
		anchor string, k int, since time.Time,
	) ([]ScoredItem, error)
	SimilarByEmbeddingTopK(ctx context.Context, orgID uuid.UUID, ns string,
		anchor string, k int,
	) ([]ScoredItem, error)

	// Personalization
	BuildUserTagProfile(ctx context.Context, orgID uuid.UUID, ns string,
		userID string, windowDays float64, topN int,
	) (map[string]float64, error)
}
