package usecase

import (
	"testing"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
)

func TestBuildDataQualityIncludesDuplicateExposureRequestIDs(t *testing.T) {
	stats := dataset.JoinStats{
		ExposureCount:               3,
		OutcomeCount:                1,
		ExposuresJoined:             1,
		OutcomesJoined:              1,
		DuplicateExposureRequestIDs: 1,
	}

	dq := buildDataQualityFromCounts(3, 0, 1, 0, stats, 0)

	if dq.JoinIntegrity.DuplicateExposureRequestIDs != 1 {
		t.Fatalf("DuplicateExposureRequestIDs = %d, want 1", dq.JoinIntegrity.DuplicateExposureRequestIDs)
	}
}
