package rules

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"recsys/internal/types"
)

type stubStore struct {
	rules []types.Rule
	calls int
}

func (s *stubStore) ListActiveRulesForScope(ctx context.Context, orgID uuid.UUID, namespace, surface, segmentID string, ts time.Time) ([]types.Rule, error) {
	s.calls++
	return append([]types.Rule(nil), s.rules...), nil
}

func TestManagerCachingAndInvalidate(t *testing.T) {
	store := &stubStore{
		rules: []types.Rule{
			{RuleID: uuid.New(), Action: types.RuleActionBlock, TargetType: types.RuleTargetItem, ItemIDs: []string{"a"}},
		},
	}
	mgr := NewManager(store, ManagerOptions{
		RefreshInterval: time.Minute,
		MaxPinSlots:     3,
		Enabled:         true,
	})

	req := EvaluateRequest{
		OrgID:     uuid.New(),
		Namespace: "default",
		Surface:   "home",
		Now:       time.Now(),
	}

	if _, err := mgr.Evaluate(context.Background(), req); err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if store.calls != 1 {
		t.Fatalf("expected 1 store call, got %d", store.calls)
	}

	// Second call within refresh interval should hit cache.
	req.Now = req.Now.Add(500 * time.Millisecond)
	if _, err := mgr.Evaluate(context.Background(), req); err != nil {
		t.Fatalf("evaluate second: %v", err)
	}
	if store.calls != 1 {
		t.Fatalf("expected cached result, store calls=%d", store.calls)
	}

	// Invalidate scope; subsequent evaluate should fetch again.
	mgr.Invalidate("default", "home")
	req.Now = req.Now.Add(2 * time.Second)
	if _, err := mgr.Evaluate(context.Background(), req); err != nil {
		t.Fatalf("evaluate post-invalidate: %v", err)
	}
	if store.calls != 2 {
		t.Fatalf("expected cache refresh after invalidate, store calls=%d", store.calls)
	}
}
