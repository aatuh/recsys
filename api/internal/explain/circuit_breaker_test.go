package explain

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestCircuitBreakerOpensAfterFailure(t *testing.T) {
	client := &sequenceClient{responses: []error{errors.New("boom")}}
	cfg := CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 1,
		ResetAfter:       20 * time.Millisecond,
	}

	wrapped := WithCircuitBreaker(client, cfg, zap.NewNop())

	if _, err := wrapped.Generate(context.Background(), "", "", "", 0); err == nil {
		t.Fatal("expected initial call to fail")
	}
	if _, err := wrapped.Generate(context.Background(), "", "", "", 0); !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected circuit open error, got %v", err)
	}

	time.Sleep(25 * time.Millisecond)
	client.append(nil)

	if _, err := wrapped.Generate(context.Background(), "", "", "", 0); err != nil {
		t.Fatalf("expected breaker to allow after reset, got %v", err)
	}
}

func TestCircuitBreakerDisabledReturnsOriginalClient(t *testing.T) {
	client := &sequenceClient{}
	wrapped := WithCircuitBreaker(client, CircuitBreakerConfig{}, zap.NewNop())
	if wrapped != client {
		t.Fatal("expected disabled breaker to return original client")
	}
}

type sequenceClient struct {
	mu        sync.Mutex
	responses []error
}

func (s *sequenceClient) Generate(ctx context.Context, model string, systemPrompt string, userPrompt string, maxTokens int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.responses) == 0 {
		return "ok", nil
	}
	err := s.responses[0]
	s.responses = s.responses[1:]
	if err != nil {
		return "", err
	}
	return "ok", nil
}

func (s *sequenceClient) append(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.responses = append(s.responses, err)
}
