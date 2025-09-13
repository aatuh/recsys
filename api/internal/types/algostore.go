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

// AlgoStore is the minimal interface the algorithm needs from persistence.
// This keeps the algorithm independent of the concrete DB layer.
type AlgoStore interface {
	// Popularity

	// PopularityTopK returns up to k items with highest time-decayed
	// popularity in the org/namespace. Optional constraints filter items.
	// Results are ordered by score descending.
	PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string,
		halfLifeDays float64, k int, c *PopConstraints,
	) ([]ScoredItem, error)

	// Metadata, filters and lookups

	// ListItemsMeta returns minimal metadata (e.g., tags) for item IDs as
	// a map item_id -> ItemMeta. Missing items may be absent.
	ListItemsMeta(ctx context.Context, orgID uuid.UUID, ns string,
		itemIDs []string,
	) (map[string]ItemMeta, error)

	// ListUserPurchasedSince returns distinct item IDs the user purchased
	// on/after the timestamp. Order is not guaranteed.
	ListUserPurchasedSince(ctx context.Context, orgID uuid.UUID, ns string,
		userID string, since time.Time,
	) ([]string, error)

	// User history

	// ListUserRecentItemIDs returns distinct recent item IDs for the user
	// since the cutoff, ordered by most recent interaction, limited by
	// the given limit.
	ListUserRecentItemIDs(ctx context.Context, orgID uuid.UUID, ns string,
		userID string, since time.Time, limit int,
	) ([]string, error)

	// Similarities

	// CooccurrenceTopKWithin returns up to k items that co-occurred most
	// with the anchor item since the cutoff. Anchor is excluded. Ordered
	// by co-occurrence score desc.
	CooccurrenceTopKWithin(ctx context.Context, orgID uuid.UUID, ns string,
		anchor string, k int, since time.Time,
	) ([]ScoredItem, error)

	// SimilarByEmbeddingTopK returns up to k nearest neighbors by item
	// embedding to the anchor. Requires embeddings. Ordered by similarity;
	// score is typically 1 - distance in [0,1].
	SimilarByEmbeddingTopK(ctx context.Context, orgID uuid.UUID, ns string,
		anchor string, k int,
	) ([]ScoredItem, error)

	// Personalization

	// BuildUserTagProfile computes a decayed, weighted tag-preference map
	// from the user's events, optionally limited by a time window. Returns
	// topN tags normalized to sum to 1.
	BuildUserTagProfile(ctx context.Context, orgID uuid.UUID, ns string,
		userID string, windowDays float64, topN int,
	) (map[string]float64, error)
}
