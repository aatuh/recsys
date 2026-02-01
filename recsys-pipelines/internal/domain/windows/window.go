package windows

import (
	"fmt"
	"time"
)

type Window struct {
	Start time.Time
	End   time.Time
}

func (w Window) Validate() error {
	if w.Start.IsZero() || w.End.IsZero() {
		return fmt.Errorf("window start and end must be set")
	}
	if !w.End.After(w.Start) {
		return fmt.Errorf("window end must be after start")
	}
	return nil
}

func (w Window) Contains(t time.Time) bool {
	return (t.Equal(w.Start) || t.After(w.Start)) && t.Before(w.End)
}

func DayWindowUTC(day time.Time) Window {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	return Window{Start: start, End: start.Add(24 * time.Hour)}
}
