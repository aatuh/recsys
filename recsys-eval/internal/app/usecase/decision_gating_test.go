package usecase

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	reportjson "github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/reporting/json"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/decision"
)

func TestDecisionGatingScenarios(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	meta := ReportMetadata{
		BinaryVersion:           "test",
		GitCommit:               "deadbeef",
		EffectiveConfig:         []byte(`{"test":true}`),
		InputDatasetFingerprint: "fingerprint",
	}

	tests := []struct {
		name             string
		clicksA          int
		clicksB          int
		latencyA         float64
		latencyB         float64
		guardrailMax     float64
		primaryThreshold float64
		wantDecision     string
		wantExit         int
	}{
		{
			name:             "ship",
			clicksA:          2,
			clicksB:          4,
			latencyA:         120,
			latencyB:         130,
			guardrailMax:     500,
			primaryThreshold: 0.0,
			wantDecision:     decision.DecisionShip,
			wantExit:         0,
		},
		{
			name:             "hold_guardrail",
			clicksA:          2,
			clicksB:          4,
			latencyA:         120,
			latencyB:         600,
			guardrailMax:     200,
			primaryThreshold: 0.0,
			wantDecision:     decision.DecisionHold,
			wantExit:         2,
		},
		{
			name:             "fail_primary",
			clicksA:          4,
			clicksB:          1,
			latencyA:         120,
			latencyB:         130,
			guardrailMax:     500,
			primaryThreshold: 0.0,
			wantDecision:     decision.DecisionFail,
			wantExit:         3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exposures, outcomes, assignments := buildExperimentData(now, tc.clicksA, tc.clicksB, tc.latencyA, tc.latencyB)

			capture := &captureDecisionWriter{}
			use := ExperimentUsecase{
				Exposures:   staticExposureReader{items: exposures},
				Outcomes:    staticOutcomeReader{items: outcomes},
				Assignments: staticAssignmentReader{items: assignments},
				Reporter:    reportjson.Writer{},
				Decision:    capture,
				Clock:       fixedClock{t: now},
				Logger:      noopLogger{},
				Metadata:    meta,
			}
			cfg := ExperimentConfig{
				ExperimentID:   "exp_1",
				ControlVariant: "A",
				PrimaryMetrics: []string{"ctr"},
				Guardrails: GuardrailConfig{
					MaxLatencyP95Ms: tc.guardrailMax,
				},
				Decision: &DecisionConfig{
					Enabled:           true,
					CandidateVariant:  "B",
					PrimaryThresholds: map[string]float64{"ctr": tc.primaryThreshold},
					GuardrailGate:     true,
				},
			}

			artifact, err := use.RunWithDecision(t.Context(), cfg, filepath.Join(t.TempDir(), "report.json"))
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}
			if artifact == nil {
				t.Fatalf("expected decision artifact")
			}
			if artifact.Decision != tc.wantDecision {
				t.Fatalf("decision mismatch: got=%s want=%s", artifact.Decision, tc.wantDecision)
			}
			if got := artifact.ExitCode(); got != tc.wantExit {
				t.Fatalf("exit code mismatch: got=%d want=%d", got, tc.wantExit)
			}
			if capture.artifact == nil {
				t.Fatalf("decision writer not invoked")
			}
		})
	}
}

func buildExperimentData(now time.Time, clicksA, clicksB int, latencyA, latencyB float64) ([]dataset.Exposure, []dataset.Outcome, []dataset.Assignment) {
	const perVariant = 10
	exposures := make([]dataset.Exposure, 0, perVariant*2)
	outcomes := make([]dataset.Outcome, 0, clicksA+clicksB)
	assignments := make([]dataset.Assignment, 0, perVariant*2)

	addVariant := func(variant string, clicks int, latency float64) {
		for i := 0; i < perVariant; i++ {
			reqID := fmt.Sprintf("%s-%02d", variant, i)
			itemID := fmt.Sprintf("item-%s-%02d", variant, i)
			exposures = append(exposures, dataset.Exposure{
				RequestID: reqID,
				UserID:    "user",
				Timestamp: now.Add(time.Duration(i) * time.Minute),
				Items:     []dataset.ExposedItem{{ItemID: itemID, Rank: 1}},
				LatencyMs: &latency,
			})
			assignments = append(assignments, dataset.Assignment{
				ExperimentID: "exp_1",
				Variant:      variant,
				RequestID:    reqID,
				UserID:       "user",
				Timestamp:    now.Add(time.Duration(i) * time.Minute),
			})
		}
		for i := 0; i < clicks; i++ {
			reqID := fmt.Sprintf("%s-%02d", variant, i)
			outcomes = append(outcomes, dataset.Outcome{
				RequestID: reqID,
				UserID:    "user",
				ItemID:    fmt.Sprintf("item-%s-%02d", variant, i),
				EventType: "click",
				Value:     1,
				Timestamp: now.Add(time.Duration(i) * time.Minute).Add(5 * time.Second),
			})
		}
	}

	addVariant("A", clicksA, latencyA)
	addVariant("B", clicksB, latencyB)
	return exposures, outcomes, assignments
}
