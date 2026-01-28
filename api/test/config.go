package test

import (
	"testing"

	"github.com/aatuh/api-toolkit-contrib/adapters/envvar"
)

// Config holds integration-test configuration loaded from environment variables.
type Config struct {
	APIHost     string
	DatabaseURL string
}

// MustLoadConfig loads integration-test configuration from the environment.
func MustLoadConfig(t *testing.T) Config {
	t.Helper()
	env := envvar.New()
	return Config{
		APIHost:     env.MustGet("API_HOST"),
		DatabaseURL: env.MustGet("DATABASE_URL"),
	}
}
