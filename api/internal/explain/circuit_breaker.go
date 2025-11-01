package explain

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrCircuitOpen indicates the provider circuit is open.
var ErrCircuitOpen = errors.New("llm circuit breaker open")

// WithCircuitBreaker wraps the provided client with a circuit breaker according to cfg.
func WithCircuitBreaker(client LLMClient, cfg CircuitBreakerConfig, logger *zap.Logger) LLMClient {
	if client == nil {
		client = NullClient{}
	}
	if !cfg.Enabled {
		return client
	}
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 3
	}
	if cfg.ResetAfter <= 0 {
		cfg.ResetAfter = time.Minute
	}
	if cfg.HalfOpenSuccesses <= 0 {
		cfg.HalfOpenSuccesses = 1
	}
	return &breakerClient{
		next:    client,
		breaker: newCircuitBreaker(cfg, logger),
	}
}

type breakerClient struct {
	next    LLMClient
	breaker *circuitBreaker
}

func (b *breakerClient) Generate(ctx context.Context, model string, systemPrompt string, userPrompt string, maxTokens int) (string, error) {
	if b.breaker == nil {
		return b.next.Generate(ctx, model, systemPrompt, userPrompt, maxTokens)
	}
	if err := b.breaker.allow(); err != nil {
		return "", err
	}
	resp, err := b.next.Generate(ctx, model, systemPrompt, userPrompt, maxTokens)
	if err != nil {
		b.breaker.failure(err)
		return "", err
	}
	b.breaker.success()
	return resp, nil
}

type breakerState uint8

const (
	stateClosed breakerState = iota
	stateOpen
	stateHalfOpen
)

type circuitBreaker struct {
	cfg               CircuitBreakerConfig
	logger            *zap.Logger
	mu                sync.Mutex
	state             breakerState
	failures          int
	lastOpened        time.Time
	halfOpenSuccesses int
}

func newCircuitBreaker(cfg CircuitBreakerConfig, logger *zap.Logger) *circuitBreaker {
	return &circuitBreaker{
		cfg:    cfg,
		logger: logger,
	}
}

func (c *circuitBreaker) allow() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case stateOpen:
		if time.Since(c.lastOpened) < c.cfg.ResetAfter {
			return ErrCircuitOpen
		}
		c.state = stateHalfOpen
		c.halfOpenSuccesses = 0
		return nil
	default:
		return nil
	}
}

func (c *circuitBreaker) success() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.failures = 0
	switch c.state {
	case stateHalfOpen:
		c.halfOpenSuccesses++
		if c.halfOpenSuccesses >= c.cfg.HalfOpenSuccesses {
			c.state = stateClosed
			c.halfOpenSuccesses = 0
		}
	default:
		c.state = stateClosed
	}
}

func (c *circuitBreaker) failure(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.failures++

	switch c.state {
	case stateHalfOpen:
		c.open(err)
	case stateClosed:
		if c.failures >= c.cfg.FailureThreshold {
			c.open(err)
		}
	}
}

func (c *circuitBreaker) open(err error) {
	c.state = stateOpen
	c.failures = 0
	c.halfOpenSuccesses = 0
	c.lastOpened = time.Now()
	if c.logger != nil {
		c.logger.Warn("llm circuit breaker opened", zap.Error(err))
	}
}
