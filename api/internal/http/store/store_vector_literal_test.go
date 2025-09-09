package store

import "testing"

// Sanity check for the pgvector text literal formatter. Ensures compact
// formatting and exact bracket/comma structure.
func TestVectorLiteral_Format(t *testing.T) {
	in := []float64{1, 0, -0.125, 3.1415926535}
	got := vectorLiteral(in)
	want := "[1,0,-0.125,3.1415926535]"
	if got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}
