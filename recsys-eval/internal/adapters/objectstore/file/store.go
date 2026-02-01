package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/objectstore"
)

// Store reads artifacts from the local filesystem.
type Store struct{}

func New() *Store {
	return &Store{}
}

func (s *Store) Get(_ context.Context, uri string) ([]byte, error) {
	if strings.TrimSpace(uri) == "" {
		return nil, fmt.Errorf("empty uri")
	}
	path := strings.TrimPrefix(uri, "file://")
	path = filepath.Clean(path)
	return os.ReadFile(path)
}

var _ objectstore.Store = (*Store)(nil)
