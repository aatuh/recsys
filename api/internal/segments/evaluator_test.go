package segments

import (
	"encoding/json"
	"testing"
	"time"
)

func mustJSON(t *testing.T, v string) json.RawMessage {
	t.Helper()
	return json.RawMessage(v)
}

func TestEvaluatorBasicOps(t *testing.T) {
	ctx := map[string]any{
		"user": map[string]any{
			"id": "u1",
			"traits": map[string]any{
				"tier": "VIP",
			},
		},
		"ctx": map[string]any{
			"region": "FI",
		},
	}
	eval := NewEvaluator(ctx, time.Now())

	t.Run("eq", func(t *testing.T) {
		ok, err := eval.Match(mustJSON(t, `{"eq":["user.traits.tier","VIP"]}`))
		if err != nil || !ok {
			t.Fatalf("expected VIP match, got ok=%v err=%v", ok, err)
		}
	})

	t.Run("in", func(t *testing.T) {
		ok, err := eval.Match(mustJSON(t, `{"in":["ctx.region",["SE","FI"]]}`))
		if err != nil || !ok {
			t.Fatalf("expected region match, got ok=%v err=%v", ok, err)
		}
	})

	t.Run("all any", func(t *testing.T) {
		rule := `{"all":[{"any":[{"eq":["user.id","u1"]},{"eq":["user.id","u2"]}]},{"eq":["ctx.region","FI"]}]}`
		ok, err := eval.Match(mustJSON(t, rule))
		if err != nil || !ok {
			t.Fatalf("expected composite match, got ok=%v err=%v", ok, err)
		}
	})
}

func TestEvaluatorGteDaysSince(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	ctx := map[string]any{
		"user": map[string]any{
			"last_play_ts": time.Date(2024, 12, 31, 10, 0, 0, 0, time.UTC),
		},
	}
	eval := NewEvaluator(ctx, now)

	rule := mustJSON(t, `{"gte_days_since":["user.last_play_ts",7]}`)
	ok, err := eval.Match(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected days since rule to match")
	}
}
