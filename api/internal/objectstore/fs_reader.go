package objectstore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FSReader reads artifacts from local filesystem paths or file:// URIs.
type FSReader struct {
	maxBytes int64
}

func NewFSReader(maxBytes int) *FSReader {
	return &FSReader{maxBytes: int64(maxBytes)}
}

func (r *FSReader) Get(ctx context.Context, uri string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	path := strings.TrimSpace(uri)
	if path == "" {
		return nil, fmt.Errorf("file uri is empty")
	}
	path = strings.TrimPrefix(path, "file://")
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound{URI: uri}
		}
		return nil, err
	}
	if r.maxBytes > 0 && info.Size() > r.maxBytes {
		return nil, fmt.Errorf("file too large: %d bytes", info.Size())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound{URI: uri}
		}
		return nil, err
	}
	if r.maxBytes > 0 && int64(len(data)) > r.maxBytes {
		return nil, fmt.Errorf("file too large: %d bytes", len(data))
	}
	return data, nil
}
