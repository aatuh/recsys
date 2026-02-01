package rules

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubStore struct {
	rules []Rule
	calls int
}

func (s *stubStore) ListActiveRulesForScope(ctx context.Context, orgID uuid.UUID, namespace, surface, segmentID string, ts time.Time) ([]Rule, error) {
	s.calls++
	return append([]Rule(nil), s.rules...), nil
}

func TestManagerCachingAndInvalidate(t *testing.T) {
	store := &stubStore{
		rules: []Rule{
			{RuleID: uuid.New(), Action: RuleActionBlock, TargetType: RuleTargetItem, ItemIDs: []string{"a"}},
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
	mgr.Invalidate(req.OrgID, "default", "home")
	req.Now = req.Now.Add(2 * time.Second)
	if _, err := mgr.Evaluate(context.Background(), req); err != nil {
		t.Fatalf("evaluate post-invalidate: %v", err)
	}
	if store.calls != 2 {
		t.Fatalf("expected cache refresh after invalidate, store calls=%d", store.calls)
	}
}

func TestManagerCacheKeyIncludesOrgID(t *testing.T) {
	store := &stubStore{
		rules: []Rule{
			{RuleID: uuid.New(), Action: RuleActionBlock, TargetType: RuleTargetItem, ItemIDs: []string{"a"}},
		},
	}
	mgr := NewManager(store, ManagerOptions{
		RefreshInterval: time.Minute,
		MaxPinSlots:     3,
		Enabled:         true,
	})

	orgA := uuid.New()
	orgB := uuid.New()
	now := time.Now()

	reqA := EvaluateRequest{
		OrgID:     orgA,
		Namespace: "default",
		Surface:   "home",
		Now:       now,
	}
	reqB := EvaluateRequest{
		OrgID:     orgB,
		Namespace: "default",
		Surface:   "home",
		Now:       now,
	}

	if _, err := mgr.Evaluate(context.Background(), reqA); err != nil {
		t.Fatalf("evaluate orgA: %v", err)
	}
	if store.calls != 1 {
		t.Fatalf("expected 1 store call, got %d", store.calls)
	}
	if _, err := mgr.Evaluate(context.Background(), reqB); err != nil {
		t.Fatalf("evaluate orgB: %v", err)
	}
	if store.calls != 2 {
		t.Fatalf("expected 2 store calls for distinct orgs, got %d", store.calls)
	}
	if _, err := mgr.Evaluate(context.Background(), reqA); err != nil {
		t.Fatalf("evaluate orgA cached: %v", err)
	}
	if store.calls != 2 {
		t.Fatalf("expected cached orgA result, store calls=%d", store.calls)
	}
}

func TestManagerInvalidateNamespaceWildcard(t *testing.T) {
	store := &stubStore{
		rules: []Rule{
			{RuleID: uuid.New(), Action: RuleActionBlock, TargetType: RuleTargetItem, ItemIDs: []string{"a"}},
		},
	}
	mgr := NewManager(store, ManagerOptions{
		RefreshInterval: time.Minute,
		MaxPinSlots:     3,
		Enabled:         true,
	})

	org := uuid.New()
	now := time.Now()

	reqHome := EvaluateRequest{
		OrgID:     org,
		Namespace: "default",
		Surface:   "home",
		Now:       now,
	}
	reqPdp := EvaluateRequest{
		OrgID:     org,
		Namespace: "default",
		Surface:   "pdp",
		Now:       now,
	}

	if _, err := mgr.Evaluate(context.Background(), reqHome); err != nil {
		t.Fatalf("evaluate home: %v", err)
	}
	if _, err := mgr.Evaluate(context.Background(), reqPdp); err != nil {
		t.Fatalf("evaluate pdp: %v", err)
	}
	if store.calls != 2 {
		t.Fatalf("expected 2 store calls, got %d", store.calls)
	}

	// Invalidate all surfaces under the namespace.
	mgr.Invalidate(org, "default", "")
	reqHome.Now = now.Add(2 * time.Second)
	if _, err := mgr.Evaluate(context.Background(), reqHome); err != nil {
		t.Fatalf("evaluate after invalidate: %v", err)
	}
	if store.calls != 3 {
		t.Fatalf("expected cache refresh after invalidate, store calls=%d", store.calls)
	}
}
