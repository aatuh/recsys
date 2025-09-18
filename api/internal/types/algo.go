package types

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PopConstraints struct {
	IncludeTagsAny []string
	MinPrice       *float64
	MaxPrice       *float64
	CreatedAfter   *time.Time
	ExcludeItemIDs []string
}

type ScoredItem struct {
	ItemID string
	Score  float64
}

// ItemTags holds tags required for diversity and caps.
type ItemTags struct {
	ItemID string
	Tags   []string
}

// RecAlgoStore is the minimal interface the algorithm needs from persistence.
// This keeps the algorithm independent of the concrete DB layer.
type RecAlgoStore interface {
	// Popularity

	// PopularityTopK returns up to k items with highest time-decayed
	// popularity in the org/namespace. Optional constraints filter items.
	// Results are ordered by score descending.
	PopularityTopK(
		ctx context.Context,
		orgID uuid.UUID,
		ns string,
		halfLifeDays float64,
		k int,
		c *PopConstraints,
	) ([]ScoredItem, error)

	// Metadata, filters and lookups

	// ListItemsTags returns minimal tags (e.g., tags) for item IDs as
	// a map item_id -> ItemTags. Missing items may be absent.
	ListItemsTags(
		ctx context.Context,
		orgID uuid.UUID,
		ns string,
		itemIDs []string,
	) (map[string]ItemTags, error)

	// ListUserEventsSince returns distinct item IDs for the user's events
	// on/after the timestamp filtered by event types. Order is not guaranteed.
	ListUserEventsSince(
		ctx context.Context,
		orgID uuid.UUID,
		ns string,
		userID string,
		since time.Time,
		eventTypes []int16,
	) ([]string, error)

	// User history

	// ListUserRecentItemIDs returns distinct recent item IDs for the user
	// since the cutoff, ordered by most recent interaction, limited by
	// the given limit.
	ListUserRecentItemIDs(
		ctx context.Context,
		orgID uuid.UUID,
		ns string,
		userID string,
		since time.Time,
		limit int,
	) ([]string, error)

	// Similarities

	// CooccurrenceTopKWithin returns up to k items that co-occurred most
	// with the anchor item since the cutoff. Anchor is excluded. Ordered
	// by co-occurrence score desc.
	CooccurrenceTopKWithin(
		ctx context.Context,
		orgID uuid.UUID,
		ns string,
		anchor string,
		k int,
		since time.Time,
	) ([]ScoredItem, error)

	// SimilarByEmbeddingTopK returns up to k nearest neighbors by item
	// embedding to the anchor. Requires embeddings. Ordered by similarity;
	// score is typically 1 - distance in [0,1].
	SimilarByEmbeddingTopK(
		ctx context.Context,
		orgID uuid.UUID,
		ns string,
		anchor string,
		k int,
	) ([]ScoredItem, error)

	// Personalization

	// BuildUserTagProfile computes a decayed, weighted tag-preference map
	// from the user's events, optionally limited by a time window. Returns
	// topN tags normalized to sum to 1.
	BuildUserTagProfile(
		ctx context.Context,
		orgID uuid.UUID,
		ns string,
		userID string,
		windowDays float64,
		topN int,
	) (map[string]float64, error)
}
