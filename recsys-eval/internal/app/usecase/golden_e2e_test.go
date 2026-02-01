package usecase

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/datasource/jsonl"
	reportjson "github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/reporting/json"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"
)

func TestGoldenReports(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping golden tests in short mode")
	}

	root := projectRoot(t)
	goldenDir := filepath.Join(root, "testdata", "golden")
	if err := os.MkdirAll(goldenDir, 0o750); err != nil {
		t.Fatalf("create golden dir: %v", err)
	}
	dataDir := filepath.Join(root, "testdata", "datasets", "tiny")

	meta := ReportMetadata{
		BinaryVersion:           "test",
		GitCommit:               "deadbeef",
		EffectiveConfig:         []byte(`{"test":true}`),
		InputDatasetFingerprint: "fingerprint",
	}
	clock := fixedClock{t: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	logger := noopLogger{}
	reporter := reportjson.Writer{}

	exposures := jsonl.NewExposureReader(filepath.Join(dataDir, "exposures.jsonl"))
	outcomes := jsonl.NewOutcomeReader(filepath.Join(dataDir, "outcomes.jsonl"))
	assignments := jsonl.NewAssignmentReader(filepath.Join(dataDir, "assignments.jsonl"))

	t.Run("offline", func(t *testing.T) {
		basePath := filepath.Join(t.TempDir(), "baseline.json")
		use := OfflineEvalUsecase{
			Exposures: exposures,
			Outcomes:  outcomes,
			Reporter:  reporter,
			Clock:     clock,
			Logger:    logger,
			Metadata:  meta,
		}
		baseCfg := OfflineConfig{
			Metrics:   []metrics.MetricSpec{{Name: "precision", K: 2}, {Name: "recall", K: 2}},
			SliceKeys: []string{"tenant", "surface"},
		}
		if _, err := use.Run(t.Context(), baseCfg, basePath, ""); err != nil {
			t.Fatalf("baseline run failed: %v", err)
		}

		out := filepath.Join(t.TempDir(), "offline.json")
		if _, err := use.Run(t.Context(), baseCfg, out, basePath); err != nil {
			t.Fatalf("offline run failed: %v", err)
		}
		assertGolden(t, filepath.Join(goldenDir, "offline.json"), out)
	})

	t.Run("experiment", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "experiment.json")
		use := ExperimentUsecase{
			Exposures:   exposures,
			Outcomes:    outcomes,
			Assignments: assignments,
			Reporter:    reporter,
			Clock:       clock,
			Logger:      logger,
			Metadata:    meta,
		}
		cfg := ExperimentConfig{
			ExperimentID:   "exp_123",
			ControlVariant: "A",
			PrimaryMetrics: []string{"ctr"},
			SliceKeys:      []string{"tenant"},
			Guardrails: GuardrailConfig{
				MaxLatencyP95Ms: 100,
				MaxErrorRate:    0.0,
				MaxEmptyRate:    0.5,
			},
			MultipleComparisons: &MultipleComparisonsConfig{
				Method: "fdr_bh",
				Alpha:  0.05,
			},
		}
		if _, err := use.Run(t.Context(), cfg, out); err != nil {
			t.Fatalf("experiment run failed: %v", err)
		}
		assertGolden(t, filepath.Join(goldenDir, "experiment.json"), out)
	})

	t.Run("ope", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "ope.json")
		use := OPEUsecase{
			Exposures: exposures,
			Outcomes:  outcomes,
			Reporter:  reporter,
			Clock:     clock,
			Logger:    logger,
			Metadata:  meta,
		}
		cfg := OPEConfig{
			RewardEvent:       "conversion",
			EnableSNIPS:       true,
			EnableDR:          false,
			Clipping:          5,
			Unit:              "request",
			RewardAggregation: "sum",
		}
		if _, err := use.Run(t.Context(), cfg, out); err != nil {
			t.Fatalf("ope run failed: %v", err)
		}
		assertGolden(t, filepath.Join(goldenDir, "ope.json"), out)
	})

	t.Run("interleaving", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "interleaving.json")
		use := InterleavingUsecase{
			RankerA:  jsonl.NewRankListReader(filepath.Join(dataDir, "ranker_a.jsonl")),
			RankerB:  jsonl.NewRankListReader(filepath.Join(dataDir, "ranker_b.jsonl")),
			Outcomes: jsonl.NewOutcomeReader(filepath.Join(dataDir, "outcomes_interleaving.jsonl")),
			Reporter: reporter,
			Clock:    clock,
			Logger:   logger,
			Metadata: meta,
		}
		cfg := InterleavingConfig{
			Algorithm: "team_draft",
			Seed:      42,
		}
		if _, err := use.Run(t.Context(), cfg, out); err != nil {
			t.Fatalf("interleaving run failed: %v", err)
		}
		assertGolden(t, filepath.Join(goldenDir, "interleaving.json"), out)
	})
}

func assertGolden(t *testing.T, goldenPath, actualPath string) {
	t.Helper()

	actual := canonicalizeJSONFile(t, actualPath)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(goldenPath, actual, 0o600); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v (set UPDATE_GOLDEN=1 to create)", goldenPath, err)
	}
	if string(golden) != string(actual) {
		t.Fatalf("golden mismatch for %s", filepath.Base(goldenPath))
	}
}

func canonicalizeJSONFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	out = append(out, '\n')
	return out
}
