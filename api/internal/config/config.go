package config

import (
	"github.com/aatuh/api-toolkit-contrib/config"
	"github.com/aatuh/api-toolkit/endpoints/health"
)

// Config extends the base toolkit config with app-specific settings.
type Config struct {
	config.Config
	DocsEnabled bool
	Health      health.Config
	App         AppConfig
}

// AppConfig holds app-wide public-facing configuration.
type AppConfig struct {
	BaseURL    string
	ProjectTag string
}

// Load loads env vars into Config structure.
func Load() Config {
	loader := config.NewLoader()
	base, err := config.LoadFromEnv(loader)
	if err != nil {
		panic(err)
	}
	cfg := Config{
		Config:      base,
		DocsEnabled: loader.Bool("DOCS_ENABLED", !config.IsProduction(base.Env)),
		Health:      health.LoadConfig(loader),
		App: AppConfig{
			BaseURL:    loader.String("FRONTEND_BASE_URL", "http://localhost:3000"),
			ProjectTag: loader.String("PROJECT_TAG", ""),
		},
	}
	if err := loader.Err(); err != nil {
		panic(err)
	}
	return cfg
}
