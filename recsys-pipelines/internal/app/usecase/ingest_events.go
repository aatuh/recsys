package usecase

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/events"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
)

type IngestEvents struct {
	rt        runtime.Runtime
	raw       datasource.RawEventSource
	canonical datasource.CanonicalStore
	maxEvents int
}

func NewIngestEvents(
	rt runtime.Runtime,
	raw datasource.RawEventSource,
	canonical datasource.CanonicalStore,
	maxEvents int,
) *IngestEvents {
	return &IngestEvents{
		rt:        rt,
		raw:       raw,
		canonical: canonical,
		maxEvents: maxEvents,
	}
}

func (uc *IngestEvents) Execute(ctx context.Context, tenant, surface string, w windows.Window) error {
	start := uc.rt.Clock.NowUTC()
	uc.rt.Logger.Info(ctx, "ingest: start",
		logger.Field{Key: "tenant", Value: tenant},
		logger.Field{Key: "surface", Value: surface},
		logger.Field{Key: "start", Value: w.Start.Format(time.RFC3339)},
		logger.Field{Key: "end", Value: w.End.Format(time.RFC3339)},
	)

	evCh, errCh := uc.raw.ReadExposureEvents(ctx, tenant, surface, w)

	// Prepare a full day-partition map for the requested window so reruns can be
	// made idempotent even when a day has zero events.
	perDay := map[string][]events.ExposureEvent{}
	startDay := time.Date(w.Start.Year(), w.Start.Month(), w.Start.Day(), 0, 0, 0, 0, time.UTC)
	endDay := time.Date(w.End.Year(), w.End.Month(), w.End.Day(), 0, 0, 0, 0, time.UTC)
	for day := startDay; day.Before(endDay); day = day.Add(24 * time.Hour) {
		perDay[day.Format("2006-01-02")] = nil
	}
	count := 0
	for ev := range evCh {
		count++
		if uc.maxEvents > 0 && count > uc.maxEvents {
			return fmt.Errorf("event limit exceeded: %d > %d", count, uc.maxEvents)
		}
		day := ev.TS.UTC().Format("2006-01-02")
		perDay[day] = append(perDay[day], ev)
	}
	if err := <-errCh; err != nil {
		return err
	}

	days := make([]string, 0, len(perDay))
	for d := range perDay {
		days = append(days, d)
	}
	sort.Strings(days)
	for _, dayStr := range days {
		day, _ := time.ParseInLocation("2006-01-02", dayStr, time.UTC)
		evs := perDay[dayStr]
		if err := uc.canonical.ReplaceExposureEvents(ctx, tenant, surface, day, evs); err != nil {
			return err
		}
	}

	dur := uc.rt.Clock.NowUTC().Sub(start)
	uc.rt.Logger.Info(ctx, "ingest: done",
		logger.Field{Key: "events", Value: count},
		logger.Field{Key: "duration_ms", Value: dur.Milliseconds()},
	)
	uc.rt.Metrics.IncCounter("ingest_events_total", int64(count),
		map[string]string{"tenant": tenant, "surface": surface})
	uc.rt.Metrics.ObserveDuration("ingest_duration", dur,
		map[string]string{"tenant": tenant, "surface": surface})
	return nil
}
