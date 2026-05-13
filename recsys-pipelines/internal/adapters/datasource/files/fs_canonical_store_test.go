package files

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/events"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

func TestFSCanonicalStoreRejectsUnsafeSegmentsOnWrite(t *testing.T) {
	store := NewFSCanonicalStore(t.TempDir())
	err := store.ReplaceExposureEvents(
		context.Background(),
		"../tenant",
		"home",
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		[]events.ExposureEvent{{RequestID: "r1"}},
	)
	if err == nil {
		t.Fatal("ReplaceExposureEvents() error = nil")
	}
	if !strings.Contains(err.Error(), "invalid path segment") {
		t.Fatalf("error = %q, want invalid path segment", err.Error())
	}
}

func TestFSCanonicalStoreRejectsUnsafeSegmentsOnRead(t *testing.T) {
	store := NewFSCanonicalStore(t.TempDir())
	_, errs := store.ReadExposureEvents(
		context.Background(),
		"tenant-a",
		"home/escape",
		windows.DayWindowUTC(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
	)

	err := <-errs
	if err == nil {
		t.Fatal("ReadExposureEvents() error = nil")
	}
	if !strings.Contains(err.Error(), "invalid path segment") {
		t.Fatalf("error = %q, want invalid path segment", err.Error())
	}
}
