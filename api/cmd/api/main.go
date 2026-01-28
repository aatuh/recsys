package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"strings"
	"time"

	"recsys/internal/config"
	"recsys/migrations"

	"github.com/aatuh/api-toolkit-contrib/adapters/logzap"
	"github.com/aatuh/api-toolkit-contrib/bootstrap"
	"github.com/aatuh/api-toolkit-contrib/middleware/metrics"
	"github.com/aatuh/api-toolkit/endpoints/health"
	versionep "github.com/aatuh/api-toolkit/endpoints/version"
	"github.com/aatuh/api-toolkit/ports"
	"github.com/aatuh/api-toolkit/specs"
)

var (
	// Overridden at build time via -ldflags.
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// @title API Boilerplate API
// @version 1.0.0
// @description REST API Boilerplate Documentation
// @BasePath /api/v1
func loggerFromEnv(env, level string) ports.Logger {
	env = strings.ToLower(strings.TrimSpace(env))
	switch env {
	case "development", "dev", "":
		return logzap.NewDevelopment(level)
	case "staging", "production", "prod":
		return logzap.NewProductionWithLevel(level)
	default:
		fmt.Printf("unknown ENV %q, defaulting to production logger\n", env)
		return logzap.NewProductionWithLevel(level)
	}
}

func main() {
	cfg := config.Load()
	log := loggerFromEnv(cfg.Env, cfg.LogLevel)

	ctx := context.Background()
	pool := bootstrap.OpenPoolOrExit(ctx, cfg.DatabaseURL, 3*time.Second, log)
	defer pool.Close()

	if cfg.MigrateOnStart {
		bootstrap.RunMigrationsOrExit(
			ctx, cfg.Config, log, []fs.FS{migrations.Migrations},
		)
	}

	r, err := bootstrap.NewDefaultRouter(log)
	if err != nil {
		log.Error("failed to initialize router", "err", err)
		os.Exit(1)
	}

	docsHandler := setupDocsHandler(cfg, log)
	healthManager := setupHealthManager(pool, cfg)

	bootstrap.MountSystemEndpoints(r, bootstrap.SystemEndpoints{
		Health: health.NewHandler(healthManager),
		Docs:   docsHandler,
		Version: versionep.NewHandler(versionep.Config{
			Path: specs.Version,
			Info: ports.VersionInfo{
				Version: version,
				Commit:  commit,
				Date:    date,
			},
		}),
		Metrics: metrics.PrometheusHandler(),
	})

	deps := buildAppDeps(log, pool)
	mountAppRoutes(r, log, deps)

	srvCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	bootstrap.StartServerOrExit(srvCtx, cfg.Addr, r, log)
}
