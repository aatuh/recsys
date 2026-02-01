package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/checkpoint"
)

type Store struct {
	baseDir string
}

var _ checkpoint.Store = (*Store)(nil)

func New(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

func (s *Store) GetLastIngested(ctx context.Context, tenant, surface string) (time.Time, bool, error) {
	if err := ctx.Err(); err != nil {
		return time.Time{}, false, err
	}
	path := s.pathFor(tenant, surface)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, err
	}
	var payload struct {
		Day string `json:"day"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return time.Time{}, false, err
	}
	day, err := time.ParseInLocation("2006-01-02", payload.Day, time.UTC)
	if err != nil {
		return time.Time{}, false, err
	}
	return day, true, nil
}

func (s *Store) SetLastIngested(ctx context.Context, tenant, surface string, day time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path := s.pathFor(tenant, surface)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload := struct {
		Day string `json:"day"`
	}{
		Day: day.UTC().Format("2006-01-02"),
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

func (s *Store) pathFor(tenant, surface string) string {
	t := sanitize(tenant)
	surf := sanitize(surface)
	return filepath.Join(s.baseDir, t, fmt.Sprintf("%s.json", surf))
}

func sanitize(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "default"
	}
	value = strings.ReplaceAll(value, "/", "_")
	value = strings.ReplaceAll(value, "\\", "_")
	return value
}
