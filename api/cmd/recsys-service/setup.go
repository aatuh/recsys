package main

import (
	"time"

	"recsys/internal/config"

	"github.com/aatuh/api-toolkit/endpoints/docs"
	"github.com/aatuh/api-toolkit/endpoints/health"
	"github.com/aatuh/api-toolkit/ports"
)

func setupDocsHandler(cfg config.Config, log ports.Logger) *docs.Handler {
	if !cfg.DocsEnabled {
		log.Info("docs endpoints disabled")
		return nil
	}
	return docs.NewHandler(docs.NewWithConfig(ports.DocsConfig{
		Title:       "Recsys Service API",
		Description: "Recommendation service API",
		Version:     "1.0.0",
		Paths:       ports.DefaultDocsPaths(),
		EnableHTML:  true,
		EnableJSON:  true,
		EnableYAML:  false,
	}))
}

func setupHealthManager(pool ports.DatabasePool, cfg config.Config) ports.HealthManager {
	readinessChecks := []string{"basic"}
	if pool != nil {
		readinessChecks = append(readinessChecks, "database")
	}
	manager := health.NewWithConfig(ports.HealthCheckConfig{
		Timeout:         5 * time.Second,
		CacheDuration:   cfg.Health.CacheDuration,
		EnableCaching:   true,
		EnableDetailed:  true,
		LivenessChecks:  []string{"basic"},
		ReadinessChecks: readinessChecks,
	})
	manager.RegisterChecker(health.NewBasicChecker())
	if pool != nil {
		manager.RegisterChecker(health.NewDatabaseChecker(pool))
	}
	return manager
}
