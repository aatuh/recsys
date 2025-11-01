package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

type mapSource map[string]string

func (m mapSource) Lookup(key string) (string, bool) {
	v, ok := m[key]
	return v, ok
}

func baseConfigSource() mapSource {
	orgID := uuid.New().String()
	return mapSource{
		"API_PORT":                       "8000",
		"DATABASE_URL":                   "postgres://user:pass@localhost:5432/recsys?sslmode=disable",
		"ORG_ID":                         orgID,
		"AUDIT_DECISIONS_QUEUE":          "1024",
		"AUDIT_DECISIONS_BATCH":          "200",
		"AUDIT_DECISIONS_FLUSH_INTERVAL": "250ms",
		"AUDIT_DECISIONS_SAMPLE_DEFAULT": "1",
		"AUDIT_DECISIONS_SALT":           "test-salt",
		"POPULARITY_HALFLIFE_DAYS":       "7",
		"COVIS_WINDOW_DAYS":              "14",
		"POPULARITY_FANOUT":              "100",
		"MMR_LAMBDA":                     "0.5",
		"BRAND_CAP":                      "2",
		"CATEGORY_CAP":                   "3",
		"BRAND_TAG_PREFIXES":             "brand",
		"CATEGORY_TAG_PREFIXES":          "category",
		"PURCHASED_WINDOW_DAYS":          "30",
		"PROFILE_WINDOW_DAYS":            "30",
		"PROFILE_BOOST":                  "1",
		"PROFILE_TOP_N":                  "20",
		"BLEND_ALPHA":                    "0.1",
		"BLEND_BETA":                     "0.2",
		"BLEND_GAMMA":                    "0.3",
		"BANDIT_ALGO":                    "thompson",
	}
}

func TestLoadAuthConfig(t *testing.T) {
	src := baseConfigSource()
	orgA := uuid.New()
	orgB := uuid.New()
	src["API_AUTH_ENABLED"] = "true"
	src["API_AUTH_KEYS"] = fmt.Sprintf("key-one:%s|%s,key-two:*", orgA.String(), orgB.String())
	src["API_RATE_LIMIT_RPM"] = "120"
	src["API_RATE_LIMIT_BURST"] = "10"

	cfg, err := Load(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Auth.Enabled {
		t.Fatalf("expected auth enabled")
	}

	keyOne, ok := cfg.Auth.APIKeys["key-one"]
	if !ok {
		t.Fatalf("expected key-one to be present")
	}
	if keyOne.AllowAll {
		t.Fatalf("expected key-one to be scoped to orgs")
	}
	if len(keyOne.OrgIDs) != 2 {
		t.Fatalf("expected two org ids, got %d", len(keyOne.OrgIDs))
	}

	seen := map[uuid.UUID]struct{}{}
	for _, id := range keyOne.OrgIDs {
		seen[id] = struct{}{}
	}
	if _, ok := seen[orgA]; !ok {
		t.Fatalf("expected orgA in key-one scope")
	}
	if _, ok := seen[orgB]; !ok {
		t.Fatalf("expected orgB in key-one scope")
	}

	keyTwo, ok := cfg.Auth.APIKeys["key-two"]
	if !ok || !keyTwo.AllowAll {
		t.Fatalf("expected key-two wildcard access")
	}

	if !cfg.Auth.RateLimit.Enabled {
		t.Fatalf("expected rate limiting enabled")
	}
	if cfg.Auth.RateLimit.RequestsPerMinute != 120 {
		t.Fatalf("unexpected rpm: %d", cfg.Auth.RateLimit.RequestsPerMinute)
	}
	if cfg.Auth.RateLimit.Burst != 10 {
		t.Fatalf("unexpected burst: %d", cfg.Auth.RateLimit.Burst)
	}
}

func TestLoadAuthConfigRequiresKeysWhenEnabled(t *testing.T) {
	src := baseConfigSource()
	src["API_AUTH_ENABLED"] = "true"

	_, err := Load(context.Background(), src)
	if err == nil {
		t.Fatalf("expected error for missing API_AUTH_KEYS")
	}
}
