package main

import (
	"os"
	"strings"

	"github.com/aatuh/recsys-suite/api/internal/config"

	"github.com/aatuh/api-toolkit/contrib/v2/adapters/chi"
	"github.com/aatuh/api-toolkit/contrib/v2/bootstrap"
	metricsmw "github.com/aatuh/api-toolkit/contrib/v2/middleware/metrics"
	"github.com/aatuh/api-toolkit/v2/httpx/identity"
	rateln "github.com/aatuh/api-toolkit/v2/middleware/ratelimit"
	"github.com/aatuh/api-toolkit/v2/ports"
)

func buildRouter(log ports.Logger, cfg config.Config) (ports.HTTPRouter, error) {
	r := chi.New()

	resolver := identity.Resolver{HeaderPolicy: identity.HeaderPolicyBoth}
	if raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES")); raw != "" {
		if prefixes, err := identity.ParseTrustedProxies(strings.Split(raw, ",")); err == nil {
			resolver.TrustedProxies = prefixes
		}
	}

	profile, err := bootstrap.ProfileStrictAPI(log,
		bootstrap.WithIdentityResolver(resolver),
		bootstrap.WithRateLimitOptions(rateln.Options{
			Capacity:                  30,
			RefillRate:                15,
			SkipEnabled:               cfg.RateLimitSkipEnabled,
			SkipHeader:                cfg.RateLimitSkipHeader,
			AllowDangerousDevBypasses: cfg.RateLimitAllowDangerousDevBypasses,
		}),
		bootstrap.WithMetricsRecorder(metricsmw.NewPrometheusRecorder(nil, nil)),
		bootstrap.WithCORSOptions(ports.CORSOptions{
			AllowedOrigins:   cfg.CORS.AllowedOrigins,
			AllowedMethods:   cfg.CORS.AllowedMethods,
			AllowedHeaders:   cfg.CORS.AllowedHeaders,
			AllowCredentials: cfg.CORS.AllowCredentials,
			MaxAge:           cfg.CORS.MaxAge,
		}),
	)
	if err != nil {
		return nil, err
	}
	profile.Apply(r)
	return r, nil
}
