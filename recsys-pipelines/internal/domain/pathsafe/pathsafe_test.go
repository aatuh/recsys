package pathsafe

import "testing"

func TestSegmentRejectsTraversal(t *testing.T) {
	for _, value := range []string{"", " ", ".", "..", "../tenant", "tenant/surface", `tenant\surface`, "tenant\x00"} {
		t.Run(value, func(t *testing.T) {
			if _, err := Segment("tenant", value); err == nil {
				t.Fatal("Segment() error = nil")
			}
		})
	}
}

func TestRelativePathRejectsTraversal(t *testing.T) {
	for _, value := range []string{"", "../escape", "safe/../../escape", `safe\escape`, "safe/\x00"} {
		t.Run(value, func(t *testing.T) {
			if _, err := RelativePath("object key", value); err == nil {
				t.Fatal("RelativePath() error = nil")
			}
		})
	}
}

func TestRelativePathAllowsNestedKeys(t *testing.T) {
	got, err := RelativePath("object key", "/tenant/surface/object.json")
	if err != nil {
		t.Fatalf("RelativePath() error = %v", err)
	}
	if got != "tenant/surface/object.json" {
		t.Fatalf("RelativePath() = %q", got)
	}
}
