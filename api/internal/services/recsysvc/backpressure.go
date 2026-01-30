package recsysvc

import (
	"context"
	"errors"
	"time"
)

// BoundedQueue limits concurrent work and queued requests.
type BoundedQueue struct {
	inFlight    chan struct{}
	queue       chan struct{}
	waitTimeout time.Duration
}

// NewBoundedQueue constructs a queue. maxInFlight <= 0 disables limiting.
// maxQueue controls how many requests may wait for a slot.
func NewBoundedQueue(maxInFlight, maxQueue int, waitTimeout time.Duration) *BoundedQueue {
	if maxInFlight <= 0 {
		return &BoundedQueue{}
	}
	q := &BoundedQueue{
		inFlight:    make(chan struct{}, maxInFlight),
		waitTimeout: waitTimeout,
	}
	if maxQueue > 0 {
		q.queue = make(chan struct{}, maxQueue)
	}
	return q
}

// Enabled returns true when limiting is active.
func (q *BoundedQueue) Enabled() bool {
	return q != nil && q.inFlight != nil
}

// Acquire waits for capacity or returns ErrOverloaded when queueing is full.
func (q *BoundedQueue) Acquire(ctx context.Context) error {
	if q == nil || q.inFlight == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if q.queue == nil && q.waitTimeout <= 0 {
		select {
		case q.inFlight <- struct{}{}:
			return nil
		default:
			return ErrOverloaded
		}
	}
	if q.queue != nil {
		select {
		case q.queue <- struct{}{}:
		default:
			return ErrOverloaded
		}
	}
	waitCtx := ctx
	var cancel func()
	if q.waitTimeout > 0 {
		waitCtx, cancel = context.WithTimeout(ctx, q.waitTimeout)
	}
	if cancel != nil {
		defer cancel()
	}
	select {
	case q.inFlight <- struct{}{}:
		if q.queue != nil {
			<-q.queue
		}
		return nil
	case <-waitCtx.Done():
		if q.queue != nil {
			<-q.queue
		}
		if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
			return ErrOverloaded
		}
		return waitCtx.Err()
	}
}

// Release releases an in-flight slot.
func (q *BoundedQueue) Release() {
	if q == nil || q.inFlight == nil {
		return
	}
	select {
	case <-q.inFlight:
	default:
	}
}
