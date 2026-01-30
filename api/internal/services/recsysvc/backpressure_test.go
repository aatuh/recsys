package recsysvc

import (
	"context"
	"testing"
	"time"
)

func TestBoundedQueueRejectsWhenFull(t *testing.T) {
	q := NewBoundedQueue(1, 0, 0)
	if err := q.Acquire(context.Background()); err != nil {
		t.Fatalf("expected acquire ok: %v", err)
	}
	if err := q.Acquire(context.Background()); err == nil {
		t.Fatalf("expected overload")
	}
	q.Release()
	if err := q.Acquire(context.Background()); err != nil {
		t.Fatalf("expected acquire ok after release: %v", err)
	}
}

func TestBoundedQueueTimeout(t *testing.T) {
	q := NewBoundedQueue(1, 1, 10*time.Millisecond)
	if err := q.Acquire(context.Background()); err != nil {
		t.Fatalf("expected acquire ok: %v", err)
	}
	if err := q.Acquire(context.Background()); err != ErrOverloaded {
		t.Fatalf("expected ErrOverloaded, got %v", err)
	}
}
