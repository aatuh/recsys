package dataset

import "time"

// FilterExposuresByTime returns exposures within [start, end].
func FilterExposuresByTime(exposures []Exposure, start, end time.Time) []Exposure {
	if start.IsZero() && end.IsZero() {
		return exposures
	}
	filtered := make([]Exposure, 0, len(exposures))
	for _, e := range exposures {
		if inWindow(e.Timestamp, start, end) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// FilterOutcomesByTime returns outcomes within [start, end].
func FilterOutcomesByTime(outcomes []Outcome, start, end time.Time) []Outcome {
	if start.IsZero() && end.IsZero() {
		return outcomes
	}
	filtered := make([]Outcome, 0, len(outcomes))
	for _, o := range outcomes {
		if inWindow(o.Timestamp, start, end) {
			filtered = append(filtered, o)
		}
	}
	return filtered
}

func inWindow(ts, start, end time.Time) bool {
	if !start.IsZero() && ts.Before(start) {
		return false
	}
	if !end.IsZero() && ts.After(end) {
		return false
	}
	return true
}
