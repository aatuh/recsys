package jsonl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/events"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
)

type FSRawEventSource struct {
	baseDir string
}

var _ datasource.RawEventSource = (*FSRawEventSource)(nil)

func New(baseDir string) *FSRawEventSource {
	return &FSRawEventSource{baseDir: baseDir}
}

func (s *FSRawEventSource) ReadExposureEvents(
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

		files := s.resolveFiles(tenant, surface, w)
		if len(files) == 0 {
			// No input is not an error: upstream systems may legitimately have no
			// events for a window. The pipeline must remain idempotent and treat
			// this as an empty stream.
			return
		}

		for _, fp := range files {
			if err := s.readFile(ctx, fp, w, out); err != nil {
				errs <- err
				return
			}
		}
	}()

	return out, errs
}

func (s *FSRawEventSource) resolveFiles(tenant, surface string, w windows.Window) []string {
	flat := filepath.Join(s.baseDir, "exposure.jsonl")
	if _, err := os.Stat(flat); err == nil {
		return []string{flat}
	}

	var files []string
	startDay := time.Date(w.Start.Year(), w.Start.Month(), w.Start.Day(), 0, 0, 0, 0, time.UTC)
	endDay := time.Date(w.End.Year(), w.End.Month(), w.End.Day(), 0, 0, 0, 0, time.UTC)
	for day := startDay; day.Before(endDay); day = day.Add(24 * time.Hour) {
		name := fmt.Sprintf("exposure.%s.jsonl", day.Format("2006-01-02"))
		fp := filepath.Join(s.baseDir, tenant, surface, name)
		if _, err := os.Stat(fp); err == nil {
			files = append(files, fp)
		}
	}
	return files
}

func (s *FSRawEventSource) readFile(
	ctx context.Context,
	fp string,
	w windows.Window,
	out chan<- events.ExposureEvent,
) error {
	f, err := os.Open(fp)
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
			return fmt.Errorf("decode exposure jsonl: %w", err)
		}
		e = e.Normalized()
		if err := e.Validate(); err != nil {
			return fmt.Errorf("invalid exposure event: %w", err)
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
