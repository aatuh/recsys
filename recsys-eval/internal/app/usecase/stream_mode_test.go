package usecase

import (
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/datasource/jsonl"
	reportjson "github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/reporting/json"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
)

func TestStreamModeMatchesMemoryOffline(t *testing.T) {
	root := projectRoot(t)
	dataDir := filepath.Join(root, "testdata", "datasets", "tiny")

	clock := fixedClock{t: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	meta := ReportMetadata{
		BinaryVersion:           "test",
		GitCommit:               "deadbeef",
		EffectiveConfig:         []byte(`{"test":true}`),
		InputDatasetFingerprint: "fingerprint",
	}

	exposures := jsonl.NewExposureReader(filepath.Join(dataDir, "exposures.jsonl"))
	outcomes := jsonl.NewOutcomeReader(filepath.Join(dataDir, "outcomes.jsonl"))
	reporter := reportjson.Writer{}

	cfg := OfflineConfig{
		Metrics:   []metrics.MetricSpec{{Name: "precision", K: 2}, {Name: "recall", K: 2}},
		SliceKeys: []string{"tenant", "surface"},
	}

	mem := OfflineEvalUsecase{
		Exposures: exposures,
		Outcomes:  outcomes,
		Reporter:  reporter,
		Clock:     clock,
		Logger:    noopLogger{},
		Metadata:  meta,
		Scale:     ScaleConfig{Mode: "memory"},
	}
	memReport, err := mem.Run(t.Context(), cfg, filepath.Join(t.TempDir(), "mem.json"), "")
	if err != nil {
		t.Fatalf("memory run failed: %v", err)
	}

	stream := OfflineEvalUsecase{
		Exposures: exposures,
		Outcomes:  outcomes,
		Reporter:  reporter,
		Clock:     clock,
		Logger:    noopLogger{},
		Metadata:  meta,
		Scale:     ScaleConfig{Mode: "stream", Stream: StreamConfig{MaxOpenRequests: 1}},
	}
	streamReport, err := stream.Run(t.Context(), cfg, filepath.Join(t.TempDir(), "stream.json"), "")
	if err != nil {
		t.Fatalf("stream run failed: %v", err)
	}

	compareMetricSets(t, memReport.Offline.Metrics, streamReport.Offline.Metrics)
	compareSegmentMetrics(t, memReport.Offline.BySegment, streamReport.Offline.BySegment)
}

func TestStreamModeMatchesMemoryExperiment(t *testing.T) {
	root := projectRoot(t)
	dataDir := filepath.Join(root, "testdata", "datasets", "tiny")

	clock := fixedClock{t: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	meta := ReportMetadata{
		BinaryVersion:           "test",
		GitCommit:               "deadbeef",
		EffectiveConfig:         []byte(`{"test":true}`),
		InputDatasetFingerprint: "fingerprint",
	}

	exposures := jsonl.NewExposureReader(filepath.Join(dataDir, "exposures.jsonl"))
	outcomes := jsonl.NewOutcomeReader(filepath.Join(dataDir, "outcomes.jsonl"))
	assignments := jsonl.NewAssignmentReader(filepath.Join(dataDir, "assignments.jsonl"))
	reporter := reportjson.Writer{}

	cfg := ExperimentConfig{
		ExperimentID:   "exp_123",
		ControlVariant: "A",
		SliceKeys:      []string{"tenant"},
		Guardrails: GuardrailConfig{
			MaxLatencyP95Ms: 500,
			MaxErrorRate:    0.1,
			MaxEmptyRate:    0.5,
		},
		MultipleComparisons: &MultipleComparisonsConfig{
			Method: "fdr_bh",
			Alpha:  0.05,
		},
	}

	mem := ExperimentUsecase{
		Exposures:   exposures,
		Outcomes:    outcomes,
		Assignments: assignments,
		Reporter:    reporter,
		Clock:       clock,
		Logger:      noopLogger{},
		Metadata:    meta,
		Scale:       ScaleConfig{Mode: "memory"},
	}
	memReport, err := mem.Run(t.Context(), cfg, filepath.Join(t.TempDir(), "mem.json"))
	if err != nil {
		t.Fatalf("memory run failed: %v", err)
	}

	stream := ExperimentUsecase{
		Exposures:   exposures,
		Outcomes:    outcomes,
		Assignments: assignments,
		Reporter:    reporter,
		Clock:       clock,
		Logger:      noopLogger{},
		Metadata:    meta,
		Scale:       ScaleConfig{Mode: "stream", Stream: StreamConfig{MaxOpenRequests: 1}},
	}
	streamReport, err := stream.Run(t.Context(), cfg, filepath.Join(t.TempDir(), "stream.json"))
	if err != nil {
		t.Fatalf("stream run failed: %v", err)
	}

	compareVariantMetrics(t, memReport.Experiment.Variants, streamReport.Experiment.Variants)
}

func compareMetricSets(t *testing.T, a, b []report.MetricResult) {
	t.Helper()
	if len(a) != len(b) {
		t.Fatalf("metric length mismatch: %d vs %d", len(a), len(b))
	}
	index := map[string]float64{}
	for _, m := range a {
		index[m.Name] = m.Value
	}
	for _, m := range b {
		av, ok := index[m.Name]
		if !ok {
			t.Fatalf("missing metric %s", m.Name)
		}
		if math.Abs(av-m.Value) > 1e-9 {
			t.Fatalf("metric %s mismatch: %.9f vs %.9f", m.Name, av, m.Value)
		}
	}
}

func compareSegmentMetrics(t *testing.T, a, b map[string][]report.MetricResult) {
	t.Helper()
	if len(a) != len(b) {
		t.Fatalf("segment count mismatch: %d vs %d", len(a), len(b))
	}
	for seg, metrics := range a {
		other, ok := b[seg]
		if !ok {
			t.Fatalf("missing segment %s", seg)
		}
		compareMetricSets(t, metrics, other)
	}
}

func compareVariantMetrics(t *testing.T, a, b []report.VariantMetrics) {
	t.Helper()
	if len(a) != len(b) {
		t.Fatalf("variant count mismatch: %d vs %d", len(a), len(b))
	}
	index := map[string]report.VariantMetrics{}
	for _, v := range a {
		index[v.Variant] = v
	}
	for _, v := range b {
		base, ok := index[v.Variant]
		if !ok {
			t.Fatalf("missing variant %s", v.Variant)
		}
		if base.Exposures != v.Exposures || base.Clicks != v.Clicks || base.Conversions != v.Conversions {
			t.Fatalf("variant %s counts mismatch", v.Variant)
		}
		if math.Abs(base.CTR-v.CTR) > 1e-9 || math.Abs(base.ConversionRate-v.ConversionRate) > 1e-9 {
			t.Fatalf("variant %s rate mismatch", v.Variant)
		}
	}
}
