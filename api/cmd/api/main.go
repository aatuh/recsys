package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"recsys/internal/http/common"
	"recsys/internal/http/config"
	"recsys/internal/http/db"
	"recsys/internal/http/handlers"
	httpmiddleware "recsys/internal/http/middleware"
	"recsys/internal/http/store"
	"recsys/shared/util"

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

// @host         localhost:8000
// @schemes      https

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

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP)
	r.Use(httpmiddleware.RequestLogger(logger))
	r.Use(httpmiddleware.JSONRecovererWithLogger(logger))
	r.Use(httpmiddleware.ErrorLogger(logger))

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

	// Serve swagger.json file
	r.Get("/swagger/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "swagger/swagger.json")
	})

	// v1 endpoints
	st := store.New(pool)
	if err := st.EnsureEventTypeDefaults(ctx); err != nil {
		logger.Fatal("failed to ensure event type defaults", zap.Error(err))
	}
	hs := &handlers.Handler{
		Store:                st,
		DefaultOrg:           cfg.DefaultOrgID,
		HalfLifeDays:         cfg.HalfLifeDays,
		PopularityWindowDays: cfg.PopularityWindowDays,
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
	}
	r.Post("/v1/items:upsert", hs.ItemsUpsert)
	r.Post("/v1/users:upsert", hs.UsersUpsert)
	r.Post("/v1/events:batch", hs.EventsBatch)
	r.Post("/v1/recommendations", hs.Recommend)
	r.Get("/v1/items/{item_id}/similar", hs.ItemSimilar)
	r.Post("/v1/event-types:upsert", hs.EventTypesUpsert)
	r.Get("/v1/event-types", hs.EventTypesList)

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
	Port string
}

func LoadServerConfig() ServerConfig {
	return ServerConfig{
		Port: util.MustGetEnv("API_PORT"),
	}
}

func pingDB(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return pool.Ping(ctx)
}
