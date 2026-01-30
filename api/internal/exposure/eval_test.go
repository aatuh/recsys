package exposure

import (
	"testing"
	"time"
)

func TestBuildEvalExposureRequiresRequestID(t *testing.T) {
	_, err := buildEvalExposure(Event{
		Items: []Item{{ItemID: "item-1", Rank: 1}},
	})
	if err == nil {
		t.Fatalf("expected error when request_id is missing")
	}
}

func TestBuildEvalExposureMapsUserAndContext(t *testing.T) {
	ts := time.Date(2026, 1, 30, 10, 0, 0, 0, time.UTC)
	event := Event{
		RequestID:     "req-1",
		TenantID:      "tenant-1",
		Surface:       "home",
		Segment:       "default",
		AlgoVersion:   "algo@1",
		ConfigVersion: "cfg@1",
		RulesVersion:  "rules@1",
		Experiment:    &Experiment{ID: "exp-1", Variant: "A"},
		Subject:       &Subject{UserIDHash: "userhash"},
		Context:       &Context{Locale: "fi-FI", Device: "mobile"},
		Timestamp:     ts,
		Items:         []Item{{ItemID: "item-1", Rank: 1}},
	}
	exp, err := buildEvalExposure(event)
	if err != nil {
		t.Fatalf("build eval exposure: %v", err)
	}
	if exp.RequestID != "req-1" {
		t.Fatalf("expected request_id to match, got %q", exp.RequestID)
	}
	if exp.UserID != "userhash" {
		t.Fatalf("expected user_id hash, got %q", exp.UserID)
	}
	if exp.Timestamp != ts {
		t.Fatalf("expected timestamp to match")
	}
	if exp.Context == nil || exp.Context["surface"] != "home" {
		t.Fatalf("expected surface context")
	}
	if exp.Context["experiment_id"] != "exp-1" || exp.Context["experiment_variant"] != "A" {
		t.Fatalf("expected experiment context")
	}
}

func TestBuildEvalExposureFallbackUserID(t *testing.T) {
	event := Event{
		RequestID: "req-2",
		Items:     []Item{{ItemID: "item-1", Rank: 1}},
	}
	exp, err := buildEvalExposure(event)
	if err != nil {
		t.Fatalf("build eval exposure: %v", err)
	}
	if exp.UserID != "req-2" {
		t.Fatalf("expected fallback user_id to be request_id, got %q", exp.UserID)
	}
}
