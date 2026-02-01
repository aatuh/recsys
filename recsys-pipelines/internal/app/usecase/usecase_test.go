package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/events"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
	plog "github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
	pmet "github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/metrics"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) NowUTC() time.Time { return c.t }

type testLogger struct{}

func (testLogger) Debug(context.Context, string, ...plog.Field) {}
func (testLogger) Info(context.Context, string, ...plog.Field)  {}
func (testLogger) Warn(context.Context, string, ...plog.Field)  {}
func (testLogger) Error(context.Context, string, ...plog.Field) {}

type testMetrics struct{}

var _ pmet.Metrics = testMetrics{}

func (testMetrics) IncCounter(string, int64, map[string]string) {}

func (testMetrics) ObserveDuration(string, time.Duration, map[string]string) {}

type memCanonical struct {
	evs []events.ExposureEvent
}

var _ datasource.CanonicalStore = (*memCanonical)(nil)

func (m *memCanonical) AppendExposureEvents(
	ctx context.Context,
	_ string,
	_ string,
	_ time.Time,
	evs []events.ExposureEvent,
) error {
	// Deprecated: kept for backwards compatibility in tests.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	m.evs = append(m.evs, evs...)
	return nil
}

func (m *memCanonical) ReplaceExposureEvents(
	ctx context.Context,
	_ string,
	_ string,
	day time.Time,
	evs []events.ExposureEvent,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Remove existing events for the day.
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	filtered := m.evs[:0]
	for _, e := range m.evs {
		if e.TS.Before(start) || !e.TS.Before(end) {
			filtered = append(filtered, e)
		}
	}
	m.evs = append(filtered, evs...)
	return nil
}

func (m *memCanonical) ReadExposureEvents(
	ctx context.Context,
	_ string,
	_ string,
	w windows.Window,
) (<-chan events.ExposureEvent, <-chan error) {
	out := make(chan events.ExposureEvent, 16)
	errs := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errs)
		for _, e := range m.evs {
			if w.Contains(e.TS) {
				out <- e
			}
		}
		errs <- nil
	}()
	return out, errs
}

type memRaw struct {
	evs []events.ExposureEvent
	err error
}

var _ datasource.RawEventSource = (*memRaw)(nil)

func (m *memRaw) ReadExposureEvents(
	ctx context.Context,
	_ string,
	_ string,
	w windows.Window,
) (<-chan events.ExposureEvent, <-chan error) {
	out := make(chan events.ExposureEvent, 16)
	errs := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errs)
		if m.err != nil {
			errs <- m.err
			return
		}
		for _, e := range m.evs {
			if w.Contains(e.TS) {
				out <- e
			}
		}
		errs <- nil
	}()
	return out, errs
}

func TestIngestAndPopularity(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rt := runtime.Runtime{
		Clock:   fixedClock{t: now},
		Logger:  testLogger{},
		Metrics: testMetrics{},
	}

	w := windows.DayWindowUTC(now)
	raw := &memRaw{
		evs: []events.ExposureEvent{
			{Version: 1, TS: now.Add(1 * time.Hour), Tenant: "demo", Surface: "home", SessionID: "s1", ItemID: "A"},
			{Version: 1, TS: now.Add(2 * time.Hour), Tenant: "demo", Surface: "home", SessionID: "s1", ItemID: "A"},
			{Version: 1, TS: now.Add(3 * time.Hour), Tenant: "demo", Surface: "home", SessionID: "s2", ItemID: "B"},
		},
	}
	canon := &memCanonical{}

	ing := NewIngestEvents(rt, raw, canon, 100)
	if err := ing.Execute(context.Background(), "demo", "home", w); err != nil {
		t.Fatalf("ingest err: %v", err)
	}

	pop := NewComputePopularity(rt, canon, 10, 100)
	ref, blob, err := pop.Execute(context.Background(), "demo", "home", "", w)
	if err != nil {
		t.Fatalf("pop err: %v", err)
	}
	if ref.Version == "" || len(blob) == 0 {
		t.Fatalf("expected output")
	}
}

func TestIngestPropagatesError(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rt := runtime.Runtime{
		Clock:   fixedClock{t: now},
		Logger:  testLogger{},
		Metrics: testMetrics{},
	}
	w := windows.DayWindowUTC(now)

	raw := &memRaw{err: errors.New("boom")}
	canon := &memCanonical{}
	ing := NewIngestEvents(rt, raw, canon, 100)
	if err := ing.Execute(context.Background(), "demo", "home", w); err == nil {
		t.Fatalf("expected error")
	}
}

func TestIngestIsIdempotentPerDay(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rt := runtime.Runtime{
		Clock:   fixedClock{t: now},
		Logger:  testLogger{},
		Metrics: testMetrics{},
	}
	w := windows.DayWindowUTC(now)

	raw := &memRaw{
		evs: []events.ExposureEvent{
			{Version: 1, TS: now.Add(1 * time.Hour), Tenant: "demo", Surface: "home", SessionID: "s1", ItemID: "A"},
			{Version: 1, TS: now.Add(2 * time.Hour), Tenant: "demo", Surface: "home", SessionID: "s2", ItemID: "B"},
		},
	}
	canon := &memCanonical{}

	ing := NewIngestEvents(rt, raw, canon, 100)
	if err := ing.Execute(context.Background(), "demo", "home", w); err != nil {
		t.Fatalf("ingest 1 err: %v", err)
	}
	if err := ing.Execute(context.Background(), "demo", "home", w); err != nil {
		t.Fatalf("ingest 2 err: %v", err)
	}
	if got, want := len(canon.evs), len(raw.evs); got != want {
		t.Fatalf("expected canonical events to be idempotent: got=%d want=%d", got, want)
	}
}
