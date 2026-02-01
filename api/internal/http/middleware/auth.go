package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aatuh/recsys-suite/api/internal/config"

	jwtint "github.com/aatuh/api-toolkit-contrib/integrations/auth/jwt"
	"github.com/aatuh/api-toolkit-contrib/middleware/auth/devheaders"
	"github.com/aatuh/api-toolkit/ports"
)

// AuthSetup bundles auth middleware setup results.
type AuthSetup struct {
	Middleware    func(http.Handler) http.Handler
	Close         func()
	HealthChecker ports.HealthChecker
}

// NewAuthSetup constructs auth middleware based on configuration.
func NewAuthSetup(ctx context.Context, cfg config.AuthConfig, log ports.Logger) (AuthSetup, error) {
	if log == nil {
		log = ports.NopLogger{}
	}
	if cfg.JWT.Enabled && cfg.DevHeaders.Enabled {
		return AuthSetup{}, fmt.Errorf("auth config invalid: dev headers cannot be enabled with JWT")
	}
	if cfg.Required && !cfg.JWT.Enabled && !cfg.DevHeaders.Enabled && !cfg.APIKeys.Enabled {
		return AuthSetup{}, fmt.Errorf("auth required but no auth provider configured")
	}
	if cfg.JWT.Enabled {
		if err := validateJWKSConfig(cfg); err != nil {
			return AuthSetup{}, err
		}
		mw, err := jwtint.NewMiddleware(ctx, cfg.JWT, log)
		if err != nil {
			return AuthSetup{}, err
		}
		required := mw.Handler
		optional := mw.OptionalHandler
		return buildAuthSetup(cfg, required, optional, mw.Close, jwtint.HealthChecker(cfg.JWT, http.DefaultClient)), nil
	}
	if cfg.DevHeaders.Enabled {
		mw, err := devheaders.New(cfg.DevHeaders, log)
		if err != nil {
			return AuthSetup{}, err
		}
		required := mw.Handler
		optional := mw.OptionalHandler
		return buildAuthSetup(cfg, required, optional, nil, nil), nil
	}
	return buildAuthSetup(cfg, nil, nil, nil, nil), nil
}

func buildAuthSetup(
	cfg config.AuthConfig,
	required func(http.Handler) http.Handler,
	optional func(http.Handler) http.Handler,
	closeFn func(),
	hc ports.HealthChecker,
) AuthSetup {
	mw := func(next http.Handler) http.Handler {
		if next == nil {
			return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
		}
		handler := optional
		if handler == nil {
			handler = required
		}
		wrapped := next
		if handler != nil {
			wrapped = handler(next)
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isProtectedPath(r) {
				next.ServeHTTP(w, r)
				return
			}
			if handler == nil {
				next.ServeHTTP(w, r)
				return
			}
			wrapped.ServeHTTP(w, r)
		})
	}
	if closeFn == nil {
		closeFn = func() {}
	}
	return AuthSetup{
		Middleware:    mw,
		Close:         closeFn,
		HealthChecker: hc,
	}
}

func validateJWKSConfig(cfg config.AuthConfig) error {
	raw := strings.TrimSpace(cfg.JWT.JWKSURL)
	if raw == "" {
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("auth jwks url invalid: %w", err)
	}
	if parsed.Scheme != "https" && !cfg.AllowInsecureJWKS {
		return fmt.Errorf("auth jwks url must use https unless AUTH_ALLOW_INSECURE_JWKS is true")
	}
	if len(cfg.JWKSAllowedHosts) == 0 {
		return nil
	}
	host := strings.ToLower(parsed.Hostname())
	for _, allowed := range cfg.JWKSAllowedHosts {
		if strings.EqualFold(strings.TrimSpace(allowed), host) {
			return nil
		}
	}
	return fmt.Errorf("auth jwks host %q not in allowlist", host)
}
