package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"recsys/internal/store"
)

type coverageSnapshot struct {
	TotalCatalog  int
	LongTailFlags []bool
}

type coverageTracker struct {
	store         *store.Store
	ttl           time.Duration
	hintThreshold float64

	mu         sync.RWMutex
	namespaces map[string]*namespaceCoverage
}

type namespaceCoverage struct {
	items    map[string]coverageItem
	total    int
	loadedAt time.Time
}

type coverageItem struct {
	longTail bool
}

func newCoverageTracker(st *store.Store, ttl time.Duration, hintThreshold float64) *coverageTracker {
	if st == nil {
		return nil
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if hintThreshold < 0 {
		hintThreshold = 0
	}
	return &coverageTracker{
		store:         st,
		ttl:           ttl,
		hintThreshold: hintThreshold,
		namespaces:    make(map[string]*namespaceCoverage),
	}
}

func (c *coverageTracker) applyConfig(ttl time.Duration, hintThreshold float64) {
	if c == nil {
		return
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if hintThreshold < 0 {
		hintThreshold = 0
	}
	c.mu.Lock()
	c.ttl = ttl
	c.hintThreshold = hintThreshold
	c.namespaces = make(map[string]*namespaceCoverage)
	c.mu.Unlock()
}

func (c *coverageTracker) snapshot(ctx context.Context, orgID uuid.UUID, namespace string, itemIDs []string) (coverageSnapshot, error) {
	snap := coverageSnapshot{
		LongTailFlags: make([]bool, len(itemIDs)),
	}
	if c == nil {
		return snap, nil
	}
	nsKey := normalizedNamespace(namespace)

	cov, err := c.getNamespace(ctx, orgID, nsKey)
	if err != nil {
		return snap, err
	}
	snap.TotalCatalog = cov.total

	missing := false
	for _, id := range itemIDs {
		if _, ok := cov.items[id]; !ok {
			missing = true
			break
		}
	}
	if missing {
		cov, err = c.reloadNamespace(ctx, orgID, nsKey)
		if err != nil {
			return snap, err
		}
		snap.TotalCatalog = cov.total
	}

	for idx, id := range itemIDs {
		if info, ok := cov.items[id]; ok {
			snap.LongTailFlags[idx] = info.longTail
		} else {
			snap.LongTailFlags[idx] = true
		}
	}
	return snap, nil
}

func (c *coverageTracker) getNamespace(ctx context.Context, orgID uuid.UUID, namespace string) (*namespaceCoverage, error) {
	c.mu.RLock()
	entry := c.namespaces[namespace]
	ttl := c.ttl
	c.mu.RUnlock()
	if entry != nil && time.Since(entry.loadedAt) < ttl {
		return entry, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	entry = c.namespaces[namespace]
	if entry != nil && time.Since(entry.loadedAt) < ttl {
		return entry, nil
	}

	loaded, err := c.loadNamespace(ctx, orgID, namespace)
	if err != nil {
		return entry, err
	}
	c.namespaces[namespace] = loaded
	return loaded, nil
}

func (c *coverageTracker) reloadNamespace(ctx context.Context, orgID uuid.UUID, namespace string) (*namespaceCoverage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	loaded, err := c.loadNamespace(ctx, orgID, namespace)
	if err != nil {
		return nil, err
	}
	c.namespaces[namespace] = loaded
	return loaded, nil
}

func (c *coverageTracker) loadNamespace(ctx context.Context, orgID uuid.UUID, namespace string) (*namespaceCoverage, error) {
	if c.store == nil {
		return nil, errors.New("store is nil")
	}
	items := make(map[string]coverageItem)
	const pageSize = 500
	offset := 0
	total := 0
	for {
		page, count, err := c.store.ListItems(ctx, orgID, namespace, pageSize, offset, map[string]interface{}{})
		if err != nil {
			return nil, fmt.Errorf("list items: %w", err)
		}
		total = count
		if len(page) == 0 {
			break
		}
		for _, row := range page {
			itemID, _ := row["item_id"].(string)
			if itemID == "" {
				continue
			}
			propsRaw, ok := row["props"]
			if !ok {
				// default to long tail when metadata missing
				items[itemID] = coverageItem{longTail: true}
				continue
			}
			hint := extractPopularityHint(propsRaw)
			items[itemID] = coverageItem{longTail: hint <= c.hintThreshold}
		}
		offset += len(page)
		if offset >= count {
			break
		}
	}

	return &namespaceCoverage{
		items:    items,
		total:    total,
		loadedAt: time.Now(),
	}, nil
}

func normalizedNamespace(namespace string) string {
	ns := strings.ToLower(strings.TrimSpace(namespace))
	if ns == "" {
		return "default"
	}
	return ns
}

func extractPopularityHint(raw interface{}) float64 {
	switch v := raw.(type) {
	case string:
		return parseHintFromJSON([]byte(v))
	case []byte:
		return parseHintFromJSON(v)
	case json.RawMessage:
		return parseHintFromJSON([]byte(v))
	case map[string]interface{}:
		if hint, ok := v["popularity_hint"]; ok {
			if f, ok := hint.(float64); ok {
				return f
			}
		}
	}
	return 0
}

func parseHintFromJSON(raw []byte) float64 {
	if len(raw) == 0 {
		return 0
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return 0
	}
	if hint, ok := decoded["popularity_hint"]; ok {
		if f, ok := hint.(float64); ok {
			return f
		}
	}
	return 0
}
