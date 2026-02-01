package datasource

import (
	"context"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
)

// ExposureReader loads exposure events.
type ExposureReader interface {
	Read(ctx context.Context) ([]dataset.Exposure, error)
}

// ExposureStreamReader streams exposure events in request_id order.
type ExposureStreamReader interface {
	Stream(ctx context.Context) (<-chan dataset.Exposure, <-chan error)
}

// OutcomeReader loads outcome events.
type OutcomeReader interface {
	Read(ctx context.Context) ([]dataset.Outcome, error)
}

// OutcomeStreamReader streams outcome events in request_id order.
type OutcomeStreamReader interface {
	Stream(ctx context.Context) (<-chan dataset.Outcome, <-chan error)
}

// AssignmentReader loads experiment assignment events.
type AssignmentReader interface {
	Read(ctx context.Context) ([]dataset.Assignment, error)
}

// RankListReader loads ranked lists for interleaving.
type RankListReader interface {
	Read(ctx context.Context) ([]dataset.RankList, error)
}
