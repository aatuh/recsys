package algorithm

import (
	"context"
	"testing"
	"time"

	recmodel "github.com/aatuh/recsys-algo/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type deterministicStore struct {
	pop           []recmodel.ScoredItem
	anchors       []string
	capturedSince time.Time
}

func (s *deterministicStore) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int, c *recmodel.PopConstraints) ([]recmodel.ScoredItem, error) {
	limit := k
	if limit <= 0 || limit > len(s.pop) {
		limit = len(s.pop)
	}
	out := make([]recmodel.ScoredItem, limit)
	copy(out, s.pop[:limit])
	return out, nil
}

func (s *deterministicStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]recmodel.ItemTags, error) {
	return map[string]recmodel.ItemTags{}, nil
}

func (s *deterministicStore) ListItemsAvailability(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]bool, error) {
	out := make(map[string]bool, len(itemIDs))
	for _, id := range itemIDs {
		if id != "" {
			out[id] = true
		}
	}
	return out, nil
}

func (s *deterministicStore) ListUserEventsSince(ctx context.Context, orgID uuid.UUID, ns string, userID string, since time.Time, eventTypes []int16) ([]string, error) {
	return nil, nil
}

func (s *deterministicStore) ListUserRecentItemIDs(ctx context.Context, orgID uuid.UUID, ns string, userID string, since time.Time, limit int) ([]string, error) {
	s.capturedSince = since
	size := len(s.anchors)
	if limit > 0 && limit < size {
		size = limit
	}
	out := make([]string, size)
	copy(out, s.anchors[:size])
	return out, nil
}

func (s *deterministicStore) CooccurrenceTopKWithin(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int, since time.Time) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (s *deterministicStore) SimilarByEmbeddingTopK(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (s *deterministicStore) CollaborativeTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, k int, excludeIDs []string) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (s *deterministicStore) ContentSimilarityTopK(ctx context.Context, orgID uuid.UUID, ns string, tags []string, k int, excludeIDs []string) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (s *deterministicStore) SessionSequenceTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, lookback int, horizonMinutes float64, excludeIDs []string, k int) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (s *deterministicStore) BuildUserTagProfile(ctx context.Context, orgID uuid.UUID, ns string, userID string, windowDays float64, topN int) (map[string]float64, error) {
	return map[string]float64{}, nil
}

func TestBuildResponseTieBreaksByItemID(t *testing.T) {
	eng := NewEngine(Config{}, nil, nil)
	data := &CandidateData{
		Candidates: []recmodel.ScoredItem{
			{ItemID: "b", Score: 1},
			{ItemID: "a", Score: 1},
		},
	}

	resp := eng.buildResponse(data, 2, "test", false, ExplainLevelTags, BlendWeights{Pop: 1}, nil, nil)
	require.Len(t, resp.Items, 2)
	require.Equal(t, "a", resp.Items[0].ItemID)
	require.Equal(t, "b", resp.Items[1].ItemID)
}

func TestRecommendDeterministicOutput(t *testing.T) {
	now := time.Date(2024, 3, 12, 10, 0, 0, 0, time.UTC)
	store := &deterministicStore{
		pop: []recmodel.ScoredItem{
			{ItemID: "b", Score: 1},
			{ItemID: "a", Score: 1},
			{ItemID: "c", Score: 0.5},
		},
		anchors: []string{"anchor"},
	}
	cfg := Config{
		BlendAlpha:      1,
		BlendBeta:       0,
		BlendGamma:      0,
		CoVisWindowDays: 7,
	}
	eng := NewEngine(cfg, store, nil, WithClock(fixedClock{now: now}))

	req := Request{
		OrgID:        uuid.New(),
		Namespace:    "default",
		UserID:       "user",
		K:            2,
		ExplainLevel: ExplainLevelTags,
	}

	resp1, _, err := eng.Recommend(context.Background(), req)
	require.NoError(t, err)
	resp2, _, err := eng.Recommend(context.Background(), req)
	require.NoError(t, err)

	require.Equal(t, resp1.Items, resp2.Items)
	require.Len(t, resp1.Items, 2)
	require.Equal(t, "a", resp1.Items[0].ItemID)
	require.Equal(t, "b", resp1.Items[1].ItemID)

	expectedSince := now.Add(-7 * 24 * time.Hour)
	require.True(t, store.capturedSince.Equal(expectedSince), "expected since %s, got %s", expectedSince, store.capturedSince)
}

func TestRecommendSourceMetricsKeys(t *testing.T) {
	now := time.Date(2024, 3, 12, 10, 0, 0, 0, time.UTC)
	store := &deterministicStore{
		pop:     []recmodel.ScoredItem{{ItemID: "x", Score: 1}},
		anchors: []string{"anchor"},
	}
	cfg := Config{BlendAlpha: 1}
	eng := NewEngine(cfg, store, nil, WithClock(fixedClock{now: now}))

	req := Request{
		OrgID:        uuid.New(),
		Namespace:    "default",
		UserID:       "user",
		K:            1,
		ExplainLevel: ExplainLevelTags,
	}

	_, trace, err := eng.Recommend(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, trace)
	_, hasExclusion := trace.SourceMetrics["post_exclusion"]
	_, hasConstraints := trace.SourceMetrics["post_constraints"]
	require.True(t, hasExclusion)
	require.True(t, hasConstraints)
}
