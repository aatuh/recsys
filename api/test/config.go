package test

import (
	"testing"

	"github.com/aatuh/api-toolkit-contrib/adapters/envvar"
)

// Config holds integration-test configuration loaded from environment variables.
type Config struct {
	APIHost     string
	DatabaseURL string
	AuthToken   string
	DevAuth     DevAuthConfig
}

// DevAuthConfig captures dev auth headers for tests.
type DevAuthConfig struct {
	Enabled       bool
	UserIDHeader  string
	TenantHeader  string
	UserIDValue   string
	TenantIDValue string
}

// MustLoadConfig loads integration-test configuration from the environment.
func MustLoadConfig(t *testing.T) Config {
	t.Helper()
	env := envvar.New()
	devEnabled := env.GetBoolOr("DEV_AUTH_ENABLED", false)
	return Config{
		APIHost:     env.MustGet("API_HOST"),
		DatabaseURL: env.MustGet("DATABASE_URL"),
		AuthToken:   env.GetOr("TEST_BEARER_TOKEN", ""),
		DevAuth: DevAuthConfig{
			Enabled:       devEnabled,
			UserIDHeader:  env.GetOr("DEV_AUTH_USER_ID_HEADER", "X-Dev-User-Id"),
			TenantHeader:  env.GetOr("DEV_AUTH_TENANT_HEADER", "X-Dev-Org-Id"),
			UserIDValue:   env.GetOr("TEST_DEV_USER_ID", "test-user"),
			TenantIDValue: env.GetOr("TEST_DEV_TENANT_ID", "test-tenant"),
		},
	}
}
