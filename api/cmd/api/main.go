package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"

	"recsys/internal/http/common"
	"recsys/internal/http/config"
	"recsys/internal/http/db"
	"recsys/internal/http/handlers"
	httpmiddleware "recsys/internal/http/middleware"
	"recsys/internal/migrator"
	"recsys/internal/store"
	"recsys/shared/util"
	"recsys/swagger"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// @title        Recsys API
// @version      0.0.1
// @description  Domain-agnostic recommendation service.
// @BasePath     /
//
// @host         localhost:8000
// @schemes      https
// Note: The host and schemes are dynamically configured at runtime using
// environment variables. Generated docs will use these above values.

func main() {
	logger, _ := zap.NewDevelopment()
	defer func() { _ = logger.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load all configurations
	serverCfg := LoadServerConfig()
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	debugCfg := common.LoadDebugConfig()

	// Configure Swagger dynamically based on environment variables
	configureSwagger(serverCfg)

	// Inject debug config into common package
	common.SetDebugConfig(debugCfg)

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if err := pingDB(ctx, pool); err != nil {
		logger.Fatal("db not reachable", zap.Error(err))
	}

	// Run migrations on start
	if util.MustGetEnv("MIGRATE_ON_START") == "true" {
		sqlDB, err := sql.Open("pgx", cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("open sql db: %v", err)
		}
		defer sqlDB.Close()

		r := migrator.New(sqlDB, nil, migrator.Options{
			MigrationsDir: util.MustGetEnv("MIGRATIONS_DIR"),
			Logger:        func(f string, a ...any) { log.Printf(f, a...) },
		})
		if err := r.Up(ctx); err != nil {
			log.Fatalf("migrate on start: %v", err)
		}
	}

	r := chi.NewRouter()
	r.Use(httpmiddleware.CORS())
	r.Use(middleware.RequestID, middleware.RealIP)
	r.Use(httpmiddleware.RequestLogger(logger))
	r.Use(httpmiddleware.JSONRecovererWithLogger(logger))
	r.Use(httpmiddleware.ErrorLogger(logger))

	// Global OPTIONS fallback to prevent 405 on unmatched preflights.
	// The CORS middleware should handle most cases, but this ensures any
	// path still returns 204 for OPTIONS.
	r.Options("/*", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Swagger UI (generated with `make swag`)
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})
	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"),
	))

	// Serve swagger.json file dynamically
	r.Get("/swagger/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Generate swagger JSON with dynamic servers section
		swaggerJSON, err := generateSwaggerWithServers(serverCfg)
		if err != nil {
			http.Error(w, "Failed to generate swagger", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(swaggerJSON)
	})

	// Serve swagger.yaml file dynamically
	r.Get("/swagger/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		// Generate swagger YAML with dynamic servers section
		swaggerYAML, err := generateSwaggerYAMLWithServers(serverCfg)
		if err != nil {
			http.Error(w, "Failed to generate swagger", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(swaggerYAML)
	})

	// v1 endpoints
	st := store.New(pool)
	hs := &handlers.Handler{
		Store:                st,
		DefaultOrg:           cfg.DefaultOrgID,
		HalfLifeDays:         cfg.HalfLifeDays,
		CoVisWindowDays:      cfg.CoVisWindowDays,
		PopularityFanout:     cfg.PopularityFanout,
		MMRLambda:            cfg.MMRLambda,
		BrandCap:             cfg.BrandCap,
		CategoryCap:          cfg.CategoryCap,
		RuleExcludePurchased: cfg.RuleExcludePurchased,
		PurchasedWindowDays:  cfg.PurchasedWindowDays,
		ProfileWindowDays:    cfg.ProfileWindowDays,
		ProfileBoost:         cfg.ProfileBoost,
		ProfileTopNTags:      cfg.ProfileTopNTags,
		BlendAlpha:           cfg.BlendAlpha,
		BlendBeta:            cfg.BlendBeta,
		BlendGamma:           cfg.BlendGamma,
		BanditAlgo:           cfg.BanditAlgo,
	}
	r.Post("/v1/items:upsert", hs.ItemsUpsert)
	r.Post("/v1/users:upsert", hs.UsersUpsert)
	r.Post("/v1/events:batch", hs.EventsBatch)
	r.Post("/v1/recommendations", hs.Recommend)
	r.Get("/v1/items/{item_id}/similar", hs.ItemSimilar)
	r.Post("/v1/event-types:upsert", hs.EventTypesUpsert)
	r.Get("/v1/event-types", hs.EventTypesList)

	// Data management endpoints
	r.Get("/v1/users", hs.ListUsers)
	r.Get("/v1/items", hs.ListItems)
	r.Get("/v1/events", hs.ListEvents)
	r.Post("/v1/users:delete", hs.DeleteUsers)
	r.Post("/v1/items:delete", hs.DeleteItems)
	r.Post("/v1/events:delete", hs.DeleteEvents)

	// Bandit endpoints
	r.Post("/v1/bandit/policies:upsert", hs.BanditPoliciesUpsert)
	r.Get("/v1/bandit/policies", hs.BanditPoliciesList)
	r.Post("/v1/bandit/decide", hs.BanditDecide)
	r.Post("/v1/bandit/reward", hs.BanditReward)
	r.Post("/v1/bandit/recommendations", hs.RecommendWithBandit)

	srv := &http.Server{
		Addr:              ":" + serverCfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	go func() {
		logger.Info("api listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

type ServerConfig struct {
	Port           string
	SwaggerHost    string
	SwaggerSchemes []string
}

func LoadServerConfig() ServerConfig {
	schemesStr := util.MustGetEnv("SWAGGER_SCHEMES")
	schemes := strings.Split(schemesStr, ",")
	// Trim whitespace from each scheme
	for i, scheme := range schemes {
		schemes[i] = strings.TrimSpace(scheme)
	}

	return ServerConfig{
		Port:           util.MustGetEnv("API_PORT"),
		SwaggerHost:    util.MustGetEnv("SWAGGER_HOST"),
		SwaggerSchemes: schemes,
	}
}

func pingDB(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return pool.Ping(ctx)
}

// configureSwagger sets the Swagger host, schemes, and server URL based on server configuration.
func configureSwagger(cfg ServerConfig) {
	swagger.SwaggerInfo.Host = cfg.SwaggerHost
	swagger.SwaggerInfo.Schemes = cfg.SwaggerSchemes
}

// generateSwaggerWithServers reads the static swagger.json and adds the servers section dynamically
func generateSwaggerWithServers(cfg ServerConfig) ([]byte, error) {
	// Read the static swagger.json file
	swaggerData, err := os.ReadFile("swagger/swagger.json")
	if err != nil {
		return nil, err
	}

	// Parse the JSON
	var swaggerDoc map[string]interface{}
	if err := json.Unmarshal(swaggerData, &swaggerDoc); err != nil {
		return nil, err
	}

	// Convert to OpenAPI 3.0+ format
	swaggerDoc["openapi"] = "3.0.0"
	delete(swaggerDoc, "swagger")  // Remove Swagger 2.0 version
	delete(swaggerDoc, "host")     // Remove host field (not used in OpenAPI 3.0+)
	delete(swaggerDoc, "basePath") // Remove basePath (not used in OpenAPI 3.0+)
	delete(swaggerDoc, "schemes")  // Remove schemes (not used in OpenAPI 3.0+)

	// Add the servers section
	if len(cfg.SwaggerSchemes) > 0 {
		serverURL := cfg.SwaggerSchemes[0] + "://" + cfg.SwaggerHost
		swaggerDoc["servers"] = []map[string]interface{}{
			{
				"url":         serverURL,
				"description": "API",
			},
		}
	}

	// Marshal back to JSON
	return json.MarshalIndent(swaggerDoc, "", "    ")
}

// generateSwaggerYAMLWithServers reads the static swagger.yaml and adds the servers section dynamically
func generateSwaggerYAMLWithServers(cfg ServerConfig) ([]byte, error) {
	// Read the static swagger.yaml file
	swaggerData, err := os.ReadFile("swagger/swagger.yaml")
	if err != nil {
		return nil, err
	}

	// Parse the YAML
	var swaggerDoc map[string]interface{}
	if err := yaml.Unmarshal(swaggerData, &swaggerDoc); err != nil {
		return nil, err
	}

	// Convert to OpenAPI 3.0+ format
	swaggerDoc["openapi"] = "3.0.0"
	delete(swaggerDoc, "swagger")  // Remove Swagger 2.0 version
	delete(swaggerDoc, "host")     // Remove host field (not used in OpenAPI 3.0+)
	delete(swaggerDoc, "basePath") // Remove basePath (not used in OpenAPI 3.0+)
	delete(swaggerDoc, "schemes")  // Remove schemes (not used in OpenAPI 3.0+)

	// Add the servers section
	if len(cfg.SwaggerSchemes) > 0 {
		serverURL := cfg.SwaggerSchemes[0] + "://" + cfg.SwaggerHost
		swaggerDoc["servers"] = []map[string]interface{}{
			{
				"url":         serverURL,
				"description": "API",
			},
		}
	}

	// Marshal back to YAML
	return yaml.Marshal(swaggerDoc)
}
