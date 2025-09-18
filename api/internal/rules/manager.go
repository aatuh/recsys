package rules

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"recsys/internal/types"
)

// Store abstracts rule persistence calls used by the manager.
type Store interface {
	ListActiveRulesForScope(ctx context.Context, orgID uuid.UUID, namespace, surface, segmentID string, ts time.Time) ([]types.Rule, error)
}

// ManagerOptions configures rule manager behaviour.
type ManagerOptions struct {
	RefreshInterval time.Duration
	MaxPinSlots     int
	Enabled         bool
}

// Manager caches active rules and evaluates them for recommendation requests.
type Manager struct {
	store Store
	opts  ManagerOptions

	mu    sync.RWMutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	rules     []types.Rule
	fetchedAt time.Time
}

// NewManager constructs a manager with sane defaults.
func NewManager(store Store, opts ManagerOptions) *Manager {
	if opts.RefreshInterval <= 0 {
		opts.RefreshInterval = 2 * time.Second
	}
	if opts.MaxPinSlots <= 0 {
		opts.MaxPinSlots = 3
	}
	return &Manager{
		store: store,
		opts:  opts,
		cache: make(map[string]cacheEntry),
	}
}

// Enabled reports whether rules evaluation is active.
func (m *Manager) Enabled() bool {
	return m != nil && m.opts.Enabled
}

func cacheKey(namespace, surface, segmentID string) string {
	return namespace + "|" + surface + "|" + segmentID
}

// Invalidate removes cached rules for a namespace/surface pair (any segment).
func (m *Manager) Invalidate(namespace, surface string) {
	if m == nil {
		return
	}
	keyPrefix := namespace + "|" + surface + "|"
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.cache {
		if strings.HasPrefix(k, keyPrefix) {
			delete(m.cache, k)
		}
	}
}

func (m *Manager) loadRules(ctx context.Context, orgID uuid.UUID, namespace, surface, segmentID string, now time.Time) ([]types.Rule, error) {
	key := cacheKey(namespace, surface, segmentID)
	m.mu.RLock()
	entry, ok := m.cache[key]
	if ok && now.Sub(entry.fetchedAt) < m.opts.RefreshInterval {
		rulesCopy := make([]types.Rule, len(entry.rules))
		copy(rulesCopy, entry.rules)
		m.mu.RUnlock()
		return rulesCopy, nil
	}
	m.mu.RUnlock()

	rules, err := m.store.ListActiveRulesForScope(ctx, orgID, namespace, surface, segmentID, now)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.cache[key] = cacheEntry{rules: rules, fetchedAt: now}
	m.mu.Unlock()

	rulesCopy := make([]types.Rule, len(rules))
	copy(rulesCopy, rules)
	return rulesCopy, nil
}

// Evaluate fetches active rules (with caching) and applies them to the request.
func (m *Manager) Evaluate(ctx context.Context, req EvaluateRequest) (*EvaluateResult, error) {
	if m == nil || !m.opts.Enabled {
		return &EvaluateResult{
			Candidates:  append([]types.ScoredItem(nil), req.Candidates...),
			Pinned:      nil,
			Matches:     nil,
			ItemEffects: map[string]ItemEffect{},
			ReasonTags:  map[string][]string{},
		}, nil
	}

	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	rules, err := m.loadRules(ctx, req.OrgID, req.Namespace, req.Surface, req.SegmentID, now)
	if err != nil {
		return nil, err
	}

	eval := evaluator{
		maxPinSlots:         m.opts.MaxPinSlots,
		brandTagPrefixes:    req.BrandTagPrefixes,
		categoryTagPrefixes: req.CategoryTagPrefixes,
	}
	return eval.apply(rules, req)
}
