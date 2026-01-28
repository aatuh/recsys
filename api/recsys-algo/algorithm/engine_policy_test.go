package algorithm

import (
	"context"
	"testing"
	"time"

	"github.com/aatuh/recsys-algo/rules"

	recmodel "github.com/aatuh/recsys-algo/model"

	"github.com/google/uuid"
)

type noopAlgoStore struct {
	recent []string
}

func (n *noopAlgoStore) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int, c *recmodel.PopConstraints) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]recmodel.ItemTags, error) {
	return map[string]recmodel.ItemTags{}, nil
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

func (n *noopAlgoStore) CooccurrenceTopKWithin(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int, since time.Time) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) SimilarByEmbeddingTopK(ctx context.Context, orgID uuid.UUID, ns string, anchor string, k int) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) CollaborativeTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, k int, excludeIDs []string) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) ContentSimilarityTopK(ctx context.Context, orgID uuid.UUID, ns string, tags []string, k int, excludeIDs []string) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) SessionSequenceTopK(ctx context.Context, orgID uuid.UUID, ns string, userID string, lookback int, horizonMinutes float64, excludeIDs []string, k int) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (n *noopAlgoStore) BuildUserTagProfile(ctx context.Context, orgID uuid.UUID, ns string, userID string, windowDays float64, topN int) (map[string]float64, error) {
	return map[string]float64{}, nil
}

func TestApplyExclusionsTracksExplicitAndRecent(t *testing.T) {
	store := &noopAlgoStore{recent: []string{"recent"}}
	eng := NewEngine(Config{RuleExcludeEvents: true, PurchasedWindowDays: 1}, store, nil)

	candidates := []recmodel.ScoredItem{{ItemID: "explicit"}, {ItemID: "recent"}, {ItemID: "ok"}}
	summary := PolicySummary{}
	req := Request{
		OrgID:       uuid.New(),
		Namespace:   "default",
		UserID:      "user",
		Constraints: &recmodel.PopConstraints{ExcludeItemIDs: []string{"explicit"}},
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
	candidates := []recmodel.ScoredItem{{ItemID: "keep"}, {ItemID: "drop"}}
	tags := map[string]recmodel.ItemTags{
		"keep": {ItemID: "keep", Tags: []string{"books"}},
		"drop": {ItemID: "drop", Tags: []string{"games"}},
	}
	summary := PolicySummary{}
	req := Request{Constraints: &recmodel.PopConstraints{IncludeTagsAny: []string{"Books", "books"}}}

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
	if got := summary.constraintFilteredReasons["drop"]; got != constraintReasonInclude {
		t.Fatalf("expected include reason recorded, got %q", got)
	}
}

func TestApplyConstraintFiltersPriceRange(t *testing.T) {
	eng := NewEngine(Config{}, nil, nil)
	min := 20.0
	max := 60.0
	now := time.Now()
	candidates := []recmodel.ScoredItem{
		{ItemID: "cheap"},
		{ItemID: "sweetspot"},
		{ItemID: "premium"},
		{ItemID: "unknown-price"},
	}
	tags := map[string]recmodel.ItemTags{
		"cheap":         {ItemID: "cheap", Tags: []string{"books"}, Price: floatPtr(10), CreatedAt: now},
		"sweetspot":     {ItemID: "sweetspot", Tags: []string{"books"}, Price: floatPtr(40), CreatedAt: now},
		"premium":       {ItemID: "premium", Tags: []string{"books"}, Price: floatPtr(80), CreatedAt: now},
		"unknown-price": {ItemID: "unknown-price", Tags: []string{"books"}, Price: nil, CreatedAt: now},
	}
	req := Request{Constraints: &recmodel.PopConstraints{MinPrice: &min, MaxPrice: &max}}

	filtered, _ := eng.applyConstraintFilters(candidates, tags, req, nil)
	if len(filtered) != 1 || filtered[0].ItemID != "sweetspot" {
		t.Fatalf("expected only sweetspot to remain, got %#v", filtered)
	}
}

func TestApplyConstraintFiltersCreatedAfter(t *testing.T) {
	eng := NewEngine(Config{}, nil, nil)
	now := time.Now().UTC()
	threshold := now.Add(-6 * time.Hour)
	candidates := []recmodel.ScoredItem{
		{ItemID: "fresh"},
		{ItemID: "stale"},
	}
	tags := map[string]recmodel.ItemTags{
		"fresh": {ItemID: "fresh", Price: floatPtr(25), CreatedAt: now},
		"stale": {ItemID: "stale", Price: floatPtr(25), CreatedAt: now.Add(-24 * time.Hour)},
	}
	req := Request{Constraints: &recmodel.PopConstraints{CreatedAfter: &threshold}}

	filtered, _ := eng.applyConstraintFilters(candidates, tags, req, nil)
	if len(filtered) != 1 || filtered[0].ItemID != "fresh" {
		t.Fatalf("expected only fresh item to remain, got %#v", filtered)
	}
}

func floatPtr(v float64) *float64 {
	return &v
}

func TestFinalizePolicySummaryDetectsLeaks(t *testing.T) {
	summary := PolicySummary{
		constraintFilteredLookup:  map[string]struct{}{"leaked": {}},
		constraintFilteredReasons: map[string]string{"leaked": constraintReasonInclude},
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
	if summary.ConstraintLeakByReason[constraintReasonInclude] != 1 {
		t.Fatalf("expected leak reason counted, got %#v", summary.ConstraintLeakByReason)
	}
	if summary.constraintFilteredLookup != nil {
		t.Fatalf("expected lookup cleared after finalize")
	}
	if summary.constraintFilteredReasons != nil {
		t.Fatalf("expected reason map cleared after finalize")
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

func TestFinalizePolicySummaryTracksRuleBlockExposureByRule(t *testing.T) {
	ruleA := uuid.New()
	ruleB := uuid.New()
	summary := PolicySummary{}
	ruleResult := &rules.EvaluateResult{
		ItemEffects: map[string]rules.ItemEffect{
			"a": {Blocked: true, BlockRules: []uuid.UUID{ruleA}},
			"b": {Blocked: true, BlockRules: []uuid.UUID{ruleA, ruleB}},
			"c": {Blocked: true},
			"d": {BoostDelta: 1},
		},
	}

	finalizePolicySummary(&summary, nil, ruleResult)

	if summary.RuleBlockExposure != 3 {
		t.Fatalf("expected block exposure 3, got %d", summary.RuleBlockExposure)
	}
	if summary.RuleBlockExposureByRule[ruleA.String()] != 2 {
		t.Fatalf("expected ruleA exposure 2, got %#v", summary.RuleBlockExposureByRule)
	}
	if summary.RuleBlockExposureByRule[ruleB.String()] != 1 {
		t.Fatalf("expected ruleB exposure 1, got %#v", summary.RuleBlockExposureByRule)
	}
	if summary.RuleBlockExposureByRule[constraintReasonUnknown] != 1 {
		t.Fatalf("expected unknown bucket for rule-less blocks, got %#v", summary.RuleBlockExposureByRule)
	}
}
