package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/datasource"
)

type streamJoinResult struct {
	JoinStats       dataset.JoinStats
	ExposureMissing int
	OutcomeMissing  int
	ExposureMin     time.Time
	ExposureMax     time.Time
	OutcomeMin      time.Time
	OutcomeMax      time.Time
}

type streamJoinOptions struct {
	OnOutcome       func(dataset.Outcome)
	MaxOpenRequests int
}

func streamJoinByRequest(ctx context.Context, exposures datasource.ExposureStreamReader, outcomes datasource.OutcomeStreamReader, handle func(dataset.JoinedCase) error, opts streamJoinOptions) (streamJoinResult, error) {
	if exposures == nil || outcomes == nil {
		return streamJoinResult{}, fmt.Errorf("stream readers are required")
	}
	if opts.MaxOpenRequests <= 0 {
		opts.MaxOpenRequests = 1
	}
	if opts.MaxOpenRequests < 1 {
		return streamJoinResult{}, fmt.Errorf("stream.max_open_requests must be >= 1")
	}

	expCh, expErrCh := exposures.Stream(ctx)
	outCh, outErrCh := outcomes.Stream(ctx)

	res := streamJoinResult{}
	var pending dataset.Outcome
	pendingValid := false
	prevExpID := ""
	prevOutID := ""

	updateExposureTime := func(ts time.Time) {
		if ts.IsZero() {
			return
		}
		if res.ExposureMin.IsZero() || ts.Before(res.ExposureMin) {
			res.ExposureMin = ts
		}
		if res.ExposureMax.IsZero() || ts.After(res.ExposureMax) {
			res.ExposureMax = ts
		}
	}

	updateOutcomeTime := func(ts time.Time) {
		if ts.IsZero() {
			return
		}
		if res.OutcomeMin.IsZero() || ts.Before(res.OutcomeMin) {
			res.OutcomeMin = ts
		}
		if res.OutcomeMax.IsZero() || ts.After(res.OutcomeMax) {
			res.OutcomeMax = ts
		}
	}

	readOutcome := func() (dataset.Outcome, bool, error) {
		o, ok := <-outCh
		if !ok {
			return dataset.Outcome{}, false, nil
		}
		res.JoinStats.OutcomeCount++
		if o.RequestID == "" || o.ItemID == "" {
			res.OutcomeMissing++
		}
		if prevOutID != "" && o.RequestID < prevOutID {
			return dataset.Outcome{}, false, fmt.Errorf("outcomes must be sorted by request_id for stream mode")
		}
		prevOutID = o.RequestID
		updateOutcomeTime(o.Timestamp)
		if opts.OnOutcome != nil {
			opts.OnOutcome(o)
		}
		return o, true, nil
	}

	for exp := range expCh {
		res.JoinStats.ExposureCount++
		if exp.RequestID == "" || len(exp.Items) == 0 {
			res.ExposureMissing++
		}
		if prevExpID != "" && exp.RequestID < prevExpID {
			return res, fmt.Errorf("exposures must be sorted by request_id for stream mode")
		}
		prevExpID = exp.RequestID
		updateExposureTime(exp.Timestamp)

		var outs []dataset.Outcome
	loopOutcomes:
		for {
			var o dataset.Outcome
			if pendingValid {
				o = pending
				pendingValid = false
			} else {
				var ok bool
				var err error
				o, ok, err = readOutcome()
				if err != nil {
					return res, err
				}
				if !ok {
					break
				}
			}

			switch {
			case o.RequestID < exp.RequestID:
				continue
			case o.RequestID > exp.RequestID:
				pending = o
				pendingValid = true
				break loopOutcomes
			default:
				outs = append(outs, o)
			}
		}

		if len(outs) > 0 {
			res.JoinStats.ExposuresJoined++
			res.JoinStats.OutcomesJoined += len(outs)
		}
		if err := handle(dataset.JoinedCase{Exposure: exp, Outcomes: outs}); err != nil {
			return res, err
		}
	}

	for {
		o, ok, err := readOutcome()
		if err != nil {
			return res, err
		}
		if !ok {
			break
		}
		_ = o
	}

	if err := drainStreamError(expErrCh); err != nil {
		return res, err
	}
	if err := drainStreamError(outErrCh); err != nil {
		return res, err
	}
	return res, nil
}

func drainStreamError(errCh <-chan error) error {
	if errCh == nil {
		return nil
	}
	if err, ok := <-errCh; ok {
		return err
	}
	return nil
}
