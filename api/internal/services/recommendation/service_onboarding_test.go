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
	cfg := algorithm.Config{ProfileTopNTags: 8, ProfileWindowDays: 30}
	req := algorithm.Request{
		OrgID:     uuid.New(),
		UserID:    "newbie",
		Namespace: "default",
	}
	selection := SegmentSelection{UserTraits: map[string]any{"segment": "new_users"}}

	starter := svc.buildStarterProfile(context.Background(), cfg, req, selection)
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
	if _, ok := starter["electronics"]; !ok {
		t.Fatalf("expected electronics tag in starter profile")
	}
}

func TestBuildStarterProfileSkipsWhenHistory(t *testing.T) {
	svc := &Service{store: onboardingStoreStub{recent: []string{"item_a"}}}
	cfg := algorithm.Config{ProfileTopNTags: 8, ProfileWindowDays: 30}
	req := algorithm.Request{
		OrgID:     uuid.New(),
		UserID:    "existing",
		Namespace: "default",
	}
	selection := SegmentSelection{UserTraits: map[string]any{"segment": "new_users"}}

	starter := svc.buildStarterProfile(context.Background(), cfg, req, selection)
	if starter != nil {
		t.Fatalf("expected no starter profile when history exists")
	}
}

func TestStarterProfileUnknownSegment(t *testing.T) {
	svc := &Service{store: onboardingStoreStub{}}
	cfg := algorithm.Config{ProfileTopNTags: 8, ProfileWindowDays: 30}
	req := algorithm.Request{
		OrgID:     uuid.New(),
		UserID:    "newbie",
		Namespace: "default",
	}
	selection := SegmentSelection{UserTraits: map[string]any{"segment": "unknown"}}

	starter := svc.buildStarterProfile(context.Background(), cfg, req, selection)
	if starter != nil {
		t.Fatalf("expected nil starter profile for unknown segment")
	}
}
