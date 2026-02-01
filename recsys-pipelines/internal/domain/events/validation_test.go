package events

import (
	"testing"
	"time"
)

func TestExposureValidate(t *testing.T) {
	e := ExposureEvent{
		Version:   1,
		TS:        time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		Tenant:    "demo",
		Surface:   "home",
		SessionID: "s1",
		ItemID:    "i1",
		Rank:      1,
	}
	if err := e.Validate(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
