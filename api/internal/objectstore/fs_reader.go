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
	rootDir  string
}

func NewFSReader(maxBytes int) *FSReader {
	return &FSReader{maxBytes: int64(maxBytes)}
}

// NewRootedFSReader constructs a reader confined to rootDir.
func NewRootedFSReader(rootDir string, maxBytes int) (*FSReader, error) {
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		return nil, fmt.Errorf("file root is required")
	}
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolve file root: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("stat file root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("file root is not a directory")
	}
	return &FSReader{maxBytes: int64(maxBytes), rootDir: abs}, nil
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
	if r.rootDir != "" {
		return r.getRooted(path, uri)
	}
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

func (r *FSReader) getRooted(path, uri string) ([]byte, error) {
	name, err := r.rootedName(path)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(r.rootDir)
	if err != nil {
		return nil, fmt.Errorf("open file root: %w", err)
	}
	defer func() { _ = root.Close() }()
	info, err := root.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound{URI: uri}
		}
		return nil, err
	}
	if r.maxBytes > 0 && info.Size() > r.maxBytes {
		return nil, fmt.Errorf("file too large: %d bytes", info.Size())
	}
	data, err := root.ReadFile(name)
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

func (r *FSReader) rootedName(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("file uri is empty")
	}
	if filepath.IsAbs(path) {
		rel, err := filepath.Rel(r.rootDir, filepath.Clean(path))
		if err != nil {
			return "", fmt.Errorf("file uri is outside configured root")
		}
		path = rel
	}
	path = filepath.Clean(path)
	if path == "." || filepath.IsAbs(path) || path == ".." || strings.HasPrefix(path, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("file uri is outside configured root")
	}
	return path, nil
}
