package usecase

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	reportjson "github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/reporting/json"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
)

func TestAACheckUniformish(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	exposures, outcomes, assignments := buildAACheckDataset(now)

	meta := ReportMetadata{
		BinaryVersion:           "test",
		GitCommit:               "deadbeef",
		EffectiveConfig:         []byte(`{"test":true}`),
		InputDatasetFingerprint: "fingerprint",
	}

	use := AACheckUsecase{
		Exposures:   staticExposureReader{items: exposures},
		Outcomes:    staticOutcomeReader{items: outcomes},
		Assignments: staticAssignmentReader{items: assignments},
		Reporter:    reportjson.Writer{},
		Clock:       fixedClock{t: now},
		Logger:      noopLogger{},
		Metadata:    meta,
	}

	cfg := ExperimentConfig{
		ExperimentID: "exp_aa",
		SliceKeys:    []string{"tenant"},
		AACheck: &AACheckConfig{
			Enabled:    true,
			Thresholds: []float64{0.05, 0.01, 0.001},
		},
	}

	rep, err := use.Run(t.Context(), cfg, filepathTemp(t, "aa_check.json"))
	if err != nil {
		t.Fatalf("aa-check run failed: %v", err)
	}
	if rep.AA == nil {
		t.Fatalf("missing aa_check report")
	}
	if rep.AA.TotalTests < 20 {
		t.Fatalf("expected more tests for uniformity check, got %d", rep.AA.TotalTests)
	}
	if !rep.AA.Uniformish {
		t.Fatalf("expected uniformish p-values")
	}
}

func buildAACheckDataset(now time.Time) ([]dataset.Exposure, []dataset.Outcome, []dataset.Assignment) {
	// #nosec G404 -- deterministic RNG for stable test data
	rng := rand.New(rand.NewSource(42))
	variants := []string{"A", "B", "C"}
	tenants := []string{"t1", "t2", "t3", "t4", "t5"}
	perVariantTenant := 50

	exposures := make([]dataset.Exposure, 0, len(variants)*len(tenants)*perVariantTenant)
	outcomes := make([]dataset.Outcome, 0, len(exposures)/5)
	assignments := make([]dataset.Assignment, 0, len(exposures))

	reqIdx := 0
	for _, tenant := range tenants {
		for _, variant := range variants {
			for i := 0; i < perVariantTenant; i++ {
				reqIdx++
				reqID := fmt.Sprintf("r%04d", reqIdx)
				itemID := fmt.Sprintf("item-%s-%d", variant, reqIdx)
				exposures = append(exposures, dataset.Exposure{
					RequestID: reqID,
					UserID:    "user",
					Timestamp: now.Add(time.Duration(reqIdx) * time.Second),
					Items:     []dataset.ExposedItem{{ItemID: itemID, Rank: 1}},
					Context:   map[string]string{"tenant": tenant},
				})
				assignments = append(assignments, dataset.Assignment{
					ExperimentID: "exp_aa",
					Variant:      variant,
					RequestID:    reqID,
					UserID:       "user",
					Timestamp:    now.Add(time.Duration(reqIdx) * time.Second),
					Context:      map[string]string{"tenant": tenant},
				})

				if rng.Float64() < 0.1 {
					outcomes = append(outcomes, dataset.Outcome{
						RequestID: reqID,
						UserID:    "user",
						ItemID:    itemID,
						EventType: "click",
						Value:     1,
						Timestamp: now.Add(time.Duration(reqIdx)*time.Second + 5*time.Second),
					})
				}
				if rng.Float64() < 0.02 {
					outcomes = append(outcomes, dataset.Outcome{
						RequestID: reqID,
						UserID:    "user",
						ItemID:    itemID,
						EventType: "conversion",
						Value:     1,
						Timestamp: now.Add(time.Duration(reqIdx)*time.Second + 10*time.Second),
					})
				}
			}
		}
	}

	return exposures, outcomes, assignments
}
