package signalstore

import (
	"context"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/signals"
)

// Store persists DB-backed signals for the online service.
type Store interface {
	UpsertItemTags(ctx context.Context, tenant, namespace string, items []signals.ItemTag) error
	UpsertPopularity(ctx context.Context, tenant, namespace string, day time.Time, items []signals.PopularityItem) error
	UpsertCooccurrence(ctx context.Context, tenant, namespace string, day time.Time, items []signals.CooccurrenceItem) error
}
