package checkpoint

import (
	"context"
	"time"
)

// Store persists ingest checkpoints per tenant/surface.
type Store interface {
	GetLastIngested(ctx context.Context, tenant, surface string) (time.Time, bool, error)
	SetLastIngested(ctx context.Context, tenant, surface string, day time.Time) error
}
