package audit

import (
	"context"
	"errors"
	"testing"
	"time"

	"recsys/internal/store"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestWriterReportsUnhealthyAfterInsertFailure(t *testing.T) {
	errStore := &stubStore{responses: []error{errors.New("boom")}}
	rec := NewWriter(context.Background(), errStore, zap.NewNop(), WriterConfig{
		Enabled:           true,
		QueueSize:         1,
		BatchSize:         1,
		FlushInterval:     10 * time.Millisecond,
		SampleDefaultRate: 1.0,
	})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()
		_ = rec.Close(ctx)
	})

	rec.Record(sampleTrace())

	if !eventually(func() bool { return rec.Healthy() != nil }, 500*time.Millisecond) {
		t.Fatalf("expected recorder to become unhealthy")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if err := rec.Close(ctx); err == nil {
		t.Fatalf("expected close to surface failure")
	}
}

func TestWriterHealthRecoversAfterSuccess(t *testing.T) {
	stub := &stubStore{responses: []error{
		errors.New("first failure"),
		nil,
	}}
	rec := NewWriter(context.Background(), stub, zap.NewNop(), WriterConfig{
		Enabled:           true,
		QueueSize:         4,
		BatchSize:         1,
		FlushInterval:     10 * time.Millisecond,
		SampleDefaultRate: 1.0,
	})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()
		_ = rec.Close(ctx)
	})

	rec.Record(sampleTrace())
	if !eventually(func() bool { return rec.Healthy() != nil }, 500*time.Millisecond) {
		t.Fatalf("expected recorder to become unhealthy after failure")
	}

	rec.Record(sampleTrace())
	if !eventually(func() bool { return rec.Healthy() == nil }, 500*time.Millisecond) {
		t.Fatalf("expected recorder health to recover after success")
	}
}

func TestNoopRecorderHealthy(t *testing.T) {
	if err := (noopRecorder{}).Healthy(); err != nil {
		t.Fatalf("expected noop recorder to be healthy, got %v", err)
	}
}

type stubStore struct {
	responses []error
}

func (s *stubStore) InsertDecisionTraces(_ context.Context, rows []store.DecisionTraceInsert) error {
	if len(rows) == 0 {
		return nil
	}
	if len(s.responses) == 0 {
		return nil
	}
	err := s.responses[0]
	s.responses = s.responses[1:]
	return err
}

func sampleTrace() *Trace {
	orgID := uuid.New()
	return &Trace{
		DecisionID: uuid.New(),
		OrgID:      orgID.String(),
		Timestamp:  time.Now().UTC(),
		Namespace:  "ns",
		Config:     TraceConfig{},
		Candidates: []TraceCandidate{},
		FinalItems: []TraceFinalItem{},
	}
}

func eventually(check func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return check()
}
