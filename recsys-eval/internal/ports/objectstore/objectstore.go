package objectstore

import "context"

// Store fetches artifact blobs by URI.
type Store interface {
	Get(ctx context.Context, uri string) ([]byte, error)
}
