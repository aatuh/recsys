package cache

import (
	"testing"
	"time"
)

func TestTTLCacheGetSetExpire(t *testing.T) {
	now := time.Date(2026, 1, 29, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	c := NewTTL[string, string](clock)

	c.Set("a", "value", 2*time.Second)
	if v, ok := c.Get("a"); !ok || v != "value" {
		t.Fatalf("expected cache hit")
	}

	now = now.Add(3 * time.Second)
	if _, ok := c.Get("a"); ok {
		t.Fatalf("expected cache miss after expiry")
	}
}

func TestTTLCacheStats(t *testing.T) {
	c := NewTTL[string, string](time.Now)
	c.Set("a", "value", time.Minute)
	_, _ = c.Get("a")
	_, _ = c.Get("b")

	stats := c.Stats()
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("expected 1 hit and 1 miss, got %+v", stats)
	}
}

func TestTTLCacheInvalidate(t *testing.T) {
	c := NewTTL[string, string](time.Now)
	c.Set("a", "value", time.Minute)
	c.Set("b", "value", time.Minute)

	removed := c.Invalidate(func(k, v string) bool {
		return k == "a"
	})
	if removed != 1 {
		t.Fatalf("expected 1 removed, got %d", removed)
	}
	if _, ok := c.Get("a"); ok {
		t.Fatalf("expected key a removed")
	}
	if _, ok := c.Get("b"); !ok {
		t.Fatalf("expected key b to remain")
	}
}
