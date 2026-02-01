package kafka

import (
	"context"
	"fmt"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/events"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
)

// RawEventSource is a placeholder for Kafka ingestion.
// It currently returns a clear error when invoked.
type RawEventSource struct {
	Brokers []string
	Topic   string
	GroupID string
}

var _ datasource.RawEventSource = (*RawEventSource)(nil)

func New(brokers []string, topic, groupID string) *RawEventSource {
	return &RawEventSource{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	}
}

func (s *RawEventSource) ReadExposureEvents(
	_ context.Context,
	_ string,
	_ string,
	_ windows.Window,
) (<-chan events.ExposureEvent, <-chan error) {
	out := make(chan events.ExposureEvent)
	errs := make(chan error, 1)
	close(out)
	errs <- fmt.Errorf("kafka raw source not implemented yet")
	close(errs)
	return out, errs
}
