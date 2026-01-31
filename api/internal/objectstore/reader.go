package objectstore

import "context"

// ErrNotFound indicates the object does not exist.
type ErrNotFound struct {
	URI string
}

func (e ErrNotFound) Error() string {
	if e.URI == "" {
		return "object not found"
	}
	return "object not found: " + e.URI
}

// Reader fetches object blobs by URI.
type Reader interface {
	Get(ctx context.Context, uri string) ([]byte, error)
}
