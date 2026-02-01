package objectstore

import "context"

type ObjectStore interface {
	Put(ctx context.Context, key string, contentType string, data []byte) (uri string, err error)
	Get(ctx context.Context, uri string) ([]byte, error)
}
