package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/api/internal/auth"
	"github.com/aatuh/recsys-suite/api/internal/config"

	"github.com/aatuh/api-toolkit/v2/authorization"
	"github.com/aatuh/api-toolkit/v2/middleware/ratelimit"
	"github.com/aatuh/api-toolkit/v2/ports"
)

// TenantRateLimitMiddleware enforces per-tenant rate limits.
type TenantRateLimitMiddleware struct {
	mw *ratelimit.Middleware
}

// NewTenantRateLimitMiddleware constructs the middleware.
func NewTenantRateLimitMiddleware(cfg config.RateLimitConfig, log ports.Logger) (*TenantRateLimitMiddleware, error) {
	if !cfg.TenantEnabled {
		return &TenantRateLimitMiddleware{}, nil
	}
	opts := ratelimit.Options{
		Capacity:   cfg.TenantCapacity,
		RefillRate: cfg.TenantRefillRate,
		RetryAfter: cfg.TenantRetryAfter,
		FailOpen:   cfg.TenantFailOpen,
		Key: func(r *http.Request) string {
			tenant, _ := authorization.TenantIDFromContext(r.Context())
			user, _ := auth.UserIDFromContext(r.Context())
			if tenant != "" && user != "" {
				return "tenant:" + tenant + ":user:" + user
			}
			if tenant != "" {
				return "tenant:" + tenant
			}
			if user != "" {
				return "user:" + user
			}
			if r == nil {
				return ""
			}
			if ip := strings.TrimSpace(r.RemoteAddr); ip != "" {
				return ip
			}
			return ""
		},
		OnError: func(err error) {
			if log != nil {
				log.Warn("tenant rate limit error", "err", err.Error())
			}
		},
	}
	mw, err := ratelimit.New(opts)
	if err != nil {
		return nil, err
	}
	return &TenantRateLimitMiddleware{mw: mw}, nil
}

// Handler wraps the next handler with tenant-based rate limiting.
func (m *TenantRateLimitMiddleware) Handler(next http.Handler) http.Handler {
	if m == nil || m.mw == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isProtectedPath(r) {
			next.ServeHTTP(w, r)
			return
		}
		m.mw.Handler(next).ServeHTTP(w, r)
	})
}

// DefaultTenantRateLimitConfig returns baseline defaults when no env config is provided.
func DefaultTenantRateLimitConfig() config.RateLimitConfig {
	return config.RateLimitConfig{
		TenantEnabled:    true,
		TenantCapacity:   60,
		TenantRefillRate: 30,
		TenantRetryAfter: time.Second,
		TenantFailOpen:   false,
	}
}
