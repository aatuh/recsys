package jsonl

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

func TestFSRawEventSourceRejectsUnsafeSegments(t *testing.T) {
	source := New(t.TempDir())
	_, errs := source.ReadExposureEvents(
		context.Background(),
		"tenant-a",
		`home\escape`,
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
