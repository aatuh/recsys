package middleware

import (
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"recsys/internal/http/common"
)

// RateLimiter applies a token bucket rate limit per identifier (API key or IP).
type RateLimiter struct {
	limiters sync.Map
	rate     rate.Limit
	burst    int
	logger   *zap.Logger
}

// NewRateLimiter constructs a limiter for the provided requests-per-minute and burst.
func NewRateLimiter(requestsPerMinute, burst int, logger *zap.Logger) *RateLimiter {
	if requestsPerMinute <= 0 || burst <= 0 {
		return nil
	}
	r := rate.Every(time.Minute / time.Duration(requestsPerMinute))
	return &RateLimiter{rate: r, burst: burst, logger: logger}
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	if key == "" {
		key = "anonymous"
	}
	if limiter, ok := rl.limiters.Load(key); ok {
		return limiter.(*rate.Limiter)
	}
	limiter := rate.NewLimiter(rl.rate, rl.burst)
	actual, _ := rl.limiters.LoadOrStore(key, limiter)
	return actual.(*rate.Limiter)
}

// Middleware returns the HTTP middleware enforcing per-identifier rate limits.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	if rl == nil {
		return next
	}
	logger := rl.logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		identifier, ok := APIKeyFromContext(r.Context())
		if !ok || identifier == "" {
			identifier = r.RemoteAddr
		}

		limiter := rl.getLimiter(identifier)
		if limiter.Allow() {
			next.ServeHTTP(w, r)
			return
		}

		logger.Warn("rate limit exceeded", zap.String("identifier", identifier), zap.String("path", r.URL.Path))
		common.TooManyRequests(w, r)
	})
}
