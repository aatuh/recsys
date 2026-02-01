package files

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/fsutil"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/events"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
)

type FSCanonicalStore struct {
	baseDir string
}

var _ datasource.CanonicalStore = (*FSCanonicalStore)(nil)

func NewFSCanonicalStore(baseDir string) *FSCanonicalStore {
	return &FSCanonicalStore{baseDir: baseDir}
}

func (s *FSCanonicalStore) ReplaceExposureEvents(
	ctx context.Context,
	tenant string,
	surface string,
	day time.Time,
	evs []events.ExposureEvent,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path := s.dayFile(tenant, surface, day)
	if len(evs) == 0 {
		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		return nil
	}

	file, commit, rollback, err := fsutil.CreateAtomicWriter(path, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = rollback() }()

	w := bufio.NewWriter(file)
	enc := json.NewEncoder(w)
	for _, e := range evs {
		if err := enc.Encode(e); err != nil {
			_ = rollback()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		_ = rollback()
		return err
	}
	if err := commit(); err != nil {
		return err
	}
	return nil
}

func (s *FSCanonicalStore) ReadExposureEvents(
	ctx context.Context,
	tenant string,
	surface string,
	w windows.Window,
) (<-chan events.ExposureEvent, <-chan error) {
	out := make(chan events.ExposureEvent, 256)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)

		startDay := time.Date(w.Start.Year(), w.Start.Month(), w.Start.Day(), 0, 0, 0, 0, time.UTC)
		endDay := time.Date(w.End.Year(), w.End.Month(), w.End.Day(), 0, 0, 0, 0, time.UTC)

		for day := startDay; day.Before(endDay); day = day.Add(24 * time.Hour) {
			path := s.dayFile(tenant, surface, day)
			if _, err := os.Stat(path); err != nil {
				continue
			}
			if err := s.readDay(ctx, path, w, out); err != nil {
				errs <- err
				return
			}
		}
	}()

	return out, errs
}

func (s *FSCanonicalStore) dayFile(tenant, surface string, day time.Time) string {
	return filepath.Join(s.baseDir, tenant, surface, "exposures", day.Format("2006-01-02")+".jsonl")
}

func (s *FSCanonicalStore) readDay(
	ctx context.Context,
	path string,
	w windows.Window,
	out chan<- events.ExposureEvent,
) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		var e events.ExposureEvent
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			return fmt.Errorf("decode canonical exposure: %w", err)
		}
		if w.Contains(e.TS.UTC()) {
			out <- e
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
