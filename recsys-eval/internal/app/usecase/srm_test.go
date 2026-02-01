package usecase

import (
	"strconv"
	"testing"
	"time"

	reportjson "github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/reporting/json"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/decision"
)

func TestSRMDetection(t *testing.T) {
	assignments := make([]dataset.Assignment, 0, 100)
	for i := 0; i < 90; i++ {
		assignments = append(assignments, dataset.Assignment{ExperimentID: "exp", Variant: "A", RequestID: "A-" + strconv.Itoa(i)})
	}
	for i := 0; i < 10; i++ {
		assignments = append(assignments, dataset.Assignment{ExperimentID: "exp", Variant: "B", RequestID: "B-" + strconv.Itoa(i)})
	}

	cfg := &SRMConfig{
		Enabled:   true,
		Alpha:     0.05,
		MinSample: 20,
		Expected:  map[string]float64{"A": 0.5, "B": 0.5},
	}
	rep := buildSRMReport(cfg, assignments, "exp")
	if rep == nil || !rep.Global.Detected {
		t.Fatalf("expected SRM detection")
	}
}

func TestSRMGateHoldDecision(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	exposures, outcomes, assignments := buildSkewedAssignments(now)

	meta := ReportMetadata{
		BinaryVersion:           "test",
		GitCommit:               "deadbeef",
		EffectiveConfig:         []byte(`{"test":true}`),
		InputDatasetFingerprint: "fingerprint",
	}

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
		ExperimentID:   "exp_srm",
		ControlVariant: "A",
		PrimaryMetrics: []string{"ctr"},
		Decision: &DecisionConfig{
			Enabled:          true,
			CandidateVariant: "B",
		},
		SRM: &SRMConfig{
			Enabled:     true,
			Alpha:       0.05,
			MinSample:   20,
			Expected:    map[string]float64{"A": 0.5, "B": 0.5},
			GateEnabled: true,
		},
	}

	artifact, err := use.RunWithDecision(t.Context(), cfg, filepathTemp(t, "srm.json"))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if artifact == nil {
		t.Fatalf("expected decision artifact")
	}
	if artifact.Decision != decision.DecisionHold {
		t.Fatalf("expected hold decision due to SRM, got %s", artifact.Decision)
	}
	if capture.artifact == nil || capture.artifact.SRM == nil || !capture.artifact.SRM.Detected {
		t.Fatalf("expected SRM decision details")
	}
}

func buildSkewedAssignments(now time.Time) ([]dataset.Exposure, []dataset.Outcome, []dataset.Assignment) {
	exposures := make([]dataset.Exposure, 0, 100)
	outcomes := make([]dataset.Outcome, 0)
	assignments := make([]dataset.Assignment, 0, 100)

	add := func(variant string, count int) {
		for i := 0; i < count; i++ {
			reqID := variant + "-" + strconv.Itoa(i)
			exposures = append(exposures, dataset.Exposure{
				RequestID: reqID,
				UserID:    "user",
				Timestamp: now.Add(time.Duration(len(exposures)) * time.Second),
				Items:     []dataset.ExposedItem{{ItemID: "item-" + reqID, Rank: 1}},
			})
			assignments = append(assignments, dataset.Assignment{
				ExperimentID: "exp_srm",
				Variant:      variant,
				RequestID:    reqID,
				UserID:       "user",
				Timestamp:    now.Add(time.Duration(len(assignments)) * time.Second),
			})
		}
	}

	add("A", 90)
	add("B", 10)
	return exposures, outcomes, assignments
}
