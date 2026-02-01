package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aatuh/recsys-suite/api/docs"
	"github.com/aatuh/recsys-suite/api/internal/config"
	appmw "github.com/aatuh/recsys-suite/api/internal/http/middleware"
	"github.com/aatuh/recsys-suite/api/migrations"

	"github.com/aatuh/api-toolkit-contrib/adapters/logzap"
	"github.com/aatuh/api-toolkit-contrib/bootstrap"
	"github.com/aatuh/api-toolkit-contrib/middleware/metrics"
	"github.com/aatuh/api-toolkit-contrib/telemetry"
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
	log.Info("recsys service starting", "version", version, "commit", commit, "date", date)

	ctx := context.Background()
	traceShutdown, tracingEnabled, traceErr := telemetry.InitTracing(ctx, cfg.Telemetry.Tracing)
	if traceErr != nil {
		log.Warn("tracing init failed", "err", traceErr.Error())
	} else if tracingEnabled {
		defer func() {
			_ = traceShutdown(context.Background())
		}()
	}

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

	securityStack, err := appmw.NewSecurityStack(ctx, cfg, log, pool)
	if err != nil {
		log.Error("failed to initialize security stack", "err", err)
		os.Exit(1)
	}
	if len(securityStack.Middlewares) > 0 {
		r.Use(securityStack.Middlewares...)
	}
	if len(securityStack.HealthChecks) > 0 {
		healthManager.RegisterCheckers(securityStack.HealthChecks...)
	}
	defer securityStack.Close()

	var pprofHandler http.Handler
	if cfg.Performance.PprofEnabled {
		pprofHandler = http.DefaultServeMux
	}
	bootstrap.MountSystemEndpoints(r, bootstrap.SystemEndpoints{
		Health: health.NewHandler(healthManager),
		Version: versionep.NewHandler(versionep.Config{
			Path: specs.Version,
			Info: ports.VersionInfo{
				Version: version,
				Commit:  commit,
				Date:    date,
			},
		}),
		Pprof:   pprofHandler,
		Metrics: metrics.PrometheusHandler(),
	})
	if docsHandler != nil {
		mountDocsRoutes(r, docsHandler, docs.OpenAPIJSON, docs.OpenAPIYAML)
	}

	deps, err := buildAppDeps(log, pool, cfg)
	if err != nil {
		log.Error("failed to initialize app dependencies", "err", err)
		os.Exit(1)
	}
	if deps.Close != nil {
		defer deps.Close()
	}
	mountAppRoutes(r, log, deps)

	srvCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	bootstrap.StartServerOrExit(srvCtx, cfg.Addr, r, log)
}
