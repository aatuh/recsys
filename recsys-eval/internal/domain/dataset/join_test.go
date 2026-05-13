package dataset

import "testing"

func TestJoinByRequestDetectsDuplicateExposureRequestIDs(t *testing.T) {
	exposures := []Exposure{
		{RequestID: "req-1", Items: []ExposedItem{{ItemID: "a"}}},
		{RequestID: "req-1", Items: []ExposedItem{{ItemID: "b"}}},
		{RequestID: "req-2", Items: []ExposedItem{{ItemID: "c"}}},
	}
	outcomes := []Outcome{{RequestID: "req-1", ItemID: "a"}}

	joined, stats := JoinByRequest(exposures, outcomes)

	if stats.DuplicateExposureRequestIDs != 1 {
		t.Fatalf("DuplicateExposureRequestIDs = %d, want 1", stats.DuplicateExposureRequestIDs)
	}
	if got := joined["req-1"].Exposure.Items[0].ItemID; got != "a" {
		t.Fatalf("joined duplicate request kept item %q, want first exposure item a", got)
	}
	if stats.ExposureCount != 3 {
		t.Fatalf("ExposureCount = %d, want 3", stats.ExposureCount)
	}
	if stats.ExposuresJoined != 1 {
		t.Fatalf("ExposuresJoined = %d, want 1", stats.ExposuresJoined)
	}
}

func TestJoinByRequestCleanDatasetHasNoDuplicateExposureRequestIDs(t *testing.T) {
	_, stats := JoinByRequest(
		[]Exposure{{RequestID: "req-1"}, {RequestID: "req-2"}},
		[]Outcome{{RequestID: "req-1", ItemID: "a"}},
	)

	if stats.DuplicateExposureRequestIDs != 0 {
		t.Fatalf("DuplicateExposureRequestIDs = %d, want 0", stats.DuplicateExposureRequestIDs)
	}
}
