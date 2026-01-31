package objectstore

import (
	"context"
	"fmt"
	"strings"
)

// MultiReader dispatches to scheme-specific readers.
type MultiReader struct {
	fs  *FSReader
	s3  *S3Reader
	max int
}

func NewMultiReader(fs *FSReader, s3 *S3Reader, maxBytes int) *MultiReader {
	return &MultiReader{fs: fs, s3: s3, max: maxBytes}
}

func (r *MultiReader) Get(ctx context.Context, uri string) ([]byte, error) {
	trimmed := strings.TrimSpace(uri)
	if trimmed == "" {
		return nil, fmt.Errorf("object uri is empty")
	}
	switch {
	case strings.HasPrefix(trimmed, "s3://"):
		if r.s3 == nil {
			return nil, fmt.Errorf("s3 reader not configured")
		}
		return r.s3.Get(ctx, trimmed)
	case strings.HasPrefix(trimmed, "file://"):
		if r.fs == nil {
			return nil, fmt.Errorf("fs reader not configured")
		}
		return r.fs.Get(ctx, trimmed)
	default:
		// Treat bare paths as filesystem.
		if r.fs != nil {
			return r.fs.Get(ctx, trimmed)
		}
		return nil, fmt.Errorf("unsupported uri: %s", trimmed)
	}
}
