package dataset

// JoinedCase ties an exposure to its outcomes.
type JoinedCase struct {
	Exposure Exposure
	Outcomes []Outcome
}

// JoinStats summarizes join integrity.
type JoinStats struct {
	ExposureCount   int
	OutcomeCount    int
	ExposuresJoined int
	OutcomesJoined  int
}

// JoinByRequest groups outcomes by request_id and joins them to exposures.
func JoinByRequest(exposures []Exposure, outcomes []Outcome) (map[string]JoinedCase, JoinStats) {
	outByReq := make(map[string][]Outcome, len(outcomes))
	for _, o := range outcomes {
		outByReq[o.RequestID] = append(outByReq[o.RequestID], o)
	}

	joined := make(map[string]JoinedCase, len(exposures))
	stats := JoinStats{ExposureCount: len(exposures), OutcomeCount: len(outcomes)}

	for _, e := range exposures {
		outs := outByReq[e.RequestID]
		if len(outs) > 0 {
			stats.ExposuresJoined++
			stats.OutcomesJoined += len(outs)
		}
		joined[e.RequestID] = JoinedCase{Exposure: e, Outcomes: outs}
	}

	return joined, stats
}
