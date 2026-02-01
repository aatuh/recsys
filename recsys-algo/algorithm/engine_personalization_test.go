package algorithm

import (
	"context"
	"testing"
	"time"

	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"

	"github.com/google/uuid"
)

type personalizationStore struct{}

func (personalizationStore) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int, c *recmodel.PopConstraints) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]recmodel.ItemTags, error) {
	return map[string]recmodel.ItemTags{}, nil
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

func (personalizationStore) CooccurrenceTopKWithin(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int, since time.Time) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) SimilarByEmbeddingTopK(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) CollaborativeTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, k int, excludeIDs []string) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) ContentSimilarityTopK(ctx context.Context, orgID uuid.UUID, ns string, tags []string, k int, excludeIDs []string) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) SessionSequenceTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, lookback int, horizonMinutes float64, excludeIDs []string, k int) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (personalizationStore) BuildUserTagProfile(ctx context.Context, orgID uuid.UUID, ns string, userID string, windowDays float64, topN int) (map[string]float64, error) {
	return map[string]float64{"books": 1.0}, nil
}

type personalizationCaseStore struct {
	personalizationStore
}

func (personalizationCaseStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]recmodel.ItemTags, error) {
	if len(itemIDs) == 0 {
		return map[string]recmodel.ItemTags{}, nil
	}
	return map[string]recmodel.ItemTags{
		itemIDs[0]: {ItemID: itemIDs[0], Tags: []string{"Books"}},
	}, nil
}

func (personalizationCaseStore) BuildUserTagProfile(ctx context.Context, orgID uuid.UUID, ns string, userID string, windowDays float64, topN int) (map[string]float64, error) {
	return map[string]float64{"BOOKS": 1.0}, nil
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
		Candidates: []recmodel.ScoredItem{{ItemID: "item_a", Score: 1.0}},
		Tags: map[string]recmodel.ItemTags{
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

func TestApplyPersonalizationBoostNormalizesTags(t *testing.T) {
	cfg := Config{
		ProfileBoost:      0.5,
		ProfileWindowDays: 30,
		ProfileTopNTags:   8,
	}
	eng := NewEngine(cfg, personalizationCaseStore{}, nil)

	req := Request{OrgID: uuid.New(), UserID: "user", Namespace: "default"}
	candidates := []recmodel.ScoredItem{{ItemID: "item_a", Score: 1.0}}
	tags, err := eng.getCandidateTags(context.Background(), candidates, req)
	if err != nil {
		t.Fatalf("getCandidateTags: %v", err)
	}

	data := &CandidateData{
		Candidates:        candidates,
		Tags:              tags,
		Boosted:           map[string]bool{},
		ProfileOverlap:    map[string]float64{},
		ProfileMultiplier: map[string]float64{},
	}

	eng.applyPersonalizationBoost(context.Background(), data, req)

	got := data.Candidates[0].Score
	want := 1.0 + 0.5
	if diff := got - want; diff < -1e-6 || diff > 1e-6 {
		t.Fatalf("expected score %.3f, got %.3f", want, got)
	}
	if data.ProfileOverlap["item_a"] != 1.0 {
		t.Fatalf("expected overlap 1.0, got %.3f", data.ProfileOverlap["item_a"])
	}
}
