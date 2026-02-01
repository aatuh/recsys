package catalog

import (
	"context"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/signals"
)

// Reader loads item catalog/tag data from an external source.
type Reader interface {
	Read(ctx context.Context) ([]signals.ItemTag, error)
}
