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
		// Serve OpenAPI 3.0.3 for Swagger UI compatibility
		doc, err := generateOpenAPIJSON(serverCfg, "3.0.3")
		if err != nil {
			http.Error(w, "Failed to generate swagger", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(doc)
	})

	// Serve swagger.yaml file dynamically
	r.Get("/swagger/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		// Serve OpenAPI 3.0.3 for Swagger UI compatibility
		doc, err := generateOpenAPIYAML(serverCfg, "3.0.3")
		if err != nil {
			http.Error(w, "Failed to generate swagger", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(doc)
	})

	// OpenAPI 3.1.0 endpoints for external consumers
	r.Get("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		doc, err := generateOpenAPIJSON(serverCfg, "3.1.0")
		if err != nil {
			http.Error(w, "Failed to generate openapi", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(doc)
	})

	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		doc, err := generateOpenAPIYAML(serverCfg, "3.1.0")
		if err != nil {
			http.Error(w, "Failed to generate openapi", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(doc)
	})

	// v1 endpoints
	st := store.New(pool)
	hs := &handlers.Handler{
		Store:               st,
		DefaultOrg:          cfg.DefaultOrgID,
		HalfLifeDays:        cfg.HalfLifeDays,
		CoVisWindowDays:     cfg.CoVisWindowDays,
		PopularityFanout:    cfg.PopularityFanout,
		MMRLambda:           cfg.MMRLambda,
		BrandCap:            cfg.BrandCap,
		CategoryCap:         cfg.CategoryCap,
		RuleExcludeEvents:   cfg.RuleExcludeEvents,
		ExcludeEventTypes:   cfg.ExcludeEventTypes,
		BrandTagPrefixes:    cfg.BrandTagPrefixes,
		CategoryTagPrefixes: cfg.CategoryTagPrefixes,
		PurchasedWindowDays: cfg.PurchasedWindowDays,
		ProfileWindowDays:   cfg.ProfileWindowDays,
		ProfileBoost:        cfg.ProfileBoost,
		ProfileTopNTags:     cfg.ProfileTopNTags,
		BlendAlpha:          cfg.BlendAlpha,
		BlendBeta:           cfg.BlendBeta,
		BlendGamma:          cfg.BlendGamma,
		BanditAlgo:          cfg.BanditAlgo,
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

func convertParametersToOpenAPI3(swaggerDoc map[string]interface{}) {
	paths, ok := swaggerDoc["paths"].(map[string]interface{})
	if !ok {
		return
	}

	for _, pathItem := range paths {
		pathItemMap, ok := pathItem.(map[string]interface{})
		if !ok {
			continue
		}

		for _, operation := range pathItemMap {
			operationMap, ok := operation.(map[string]interface{})
			if !ok {
				continue
			}

			parameters, ok := operationMap["parameters"].([]interface{})
			if !ok {
				continue
			}

			var newParameters []interface{}
			var requestBody map[string]interface{}

			for _, param := range parameters {
				paramMap, ok := param.(map[string]interface{})
				if !ok {
					continue
				}

				// Check if this is a body parameter (Swagger 2.0)
				if paramIn, ok := paramMap["in"].(string); ok && paramIn == "body" {
					// Convert body parameter to requestBody (OpenAPI 3.0+)
					requestBody = map[string]interface{}{
						"required": paramMap["required"],
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": paramMap["schema"],
							},
						},
					}
					if description, ok := paramMap["description"].(string); ok {
						requestBody["description"] = description
					}
				} else {
					// Keep non-body parameters as they are
					newParameters = append(newParameters, param)
				}
			}

			// Update the operation
			if len(newParameters) > 0 {
				operationMap["parameters"] = newParameters
			} else {
				delete(operationMap, "parameters")
			}

			if requestBody != nil {
				operationMap["requestBody"] = requestBody
			}
		}
	}
}

func generateOpenAPIJSON(cfg ServerConfig, version string) ([]byte, error) {
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

	if strings.HasPrefix(version, "3.") {
		// Convert to OpenAPI 3.x format
		swaggerDoc["openapi"] = version
		delete(swaggerDoc, "swagger")
		delete(swaggerDoc, "host")
		delete(swaggerDoc, "basePath")
		delete(swaggerDoc, "schemes")
		convertParametersToOpenAPI3(swaggerDoc)
	} else if version == "2.0" {
		// Keep Swagger 2.0 and override host
		swaggerDoc["swagger"] = "2.0"
		swaggerDoc["host"] = cfg.SwaggerHost
	}

	// Add the servers section (OpenAPI 3.x; ignored by Swagger 2.0)
	if len(cfg.SwaggerSchemes) > 0 {
		serverURL := cfg.SwaggerSchemes[0] + "://" + cfg.SwaggerHost
		swaggerDoc["servers"] = []map[string]interface{}{
			{
				"url":         serverURL,
				"description": "API",
			},
		}
	}

	return json.MarshalIndent(swaggerDoc, "", "    ")
}

func generateOpenAPIYAML(cfg ServerConfig, version string) ([]byte, error) {
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

	if strings.HasPrefix(version, "3.") {
		swaggerDoc["openapi"] = version
		delete(swaggerDoc, "swagger")
		delete(swaggerDoc, "host")
		delete(swaggerDoc, "basePath")
		delete(swaggerDoc, "schemes")
		convertParametersToOpenAPI3(swaggerDoc)
	} else if version == "2.0" {
		swaggerDoc["swagger"] = "2.0"
		swaggerDoc["host"] = cfg.SwaggerHost
	}

	if len(cfg.SwaggerSchemes) > 0 {
		serverURL := cfg.SwaggerSchemes[0] + "://" + cfg.SwaggerHost
		swaggerDoc["servers"] = []map[string]interface{}{
			{
				"url":         serverURL,
				"description": "API",
			},
		}
	}

	return yaml.Marshal(swaggerDoc)
}
