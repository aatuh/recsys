package datasource

import (
	"context"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/events"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

type RawEventSource interface {
	ReadExposureEvents(
		ctx context.Context,
		tenant string,
		surface string,
		w windows.Window,
	) (<-chan events.ExposureEvent, <-chan error)
}

type CanonicalStore interface {
	// ReplaceExposureEvents writes a full canonical partition for a day.
	//
	// The operation must be idempotent: multiple calls with the same input must
	// yield the same output without accumulating duplicates.
	//
	// For empty input, implementations should remove any existing partition.
	ReplaceExposureEvents(
		ctx context.Context,
		tenant string,
		surface string,
		day time.Time,
		events []events.ExposureEvent,
	) error

	ReadExposureEvents(
		ctx context.Context,
		tenant string,
		surface string,
		w windows.Window,
	) (<-chan events.ExposureEvent, <-chan error)
}
