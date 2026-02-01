package windows

import (
	"testing"
	"time"
)

func TestWindowValidate(t *testing.T) {
	w := Window{}
	if w.Validate() == nil {
		t.Fatalf("expected error for zero window")
	}
	w = Window{
		Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	if err := w.Validate(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
