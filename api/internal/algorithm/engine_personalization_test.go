package algorithm

import (
	"context"
	"testing"
	"time"

	"recsys/internal/types"

	"github.com/google/uuid"
)

type personalizationStore struct{}

func (personalizationStore) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int, c *types.PopConstraints) ([]types.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]types.ItemTags, error) {
	return nil, nil
}

func (personalizationStore) ListItemsAvailability(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]bool, error) {
	out := make(map[string]bool, len(itemIDs))
	for _, id := range itemIDs {
		out[id] = true
	}
	return out, nil
}

func (personalizationStore) ListUserEventsSince(ctx context.Context, orgID uuid.UUID, ns string, userID string, since time.Time, eventTypes []int16) ([]string, error) {
	return nil, nil
}

func (personalizationStore) ListUserRecentItemIDs(ctx context.Context, orgID uuid.UUID, ns string, userID string, since time.Time, limit int) ([]string, error) {
	return nil, nil
}

func (personalizationStore) CooccurrenceTopKWithin(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int, since time.Time) ([]types.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) SimilarByEmbeddingTopK(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int) ([]types.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) CollaborativeTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, k int, excludeIDs []string) ([]types.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) ContentSimilarityTopK(ctx context.Context, orgID uuid.UUID, ns string, tags []string, k int, excludeIDs []string) ([]types.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) SessionSequenceTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, lookback int, horizonMinutes float64, excludeIDs []string, k int) ([]types.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) BuildUserTagProfile(ctx context.Context, orgID uuid.UUID, ns string, userID string, windowDays float64, topN int) (map[string]float64, error) {
	return map[string]float64{"books": 1.0}, nil
}

func TestApplyPersonalizationBoostAttenuatesSparseHistory(t *testing.T) {
	cfg := Config{
		ProfileBoost:               0.7,
		ProfileWindowDays:          30,
		ProfileTopNTags:            8,
		ProfileMinEventsForBoost:   3,
		ProfileColdStartMultiplier: 0.4,
	}
	eng := NewEngine(cfg, personalizationStore{}, nil)

	data := &CandidateData{
		Candidates: []types.ScoredItem{{ItemID: "item_a", Score: 1.0}},
		Tags: map[string]types.ItemTags{
			"item_a": {ItemID: "item_a", Tags: []string{"books"}},
		},
		Boosted:           map[string]bool{},
		ProfileOverlap:    map[string]float64{},
		ProfileMultiplier: map[string]float64{},
		Anchors:           []string{"item_history"},
	}

	req := Request{OrgID: uuid.New(), UserID: "user", Namespace: "default"}

	eng.applyPersonalizationBoost(context.Background(), data, req)

	got := data.Candidates[0].Score
	want := 1.0 + (0.7 * 0.4)
	if diff := got - want; diff < -1e-6 || diff > 1e-6 {
		t.Fatalf("expected score %.3f, got %.3f", want, got)
	}
	if data.ProfileMultiplier["item_a"] != want {
		t.Fatalf("expected multiplier %.3f, got %.3f", want, data.ProfileMultiplier["item_a"])
	}
}
