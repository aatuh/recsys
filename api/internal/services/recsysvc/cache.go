package recsysvc

import (
	"context"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/api/internal/cache"
	appmetrics "github.com/aatuh/recsys-suite/api/internal/metrics"
)

type tenantKey struct {
	Tenant  string
	Surface string
}

// CachedConfigStore adds TTL caching to a ConfigStore.
type CachedConfigStore struct {
	store ConfigStore
	cache *cache.TTLCache[tenantKey, TenantConfig]
	ttl   time.Duration
}

// NewCachedConfigStore wraps a ConfigStore with TTL caching.
func NewCachedConfigStore(store ConfigStore, ttl time.Duration) ConfigStore {
	if store == nil {
		return nil
	}
	if ttl <= 0 {
		return store
	}
	return &CachedConfigStore{
		store: store,
		cache: cache.NewTTL[tenantKey, TenantConfig](nil),
		ttl:   ttl,
	}
}

// GetConfig returns cached config or loads from the store.
func (c *CachedConfigStore) GetConfig(ctx context.Context, tenantID, surface string) (TenantConfig, error) {
	if c == nil || c.store == nil {
		return TenantConfig{}, ErrConfigNotFound
	}
	key := tenantKey{Tenant: strings.TrimSpace(tenantID), Surface: strings.TrimSpace(surface)}
	if val, ok := c.cache.Get(key); ok {
		appmetrics.RecordCacheResult("tenant_config", true)
		return val, nil
	}
	appmetrics.RecordCacheResult("tenant_config", false)
	val, err := c.store.GetConfig(ctx, tenantID, surface)
	if err != nil {
		return TenantConfig{}, err
	}
	c.cache.Set(key, val, c.ttl)
	return val, nil
}

// Invalidate removes cached entries for a tenant/surface.
func (c *CachedConfigStore) Invalidate(tenantID, surface string) int {
	if c == nil {
		return 0
	}
	tenantID = strings.TrimSpace(tenantID)
	surface = strings.TrimSpace(surface)
	return c.cache.Invalidate(func(k tenantKey, _ TenantConfig) bool {
		if tenantID != "" && k.Tenant != tenantID {
			return false
		}
		if surface != "" && k.Surface != surface {
			return false
		}
		return true
	})
}

// Stats exposes cache hit/miss counts.
func (c *CachedConfigStore) Stats() cache.Stats {
	if c == nil {
		return cache.Stats{}
	}
	return c.cache.Stats()
}

// CachedRulesStore adds TTL caching to a RulesStore.
type CachedRulesStore struct {
	store RulesStore
	cache *cache.TTLCache[tenantKey, TenantRules]
	ttl   time.Duration
}

// NewCachedRulesStore wraps a RulesStore with TTL caching.
func NewCachedRulesStore(store RulesStore, ttl time.Duration) RulesStore {
	if store == nil {
		return nil
	}
	if ttl <= 0 {
		return store
	}
	return &CachedRulesStore{
		store: store,
		cache: cache.NewTTL[tenantKey, TenantRules](nil),
		ttl:   ttl,
	}
}

// GetRules returns cached rules or loads from the store.
func (c *CachedRulesStore) GetRules(ctx context.Context, tenantID, surface string) (TenantRules, error) {
	if c == nil || c.store == nil {
		return TenantRules{}, ErrRulesNotFound
	}
	key := tenantKey{Tenant: strings.TrimSpace(tenantID), Surface: strings.TrimSpace(surface)}
	if val, ok := c.cache.Get(key); ok {
		appmetrics.RecordCacheResult("tenant_rules", true)
		return val, nil
	}
	appmetrics.RecordCacheResult("tenant_rules", false)
	val, err := c.store.GetRules(ctx, tenantID, surface)
	if err != nil {
		return TenantRules{}, err
	}
	c.cache.Set(key, val, c.ttl)
	return val, nil
}

// Invalidate removes cached entries for a tenant/surface.
func (c *CachedRulesStore) Invalidate(tenantID, surface string) int {
	if c == nil {
		return 0
	}
	tenantID = strings.TrimSpace(tenantID)
	surface = strings.TrimSpace(surface)
	return c.cache.Invalidate(func(k tenantKey, _ TenantRules) bool {
		if tenantID != "" && k.Tenant != tenantID {
			return false
		}
		if surface != "" && k.Surface != surface {
			return false
		}
		return true
	})
}

// Stats exposes cache hit/miss counts.
func (c *CachedRulesStore) Stats() cache.Stats {
	if c == nil {
		return cache.Stats{}
	}
	return c.cache.Stats()
}
