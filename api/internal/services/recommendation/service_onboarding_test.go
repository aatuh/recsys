package recommendation

import (
	"context"
	"math"
	"testing"
	"time"

	"recsys/internal/algorithm"
	"recsys/internal/types"

	"github.com/google/uuid"
)

type onboardingStoreStub struct {
	recent []string
}

func (s onboardingStoreStub) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int, c *types.PopConstraints) ([]types.ScoredItem, error) {
	return nil, nil
}

func (s onboardingStoreStub) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]types.ItemTags, error) {
	return map[string]types.ItemTags{}, nil
}

func (s onboardingStoreStub) ListItemsAvailability(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]bool, error) {
	out := make(map[string]bool, len(itemIDs))
	for _, id := range itemIDs {
		out[id] = true
	}
	return out, nil
}

func (s onboardingStoreStub) ListUserEventsSince(ctx context.Context, orgID uuid.UUID, ns string, userID string, since time.Time, eventTypes []int16) ([]string, error) {
	return nil, nil
}

func (s onboardingStoreStub) ListUserRecentItemIDs(ctx context.Context, orgID uuid.UUID, ns string, userID string, since time.Time, limit int) ([]string, error) {
	if len(s.recent) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(s.recent))
	out = append(out, s.recent...)
	return out, nil
}

func (s onboardingStoreStub) CooccurrenceTopKWithin(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int, since time.Time) ([]types.ScoredItem, error) {
	return nil, nil
}

func (s onboardingStoreStub) SimilarByEmbeddingTopK(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int) ([]types.ScoredItem, error) {
	return nil, nil
}

func (s onboardingStoreStub) CollaborativeTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, k int, excludeIDs []string) ([]types.ScoredItem, error) {
	return nil, nil
}

func (s onboardingStoreStub) ContentSimilarityTopK(ctx context.Context, orgID uuid.UUID, ns string, tags []string, k int, excludeIDs []string) ([]types.ScoredItem, error) {
	return nil, nil
}

func (s onboardingStoreStub) SessionSequenceTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, lookback int, horizonMinutes float64, excludeIDs []string, k int) ([]types.ScoredItem, error) {
	return nil, nil
}

func (s onboardingStoreStub) BuildUserTagProfile(ctx context.Context, orgID uuid.UUID, ns string, userID string, windowDays float64, topN int) (map[string]float64, error) {
	return map[string]float64{}, nil
}

func TestBuildStarterProfileForNewUser(t *testing.T) {
	svc := &Service{store: onboardingStoreStub{}}
	cfg := algorithm.Config{
		ProfileTopNTags:           8,
		ProfileWindowDays:         30,
		ProfileMinEventsForBoost:  3,
		ProfileStarterBlendWeight: 0.6,
	}
	req := algorithm.Request{
		OrgID:     uuid.New(),
		UserID:    "newbie",
		Namespace: "default",
	}
	selection := SegmentSelection{UserTraits: map[string]any{"segment": "new_users"}}

	starter, weight := svc.buildStarterProfile(context.Background(), cfg, req, selection, 0, true, nil)
	if len(starter) == 0 {
		t.Fatalf("expected starter profile for new user")
	}
	sum := 0.0
	for _, v := range starter {
		sum += v
	}
	if math.Abs(sum-1.0) > 1e-6 {
		t.Fatalf("expected normalized weights, got sum=%f", sum)
	}
	if math.Abs(weight-1.0) > 1e-6 {
		t.Fatalf("expected starter blend weight to be 1 for zero history, got %f", weight)
	}
	if _, ok := starter["electronics"]; !ok {
		t.Fatalf("expected electronics tag in starter profile")
	}
}

func TestBuildStarterProfileSkipsWhenHistoryForNonNewSegment(t *testing.T) {
	svc := &Service{store: onboardingStoreStub{recent: []string{"item_a"}}}
	cfg := algorithm.Config{
		ProfileTopNTags:           8,
		ProfileWindowDays:         30,
		ProfileMinEventsForBoost:  3,
		ProfileStarterBlendWeight: 0.6,
	}
	req := algorithm.Request{
		OrgID:     uuid.New(),
		UserID:    "existing",
		Namespace: "default",
	}
	selection := SegmentSelection{UserTraits: map[string]any{"segment": "trend_seekers"}}

	starter, weight := svc.buildStarterProfile(context.Background(), cfg, req, selection, 4, true, []string{"item_a"})
	if starter != nil || weight != 0 {
		t.Fatalf("expected no starter profile when history exists for non-new segment")
	}
}

func TestStarterProfileFallbackToDefaultSegment(t *testing.T) {
	svc := &Service{store: onboardingStoreStub{}}
	cfg := algorithm.Config{
		ProfileTopNTags:           8,
		ProfileWindowDays:         30,
		ProfileMinEventsForBoost:  3,
		ProfileStarterBlendWeight: 0.6,
	}
	req := algorithm.Request{
		OrgID:     uuid.New(),
		UserID:    "newbie",
		Namespace: "default",
	}
	selection := SegmentSelection{UserTraits: map[string]any{"segment": "unknown"}}

	starter, weight := svc.buildStarterProfile(context.Background(), cfg, req, selection, 0, true, nil)
	if len(starter) == 0 {
		t.Fatalf("expected fallback starter profile for unknown segment")
	}
	if math.Abs(weight-1.0) > 1e-6 {
		t.Fatalf("expected full starter weight for unknown segment with zero history, got %f", weight)
	}
}

func TestBuildStarterProfileForNewUserEvenWithHistory(t *testing.T) {
	svc := &Service{store: onboardingStoreStub{recent: []string{"item_a", "item_b", "item_c", "item_d", "item_e"}}}
	cfg := algorithm.Config{
		ProfileTopNTags:           8,
		ProfileWindowDays:         30,
		ProfileMinEventsForBoost:  3,
		ProfileStarterBlendWeight: 0.6,
	}
	req := algorithm.Request{
		OrgID:     uuid.New(),
		UserID:    "newbie",
		Namespace: "default",
	}
	selection := SegmentSelection{UserTraits: map[string]any{"segment": "new_users"}}

	if profile := starterTagProfileForSegment("new_users"); len(profile) == 0 {
		t.Fatalf("expected preset for new_users")
	}

	starter, weight := svc.buildStarterProfile(context.Background(), cfg, req, selection, 5, true, []string{"item_a", "item_b"})
	if len(starter) == 0 {
		t.Fatalf("expected starter profile for new user despite recent history")
	}
	if weight <= 0 {
		t.Fatalf("expected positive starter blend weight when history exists, got %f", weight)
	}
}

func TestBuildStarterProfileForSparseHistoryWithoutSegment(t *testing.T) {
	svc := &Service{store: onboardingStoreStub{}}
	cfg := algorithm.Config{
		ProfileTopNTags:           8,
		ProfileWindowDays:         30,
		ProfileMinEventsForBoost:  3,
		ProfileStarterBlendWeight: 0.6,
	}
	req := algorithm.Request{
		OrgID:     uuid.New(),
		UserID:    "sparse",
		Namespace: "default",
	}

	starter, weight := svc.buildStarterProfile(context.Background(), cfg, req, SegmentSelection{}, 1, true, nil)
	if len(starter) == 0 {
		t.Fatalf("expected starter profile for sparse history user without explicit segment")
	}
	if weight <= cfg.ProfileStarterBlendWeight {
		t.Fatalf("expected starter weight boosted above base for sparse history, got %f", weight)
	}
}
