package algorithm

import (
	"context"
	"time"

	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"
)

// SimilarItemsEngine handles similar items recommendations
type SimilarItemsEngine struct {
	store           any
	coVisWindowDays int
	clock           Clock
}

// NewSimilarItemsEngine creates a new similar items engine
func NewSimilarItemsEngine(store any, coVisWindowDays int) *SimilarItemsEngine {
	return &SimilarItemsEngine{
		store:           store,
		coVisWindowDays: coVisWindowDays,
		clock:           realClock{},
	}
}

// WithClock overrides the clock used for time-based behavior.
func (e *SimilarItemsEngine) WithClock(clock Clock) *SimilarItemsEngine {
	if clock != nil {
		e.clock = clock
	}
	return e
}

// FindSimilar finds similar items using embedding similarity or co-visitation
func (e *SimilarItemsEngine) FindSimilar(ctx context.Context, req SimilarItemsRequest) (*SimilarItemsResponse, error) {
	k := req.K
	if k <= 0 {
		k = 20
	}

	// Try embedding similarity first
	if embeddingStore, ok := e.store.(recmodel.EmbeddingStore); ok {
		embeddingItems, err := embeddingStore.SimilarByEmbeddingTopK(
			ctx,
			req.OrgID,
			req.Namespace,
			req.ItemID,
			k,
		)
		if err == nil && len(embeddingItems) > 0 {
			embeddingItems = e.filterAvailable(ctx, req, embeddingItems)
		}
		if err == nil && len(embeddingItems) > 0 {
			// Convert to response format
			items := make([]ScoredItem, 0, len(embeddingItems))
			for _, item := range embeddingItems {
				items = append(items, ScoredItem{
					ItemID:  item.ItemID,
					Score:   item.Score,
					Reasons: []string{"embedding_similarity"},
				})
			}
			return &SimilarItemsResponse{Items: items}, nil
		}
	}

	// Fall back to co-visitation
	coocStore, ok := e.store.(recmodel.CooccurrenceStore)
	if !ok {
		return nil, recmodel.ErrFeatureUnavailable
	}
	days := e.coVisWindowDays
	if days <= 0 {
		days = 30
	}
	window := time.Duration(days*24.0) * time.Hour
	since := e.clock.Now().Add(-window)

	coVisItems, err := coocStore.CooccurrenceTopKWithin(
		ctx,
		req.OrgID,
		req.Namespace,
		req.ItemID,
		k,
		since,
	)
	if err != nil {
		return nil, err
	}
	coVisItems = e.filterAvailable(ctx, req, coVisItems)

	// Convert to response format
	items := make([]ScoredItem, 0, len(coVisItems))
	for _, item := range coVisItems {
		items = append(items, ScoredItem{
			ItemID:  item.ItemID,
			Score:   item.Score,
			Reasons: []string{"co_visitation"},
		})
	}

	return &SimilarItemsResponse{Items: items}, nil
}

func (e *SimilarItemsEngine) filterAvailable(
	ctx context.Context,
	req SimilarItemsRequest,
	items []recmodel.ScoredItem,
) []recmodel.ScoredItem {
	if len(items) == 0 {
		return nil
	}
	store, ok := e.store.(recmodel.AvailabilityStore)
	if !ok {
		return items
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item.ItemID != "" {
			ids = append(ids, item.ItemID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	availability, err := store.ListItemsAvailability(ctx, req.OrgID, req.Namespace, ids)
	if err != nil {
		return items
	}
	filtered := make([]recmodel.ScoredItem, 0, len(items))
	for _, item := range items {
		if ok, present := availability[item.ItemID]; present && ok {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
