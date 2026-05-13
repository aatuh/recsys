package pathsafe

import "testing"

func TestSegmentAllowsCleanIDs(t *testing.T) {
	got, err := Segment("surface", " home ")
	if err != nil {
		t.Fatalf("Segment() error = %v", err)
	}
	if got != "home" {
		t.Fatalf("Segment() = %q, want home", got)
	}
}

func TestSegmentRejectsTraversalAndSeparators(t *testing.T) {
	for _, value := range []string{"", "   ", ".", "..", "../x", "x/y", `x\y`, "x\x00y"} {
		t.Run(value, func(t *testing.T) {
			if _, err := Segment("surface", value); err == nil {
				t.Fatal("Segment() error = nil")
			}
		})
	}
}
