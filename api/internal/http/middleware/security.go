package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aatuh/recsys-suite/api/internal/audit"
	"github.com/aatuh/recsys-suite/api/internal/config"
	"github.com/aatuh/recsys-suite/api/internal/store"

	"github.com/aatuh/api-toolkit/ports"
)

// SecurityStack bundles configured middleware and cleanup hooks.
type SecurityStack struct {
	Middlewares  []func(http.Handler) http.Handler
	Close        func()
	HealthChecks []ports.HealthChecker
}

// NewSecurityStack builds the security middleware stack for the service.
func NewSecurityStack(ctx context.Context, cfg config.Config, log ports.Logger, pool ports.DatabasePool) (SecurityStack, error) {
	middlewares := []func(http.Handler) http.Handler{}
	healthChecks := []ports.HealthChecker{}
	closers := []func(){}

	appendMw := func(mw func(http.Handler) http.Handler) {
		if mw != nil {
			middlewares = append(middlewares, mw)
		}
	}

	appendMw(EnsureRequestIDHeader)

	authSetup, err := NewAuthSetup(ctx, cfg.Auth, log)
	if err != nil {
		return SecurityStack{}, err
	}
	appendMw(authSetup.Middleware)
	if authSetup.Close != nil {
		closers = append(closers, authSetup.Close)
	}
	if authSetup.HealthChecker != nil {
		healthChecks = append(healthChecks, authSetup.HealthChecker)
	}

	appendMw(NewClaimsMiddleware(cfg.Auth, log).Handler)

	if cfg.Auth.APIKeys.Enabled && pool == nil {
		return SecurityStack{}, fmt.Errorf("api key auth enabled but database is not configured")
	}
	apiKeyStore := store.NewAPIKeyStore(pool)
	apiKeyMw, err := NewAPIKeyMiddleware(cfg.Auth, apiKeyStore, log)
	if err != nil {
		return SecurityStack{}, err
	}
	appendMw(apiKeyMw.Handler)

	appendMw(NewRequireAuth(cfg.Auth).Handler)
	appendMw(NewRequireTenantClaim(cfg.Auth).Handler)

	tenantMw, err := NewTenantMiddleware(cfg.Auth)
	if err != nil {
		return SecurityStack{}, err
	}
	appendMw(tenantMw.Handler)

	tenantRL, err := NewTenantRateLimitMiddleware(cfg.RateLimit, log)
	if err != nil {
		return SecurityStack{}, err
	}
	appendMw(tenantRL.Handler)

	appendMw(NewAdminRoleMiddleware(cfg.Auth).Handler)

	var auditLogger audit.Logger
	if cfg.Audit.Enabled {
		logger, err := audit.NewFileLogger(audit.FileLoggerOptions{
			Path:  cfg.Audit.Path,
			Fsync: cfg.Audit.Fsync,
		})
		if err != nil {
			return SecurityStack{}, err
		}
		auditLogger = logger
		closers = append(closers, func() {
			_ = logger.Close()
		})
	}
	appendMw(NewAuditMiddleware(cfg.Audit, auditLogger).Handler)

	closeFn := func() {
		for i := len(closers) - 1; i >= 0; i-- {
			closers[i]()
		}
	}

	return SecurityStack{
		Middlewares:  middlewares,
		Close:        closeFn,
		HealthChecks: healthChecks,
	}, nil
}
