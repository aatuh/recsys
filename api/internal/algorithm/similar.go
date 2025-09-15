package algorithm

import (
	"context"
	"time"

	"recsys/internal/types"
)

// SimilarItemsEngine handles similar items recommendations
type SimilarItemsEngine struct {
	store           types.RecAlgoStore
	coVisWindowDays int
}

// NewSimilarItemsEngine creates a new similar items engine
func NewSimilarItemsEngine(store types.RecAlgoStore, coVisWindowDays int) *SimilarItemsEngine {
	return &SimilarItemsEngine{
		store:           store,
		coVisWindowDays: coVisWindowDays,
	}
}

// FindSimilar finds similar items using embedding similarity or co-visitation
func (e *SimilarItemsEngine) FindSimilar(ctx context.Context, req SimilarItemsRequest) (*SimilarItemsResponse, error) {
	k := req.K
	if k <= 0 {
		k = 20
	}

	// Try embedding similarity first
	embeddingItems, err := e.store.SimilarByEmbeddingTopK(
		ctx,
		req.OrgID,
		req.Namespace,
		req.ItemID,
		k,
	)
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

	// Fall back to co-visitation
	days := e.coVisWindowDays
	if days <= 0 {
		days = 30
	}
	window := time.Duration(days*24.0) * time.Hour
	since := time.Now().UTC().Add(-window)

	coVisItems, err := e.store.CooccurrenceTopKWithin(
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
