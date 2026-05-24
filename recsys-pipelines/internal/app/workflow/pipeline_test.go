package workflow

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/usecase"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/events"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/signals"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
	plog "github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
	pmet "github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/metrics"
)

func TestRunDayPublishesEnabledRichArtifacts(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	raw := &workflowRaw{events: exposureFixture(now)}
	canonical := &workflowCanonical{}
	registry := &workflowRegistry{}
	pipe := newTestPipeline(t, now, raw, canonical, registry, config.ArtifactSelection{
		Popularity: true,
		Cooc:       true,
		Implicit:   true,
		ContentSim: true,
		SessionSeq: true,
	})

	err := pipe.RunDay(context.Background(), "demo", "home", "default", windows.DayWindowUTC(now))
	if err != nil {
		t.Fatalf("RunDay() error = %v", err)
	}

	manifest, ok, err := registry.LoadManifest(context.Background(), "demo", "home")
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	if !ok {
		t.Fatalf("manifest was not published")
	}
	for _, kind := range []string{"popularity", "cooc", "implicit", "content_sim", "session_seq"} {
		if strings.TrimSpace(manifest.Current[kind]) == "" {
			t.Fatalf("manifest.Current[%q] is empty: %+v", kind, manifest.Current)
		}
	}
}

func TestRunDayPreservesPopularityCoocOnlySelection(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	registry := &workflowRegistry{}
	pipe := newTestPipeline(t, now, &workflowRaw{events: exposureFixture(now)}, &workflowCanonical{}, registry, config.ArtifactSelection{
		Popularity: true,
		Cooc:       true,
	})

	err := pipe.RunDay(context.Background(), "demo", "home", "default", windows.DayWindowUTC(now))
	if err != nil {
		t.Fatalf("RunDay() error = %v", err)
	}

	manifest, ok, err := registry.LoadManifest(context.Background(), "demo", "home")
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	if !ok {
		t.Fatalf("manifest was not published")
	}
	if manifest.Current["popularity"] == "" || manifest.Current["cooc"] == "" {
		t.Fatalf("manifest missing default artifacts: %+v", manifest.Current)
	}
	if manifest.Current["implicit"] != "" || manifest.Current["content_sim"] != "" || manifest.Current["session_seq"] != "" {
		t.Fatalf("manifest published disabled rich artifacts: %+v", manifest.Current)
	}
}

func TestRunDayPropagatesRichArtifactLimitFailures(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rt := testRuntime(now)
	canonical := &workflowCanonical{}
	pipe := &Pipeline{
		RT:       rt,
		Ingest:   usecase.NewIngestEvents(rt, &workflowRaw{events: exposureFixture(now)}, canonical, 100),
		Validate: usecase.NewValidateQuality(rt, workflowValidator{}),
		Pop:      usecase.NewComputePopularity(rt, canonical, 10, 100),
		Cooc:     usecase.NewComputeCooc(rt, canonical, 10, 1, 10, 100, 10, 100),
		Implicit: usecase.NewComputeImplicit(rt, canonical, 10, 1, 10),
		Publish:  usecase.NewPublishArtifacts(rt, &workflowObjectStore{objects: map[string][]byte{}}, &workflowRegistry{}, workflowValidator{}),
		Artifacts: config.ArtifactSelection{
			Popularity: true,
			Implicit:   true,
		},
	}

	err := pipe.RunDay(context.Background(), "demo", "home", "default", windows.DayWindowUTC(now))
	if err == nil {
		t.Fatalf("RunDay() error = nil")
	}
	if !strings.Contains(err.Error(), "distinct user limit exceeded") {
		t.Fatalf("RunDay() error = %q, want distinct user limit", err)
	}
}

func newTestPipeline(
	t *testing.T,
	now time.Time,
	raw *workflowRaw,
	canonical *workflowCanonical,
	registry *workflowRegistry,
	selection config.ArtifactSelection,
) *Pipeline {
	t.Helper()
	rt := testRuntime(now)
	return &Pipeline{
		RT:       rt,
		Ingest:   usecase.NewIngestEvents(rt, raw, canonical, 100),
		Validate: usecase.NewValidateQuality(rt, workflowValidator{}),
		Pop:      usecase.NewComputePopularity(rt, canonical, 10, 100),
		Cooc:     usecase.NewComputeCooc(rt, canonical, 10, 1, 10, 100, 10, 100),
		Implicit: usecase.NewComputeImplicit(rt, canonical, 10, 100, 10),
		Content: usecase.NewComputeContentSim(rt, workflowCatalog{items: []signals.ItemTag{
			{ItemID: "sku-1", Namespace: "home", Tags: []string{"brand:a", "category:x"}},
			{ItemID: "sku-2", Namespace: "other", Tags: []string{"brand:b"}},
		}}, 10),
		Session:   usecase.NewComputeSessionSeq(rt, canonical, 10, 100, 10),
		Publish:   usecase.NewPublishArtifacts(rt, &workflowObjectStore{objects: map[string][]byte{}}, registry, workflowValidator{}),
		Artifacts: selection,
	}
}

func exposureFixture(now time.Time) []events.ExposureEvent {
	return []events.ExposureEvent{
		{Version: 1, TS: now.Add(time.Hour), Tenant: "demo", Surface: "home", UserID: "u1", SessionID: "s1", ItemID: "sku-1", Rank: 1},
		{Version: 1, TS: now.Add(time.Hour + time.Second), Tenant: "demo", Surface: "home", UserID: "u1", SessionID: "s1", ItemID: "sku-2", Rank: 2},
		{Version: 1, TS: now.Add(2 * time.Hour), Tenant: "demo", Surface: "home", UserID: "u2", SessionID: "s2", ItemID: "sku-2", Rank: 1},
	}
}

func testRuntime(now time.Time) runtime.Runtime {
	return runtime.Runtime{Clock: workflowClock{now: now}, Logger: workflowLogger{}, Metrics: workflowMetrics{}}
}

type workflowClock struct{ now time.Time }

func (c workflowClock) NowUTC() time.Time { return c.now }

type workflowLogger struct{}

func (workflowLogger) Debug(context.Context, string, ...plog.Field) {}
func (workflowLogger) Info(context.Context, string, ...plog.Field)  {}
func (workflowLogger) Warn(context.Context, string, ...plog.Field)  {}
func (workflowLogger) Error(context.Context, string, ...plog.Field) {}

type workflowMetrics struct{}

func (workflowMetrics) IncCounter(string, int64, map[string]string)              {}
func (workflowMetrics) ObserveDuration(string, time.Duration, map[string]string) {}

var _ pmet.Metrics = workflowMetrics{}

type workflowRaw struct {
	events []events.ExposureEvent
	err    error
}

func (r *workflowRaw) ReadExposureEvents(
	ctx context.Context,
	_ string,
	_ string,
	w windows.Window,
) (<-chan events.ExposureEvent, <-chan error) {
	out := make(chan events.ExposureEvent, 8)
	errs := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errs)
		if r.err != nil {
			errs <- r.err
			return
		}
		for _, ev := range r.events {
			if w.Contains(ev.TS) {
				out <- ev
			}
		}
		errs <- ctx.Err()
	}()
	return out, errs
}

var _ datasource.RawEventSource = (*workflowRaw)(nil)

type workflowCanonical struct {
	events []events.ExposureEvent
}

func (c *workflowCanonical) AppendExposureEvents(
	ctx context.Context,
	_ string,
	_ string,
	_ time.Time,
	evs []events.ExposureEvent,
) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.events = append(c.events, evs...)
	return nil
}

func (c *workflowCanonical) ReplaceExposureEvents(
	ctx context.Context,
	_ string,
	_ string,
	day time.Time,
	evs []events.ExposureEvent,
) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	w := windows.DayWindowUTC(day)
	next := c.events[:0]
	for _, ev := range c.events {
		if !w.Contains(ev.TS) {
			next = append(next, ev)
		}
	}
	c.events = append(next, evs...)
	return nil
}

func (c *workflowCanonical) ReadExposureEvents(
	ctx context.Context,
	_ string,
	_ string,
	w windows.Window,
) (<-chan events.ExposureEvent, <-chan error) {
	out := make(chan events.ExposureEvent, 8)
	errs := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errs)
		for _, ev := range c.events {
			if w.Contains(ev.TS) {
				out <- ev
			}
		}
		errs <- ctx.Err()
	}()
	return out, errs
}

var _ datasource.CanonicalStore = (*workflowCanonical)(nil)

type workflowCatalog struct {
	items []signals.ItemTag
	err   error
}

func (c workflowCatalog) Read(ctx context.Context) ([]signals.ItemTag, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if c.err != nil {
		return nil, c.err
	}
	return c.items, nil
}

type workflowObjectStore struct {
	objects map[string][]byte
}

func (s *workflowObjectStore) Put(_ context.Context, key string, _ string, data []byte) (string, error) {
	uri := "mem://" + key
	s.objects[uri] = append([]byte(nil), data...)
	return uri, nil
}

func (s *workflowObjectStore) Get(_ context.Context, uri string) ([]byte, error) {
	data, ok := s.objects[uri]
	if !ok {
		return nil, errors.New("not found")
	}
	return append([]byte(nil), data...), nil
}

type workflowRegistry struct {
	manifest artifacts.ManifestV1
	ok       bool
}

func (r *workflowRegistry) Record(context.Context, artifacts.Ref) error { return nil }

func (r *workflowRegistry) LoadManifest(_ context.Context, tenant, surface string) (artifacts.ManifestV1, bool, error) {
	if !r.ok || r.manifest.Tenant != tenant || r.manifest.Surface != surface {
		return artifacts.ManifestV1{}, false, nil
	}
	return r.manifest, true, nil
}

func (r *workflowRegistry) SwapManifest(_ context.Context, _ string, _ string, next artifacts.ManifestV1) error {
	r.manifest = next
	r.ok = true
	return nil
}

type workflowValidator struct{}

func (workflowValidator) ValidateCanonical(context.Context, string, string, windows.Window) error {
	return nil
}

func (workflowValidator) ValidateArtifact(context.Context, artifacts.Ref, []byte) error {
	return nil
}
