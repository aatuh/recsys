package usecase

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/decision"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type noopLogger struct{}

func (noopLogger) Infof(string, ...any)  {}
func (noopLogger) Errorf(string, ...any) {}

type staticExposureReader struct{ items []dataset.Exposure }

func (r staticExposureReader) Read(_ context.Context) ([]dataset.Exposure, error) {
	return r.items, nil
}

type staticOutcomeReader struct{ items []dataset.Outcome }

func (r staticOutcomeReader) Read(_ context.Context) ([]dataset.Outcome, error) { return r.items, nil }

type staticAssignmentReader struct{ items []dataset.Assignment }

func (r staticAssignmentReader) Read(_ context.Context) ([]dataset.Assignment, error) {
	return r.items, nil
}

type captureDecisionWriter struct {
	artifact *decision.Artifact
	path     string
}

func (w *captureDecisionWriter) Write(_ context.Context, artifact decision.Artifact, path string) error {
	w.artifact = &artifact
	w.path = path
	return nil
}

func projectRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve project root: %v", err)
	}
	return dir
}
