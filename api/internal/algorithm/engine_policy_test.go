package algorithm

import (
	"context"
	"testing"
	"time"

	"recsys/internal/rules"
	"recsys/internal/types"

	"github.com/google/uuid"
)

type noopAlgoStore struct {
	recent []string
}

func (n *noopAlgoStore) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int, c *types.PopConstraints) ([]types.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]types.ItemTags, error) {
	return nil, nil
}

func (n *noopAlgoStore) ListItemsAvailability(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]bool, error) {
	out := make(map[string]bool, len(itemIDs))
	for _, id := range itemIDs {
		out[id] = true
	}
	return out, nil
}

func (n *noopAlgoStore) ListUserEventsSince(ctx context.Context, orgID uuid.UUID, ns string, userID string, since time.Time, eventTypes []int16) ([]string, error) {
	if len(n.recent) == 0 {
		return nil, nil
	}
	out := make([]string, len(n.recent))
	copy(out, n.recent)
	return out, nil
}

func (n *noopAlgoStore) ListUserRecentItemIDs(ctx context.Context, orgID uuid.UUID, ns string, userID string, since time.Time, limit int) ([]string, error) {
	return nil, nil
}

func (n *noopAlgoStore) CooccurrenceTopKWithin(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int, since time.Time) ([]types.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) SimilarByEmbeddingTopK(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int) ([]types.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) CollaborativeTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, k int, excludeIDs []string) ([]types.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) ContentSimilarityTopK(ctx context.Context, orgID uuid.UUID, ns string, tags []string, k int, excludeIDs []string) ([]types.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) SessionSequenceTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, lookback int, horizonMinutes float64, excludeIDs []string, k int) ([]types.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) BuildUserTagProfile(ctx context.Context, orgID uuid.UUID, ns string, userID string, windowDays float64, topN int) (map[string]float64, error) {
	return nil, nil
}

func TestApplyExclusionsTracksExplicitAndRecent(t *testing.T) {
	store := &noopAlgoStore{recent: []string{"recent"}}
	eng := NewEngine(Config{RuleExcludeEvents: true, PurchasedWindowDays: 1}, store, nil)

	candidates := []types.ScoredItem{{ItemID: "explicit"}, {ItemID: "recent"}, {ItemID: "ok"}}
	summary := PolicySummary{}
	req := Request{
		OrgID:       uuid.New(),
		Namespace:   "default",
		UserID:      "user",
		Constraints: &types.PopConstraints{ExcludeItemIDs: []string{"explicit"}},
	}

	filtered, err := eng.applyExclusions(context.Background(), candidates, req, &summary)
	if err != nil {
		t.Fatalf("applyExclusions returned error: %v", err)
	}

	if len(filtered) != 1 || filtered[0].ItemID != "ok" {
		t.Fatalf("expected only ok candidate to remain, got %#v", filtered)
	}
	if summary.ExplicitExcludeHits != 1 {
		t.Fatalf("expected explicit exclude hit count 1, got %d", summary.ExplicitExcludeHits)
	}
	if summary.RecentEventExcludeHits != 1 {
		t.Fatalf("expected recent exclude hit count 1, got %d", summary.RecentEventExcludeHits)
	}
	if summary.AfterExclusions != 1 {
		t.Fatalf("expected AfterExclusions=1, got %d", summary.AfterExclusions)
	}
}

func TestApplyConstraintFiltersUpdatesSummary(t *testing.T) {
	eng := NewEngine(Config{}, nil, nil)
	candidates := []types.ScoredItem{{ItemID: "keep"}, {ItemID: "drop"}}
	tags := map[string]types.ItemTags{
		"keep": {ItemID: "keep", Tags: []string{"books"}},
		"drop": {ItemID: "drop", Tags: []string{"games"}},
	}
	summary := PolicySummary{}
	req := Request{Constraints: &types.PopConstraints{IncludeTagsAny: []string{"Books", "books"}}}

	filtered, _ := eng.applyConstraintFilters(candidates, tags, req, &summary)
	if len(filtered) != 1 || filtered[0].ItemID != "keep" {
		t.Fatalf("expected only keep candidate, got %#v", filtered)
	}
	if summary.ConstraintFilteredCount != 1 {
		t.Fatalf("expected filtered count 1, got %d", summary.ConstraintFilteredCount)
	}
	if summary.AfterConstraintFilters != 1 {
		t.Fatalf("expected AfterConstraintFilters=1, got %d", summary.AfterConstraintFilters)
	}
	if len(summary.ConstraintIncludeTags) != 1 || summary.ConstraintIncludeTags[0] != "books" {
		t.Fatalf("expected include tag 'books', got %#v", summary.ConstraintIncludeTags)
	}
	if summary.constraintFilteredLookup == nil || len(summary.constraintFilteredLookup) != 1 {
		t.Fatalf("expected lookup map populated")
	}
}

func TestFinalizePolicySummaryDetectsLeaks(t *testing.T) {
	summary := PolicySummary{
		constraintFilteredLookup: map[string]struct{}{"leaked": {}},
	}
	resp := &Response{Items: []ScoredItem{{ItemID: "leaked"}, {ItemID: "ok"}}}

	finalizePolicySummary(&summary, resp, nil)

	if summary.FinalCount != 2 {
		t.Fatalf("expected final count 2, got %d", summary.FinalCount)
	}
	if summary.ConstraintLeakCount != 1 {
		t.Fatalf("expected leak count 1, got %d", summary.ConstraintLeakCount)
	}
	if len(summary.ConstraintLeakIDs) != 1 || summary.ConstraintLeakIDs[0] != "leaked" {
		t.Fatalf("expected leaked ID recorded, got %#v", summary.ConstraintLeakIDs)
	}
	if summary.constraintFilteredLookup != nil {
		t.Fatalf("expected lookup cleared after finalize")
	}
}

func TestFinalizePolicySummaryTracksRuleExposure(t *testing.T) {
	summary := PolicySummary{}
	resp := &Response{Items: []ScoredItem{{ItemID: "boosted"}, {ItemID: "pinned"}, {ItemID: "plain"}}}
	ruleResult := &rules.EvaluateResult{
		ItemEffects: map[string]rules.ItemEffect{
			"boosted": {BoostDelta: 1.0},
			"pinned":  {Pinned: true},
			"other":   {BoostDelta: 2.0},
		},
	}

	finalizePolicySummary(&summary, resp, ruleResult)

	if summary.RuleBoostExposure != 1 {
		t.Fatalf("expected boost exposure 1, got %d", summary.RuleBoostExposure)
	}
	if summary.RulePinExposure != 1 {
		t.Fatalf("expected pin exposure 1, got %d", summary.RulePinExposure)
	}
}
