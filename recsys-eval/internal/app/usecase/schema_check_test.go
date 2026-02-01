package usecase

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/datasource/jsonl"
	reportjson "github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/reporting/json"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
)

func TestReportSchemaValidation(t *testing.T) {
	root := projectRoot(t)
	schemaPath := filepath.Join(root, "api", "schemas", "report.v1.json")
	dataDir := filepath.Join(root, "testdata", "datasets", "tiny")

	meta := ReportMetadata{
		BinaryVersion:           "test",
		GitCommit:               "test",
		EffectiveConfig:         []byte(`{"test":true}`),
		InputDatasetFingerprint: "test",
	}
	clock := fixedClock{t: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	reporter := reportjson.Writer{}
	logger := noopLogger{}

	exposures := jsonl.NewExposureReader(filepath.Join(dataDir, "exposures.jsonl"))
	outcomes := jsonl.NewOutcomeReader(filepath.Join(dataDir, "outcomes.jsonl"))
	assignments := jsonl.NewAssignmentReader(filepath.Join(dataDir, "assignments.jsonl"))

	runnerByMode := map[string]func(string) error{
		report.ModeOffline: func(out string) error {
			use := OfflineEvalUsecase{
				Exposures: exposures,
				Outcomes:  outcomes,
				Reporter:  reporter,
				Clock:     clock,
				Logger:    logger,
				Metadata:  meta,
			}
			cfg := OfflineConfig{
				Metrics: []metrics.MetricSpec{{Name: "precision", K: 2}},
			}
			_, err := use.Run(t.Context(), cfg, out, "")
			return err
		},
		report.ModeExperiment: func(out string) error {
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
			}
			_, err := use.Run(t.Context(), cfg, out)
			return err
		},
		report.ModeOPE: func(out string) error {
			use := OPEUsecase{
				Exposures: exposures,
				Outcomes:  outcomes,
				Reporter:  reporter,
				Clock:     clock,
				Logger:    logger,
				Metadata:  meta,
			}
			cfg := OPEConfig{
				RewardEvent: "conversion",
				EnableSNIPS: true,
				EnableDR:    true,
				Clipping:    5,
			}
			_, err := use.Run(t.Context(), cfg, out)
			return err
		},
		report.ModeInterleaving: func(out string) error {
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
			_, err := use.Run(t.Context(), cfg, out)
			return err
		},
		report.ModeAACheck: func(out string) error {
			use := AACheckUsecase{
				Exposures:   exposures,
				Outcomes:    outcomes,
				Assignments: assignments,
				Reporter:    reporter,
				Clock:       clock,
				Logger:      logger,
				Metadata:    meta,
			}
			cfg := ExperimentConfig{
				ExperimentID: "exp_123",
				AACheck: &AACheckConfig{
					Enabled:    true,
					Thresholds: []float64{0.05, 0.01, 0.001},
				},
			}
			_, err := use.Run(t.Context(), cfg, out)
			return err
		},
	}

	for _, mode := range report.SupportedModes() {
		runner, ok := runnerByMode[mode]
		if !ok {
			t.Fatalf("missing schema check runner for mode %q", mode)
		}
		t.Run(mode, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), mode+".json")
			if err := runner(out); err != nil {
				t.Fatalf("run failed for mode %s: %v", mode, err)
			}
			if err := ValidateJSONFileAgainstSchema(schemaPath, out); err != nil {
				t.Fatalf("schema validation failed: %v", err)
			}
		})
	}
}
