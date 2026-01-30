package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// Stats captures cache hit/miss counts.
type Stats struct {
	Hits   uint64
	Misses uint64
}

// HitRate returns the hit ratio in the range [0,1].
func (s Stats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total)
}

type entry[V any] struct {
	value     V
	expiresAt time.Time
}

// TTLCache is a simple in-memory TTL cache with hit/miss stats.
type TTLCache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]entry[V]
	clock func() time.Time
	hits  uint64
	miss  uint64
}

// NewTTL constructs a TTL cache with an optional clock.
func NewTTL[K comparable, V any](clock func() time.Time) *TTLCache[K, V] {
	if clock == nil {
		clock = time.Now
	}
	return &TTLCache[K, V]{
		items: make(map[K]entry[V]),
		clock: clock,
	}
}

// Get returns the cached value when present and not expired.
func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	var zero V
	if c == nil {
		return zero, false
	}
	now := c.clock()
	c.mu.RLock()
	ent, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		atomic.AddUint64(&c.miss, 1)
		return zero, false
	}
	if !ent.expiresAt.IsZero() && now.After(ent.expiresAt) {
		atomic.AddUint64(&c.miss, 1)
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return zero, false
	}
	atomic.AddUint64(&c.hits, 1)
	return ent.value, true
}

// Set stores a value with TTL. ttl <= 0 removes the key.
func (c *TTLCache[K, V]) Set(key K, value V, ttl time.Duration) {
	if c == nil {
		return
	}
	if ttl <= 0 {
		c.Delete(key)
		return
	}
	exp := c.clock().Add(ttl)
	c.mu.Lock()
	c.items[key] = entry[V]{value: value, expiresAt: exp}
	c.mu.Unlock()
}

// Delete removes a key from the cache.
func (c *TTLCache[K, V]) Delete(key K) {
	if c == nil {
		return
	}
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Invalidate deletes entries matching the predicate and returns the count.
func (c *TTLCache[K, V]) Invalidate(match func(K, V) bool) int {
	if c == nil || match == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for k, v := range c.items {
		if match(k, v.value) {
			delete(c.items, k)
			n++
		}
	}
	return n
}

// Stats returns current hit/miss counts.
func (c *TTLCache[K, V]) Stats() Stats {
	if c == nil {
		return Stats{}
	}
	return Stats{
		Hits:   atomic.LoadUint64(&c.hits),
		Misses: atomic.LoadUint64(&c.miss),
	}
}
