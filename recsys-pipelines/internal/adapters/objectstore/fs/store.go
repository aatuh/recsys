package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/fsutil"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/objectstore"
)

type FSObjectStore struct {
	baseDir string
}

var _ objectstore.ObjectStore = (*FSObjectStore)(nil)

func New(baseDir string) *FSObjectStore {
	return &FSObjectStore{baseDir: baseDir}
}

func (s *FSObjectStore) Put(ctx context.Context, key string, _ string, data []byte) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	key = strings.TrimPrefix(key, "/")
	path := filepath.Join(s.baseDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := fsutil.WriteFileAtomic(path, data, 0o644); err != nil {
		return "", err
	}
	return "file://" + path, nil
}

func (s *FSObjectStore) Get(ctx context.Context, uri string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if !strings.HasPrefix(uri, "file://") {
		return nil, fmt.Errorf("unsupported uri: %s", uri)
	}
	path := strings.TrimPrefix(uri, "file://")
	return os.ReadFile(path)
}
