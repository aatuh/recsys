package usecase

import (
	"path/filepath"
	"testing"
	"time"

	reportjson "github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/reporting/json"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
)

func TestOPEUnitAggregation(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	exposures := []dataset.Exposure{
		{
			RequestID: "r1",
			UserID:    "u1",
			Timestamp: now,
			Items: []dataset.ExposedItem{
				item("i1", 1, 0.5, 0.6),
				item("i2", 2, 0.4, 0.3),
			},
		},
		{
			RequestID: "r2",
			UserID:    "u2",
			Timestamp: now.Add(time.Minute),
			Items: []dataset.ExposedItem{
				item("i1", 1, 0.6, 0.5),
				item("i2", 2, 0.3, 0.2),
			},
		},
	}
	outcomes := []dataset.Outcome{
		{RequestID: "r1", UserID: "u1", ItemID: "i1", EventType: "click", Value: 1, Timestamp: now.Add(5 * time.Second)},
		{RequestID: "r2", UserID: "u2", ItemID: "i2", EventType: "click", Value: 1, Timestamp: now.Add(time.Minute + 5*time.Second)},
	}

	meta := ReportMetadata{
		BinaryVersion:           "test",
		GitCommit:               "deadbeef",
		EffectiveConfig:         []byte(`{"test":true}`),
		InputDatasetFingerprint: "fingerprint",
	}

	use := OPEUsecase{
		Exposures: staticExposureReader{items: exposures},
		Outcomes:  staticOutcomeReader{items: outcomes},
		Reporter:  reportjson.Writer{},
		Clock:     fixedClock{t: now},
		Logger:    noopLogger{},
		Metadata:  meta,
	}

	itemCfg := OPEConfig{
		RewardEvent: "click",
		Unit:        "item",
	}
	itemReport, err := use.Run(t.Context(), itemCfg, filepathTemp(t, "ope_item.json"))
	if err != nil {
		t.Fatalf("item mode run failed: %v", err)
	}

	reqCfg := OPEConfig{
		RewardEvent:       "click",
		Unit:              "request",
		RewardAggregation: "sum",
	}
	reqReport, err := use.Run(t.Context(), reqCfg, filepathTemp(t, "ope_request.json"))
	if err != nil {
		t.Fatalf("request mode run failed: %v", err)
	}

	if itemReport.OPE == nil || reqReport.OPE == nil {
		t.Fatalf("expected OPE report")
	}
	if itemReport.OPE.Unit != "item" || reqReport.OPE.Unit != "request" {
		t.Fatalf("unit metadata mismatch")
	}
	if itemReport.Summary.CasesEvaluated == reqReport.Summary.CasesEvaluated {
		t.Fatalf("expected different case counts for item vs request")
	}
	if len(itemReport.OPE.Estimators) == 0 || len(reqReport.OPE.Estimators) == 0 {
		t.Fatalf("missing estimators")
	}
	if itemReport.OPE.Estimators[0].Value == reqReport.OPE.Estimators[0].Value {
		t.Fatalf("expected estimator values to differ between units")
	}
}

func item(id string, rank int, logProp, tgtProp float64) dataset.ExposedItem {
	return dataset.ExposedItem{
		ItemID:            id,
		Rank:              rank,
		LoggingPropensity: &logProp,
		TargetPropensity:  &tgtProp,
	}
}

func filepathTemp(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name)
}
