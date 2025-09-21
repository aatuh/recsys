package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"recsys/internal/audit"
	"recsys/internal/explain"
	"recsys/internal/http/common"
	"recsys/internal/http/config"
	"recsys/internal/http/db"
	"recsys/internal/http/handlers"
	httpmiddleware "recsys/internal/http/middleware"
	"recsys/internal/migrator"
	"recsys/internal/rules"
	"recsys/internal/store"
	"recsys/shared/util"
	"recsys/specs/endpoints"

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

	// Initialize error metrics
	errorMetrics := httpmiddleware.NewErrorMetrics()
	r.Use(httpmiddleware.ErrorMetricsMiddleware(errorMetrics))

	// Start error metrics logging
	go httpmiddleware.LogErrorMetrics(logger, errorMetrics, 5*time.Minute)

	// Global OPTIONS fallback to prevent 405 on unmatched preflights.
	// The CORS middleware should handle most cases, but this ensures any
	// path still returns 204 for OPTIONS.
	r.Options("/*", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r.Get(endpoints.Health, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// v1 endpoints
	st := store.New(pool)
	rulesManager := rules.NewManager(st, rules.ManagerOptions{
		RefreshInterval: cfg.RulesCacheRefresh,
		MaxPinSlots:     cfg.RulesMaxPinSlots,
		Enabled:         cfg.RulesEnabled,
	})

	explainService := &explain.Service{
		Collector: &explain.Collector{Store: st},
		Client:    explain.NullClient{},
		Config: explain.Config{
			Enabled:       cfg.LLMExplainEnabled,
			Provider:      cfg.LLMProvider,
			ModelPrimary:  cfg.LLMModelPrimary,
			ModelEscalate: cfg.LLMModelEscalate,
			Timeout:       cfg.LLMTimeout,
			MaxTokens:     cfg.LLMMaxTokens,
			CacheTTL:      30 * time.Minute,
		},
		Logger: logger,
	}
	if cfg.LLMExplainEnabled {
		httpClient := &http.Client{Timeout: cfg.LLMTimeout}
		if cfg.LLMTimeout <= 0 {
			httpClient.Timeout = 6 * time.Second
		}
		if strings.EqualFold(cfg.LLMProvider, "openai") && cfg.LLMAPIKey != "" {
			explainService.Client = &explain.OpenAIClient{
				HTTP:    httpClient,
				APIKey:  cfg.LLMAPIKey,
				BaseURL: cfg.LLMBaseURL,
				Logger:  logger,
			}
		}
	}
	decisionRecorder := audit.NewWriter(ctx, st, logger, audit.WriterConfig{
		Enabled:           cfg.DecisionTraceEnabled,
		QueueSize:         cfg.DecisionTraceQueueSize,
		BatchSize:         cfg.DecisionTraceBatchSize,
		FlushInterval:     cfg.DecisionTraceFlushInterval,
		SampleDefaultRate: cfg.DecisionTraceSampleDefault,
		NamespaceRates:    cfg.DecisionTraceNamespaceSamples,
	})
	defer func() {
		closeCtx, cancelClose := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelClose()
		if err := decisionRecorder.Close(closeCtx); err != nil && err != context.DeadlineExceeded {
			logger.Warn("decision recorder close", zap.Error(err))
		}
	}()
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
		RulesManager:        rulesManager,
		RulesAuditSample:    cfg.RulesAuditSample,
		ExplainService:      explainService,
		PurchasedWindowDays: cfg.PurchasedWindowDays,
		ProfileWindowDays:   cfg.ProfileWindowDays,
		ProfileBoost:        cfg.ProfileBoost,
		ProfileTopNTags:     cfg.ProfileTopNTags,
		BlendAlpha:          cfg.BlendAlpha,
		BlendBeta:           cfg.BlendBeta,
		BlendGamma:          cfg.BlendGamma,
		BanditAlgo:          cfg.BanditAlgo,
		Logger:              logger,
		DecisionRecorder:    decisionRecorder,
		DecisionTraceSalt:   cfg.DecisionTraceSalt,
	}
	r.Post(endpoints.ItemsUpsert, hs.ItemsUpsert)
	r.Post(endpoints.UsersUpsert, hs.UsersUpsert)
	r.Post(endpoints.EventsBatch, hs.EventsBatch)
	r.Post(endpoints.Recommendations, hs.Recommend)
	r.Get(endpoints.ItemsSimilar, hs.ItemSimilar)
	r.Post(endpoints.EventTypesUpsert, hs.EventTypesUpsert)
	r.Get(endpoints.EventTypes, hs.EventTypesList)

	// Data management endpoints
	r.Get(endpoints.UsersList, hs.ListUsers)
	r.Get(endpoints.ItemsList, hs.ListItems)
	r.Get(endpoints.EventsList, hs.ListEvents)
	r.Post(endpoints.UsersDelete, hs.DeleteUsers)
	r.Post(endpoints.ItemsDelete, hs.DeleteItems)
	r.Post(endpoints.EventsDelete, hs.DeleteEvents)

	// Segment profiles and segments
	r.Get(endpoints.SegmentProfiles, hs.SegmentProfilesList)
	r.Post(endpoints.SegmentProfilesUpsert, hs.SegmentProfilesUpsert)
	r.Post(endpoints.SegmentProfilesDelete, hs.SegmentProfilesDelete)
	r.Get(endpoints.Segments, hs.SegmentsList)
	r.Post(endpoints.SegmentsUpsert, hs.SegmentsUpsert)
	r.Post(endpoints.SegmentsDelete, hs.SegmentsDelete)
	r.Post(endpoints.SegmentDryRun, hs.SegmentsDryRun)

	// Audit endpoints
	r.Get(endpoints.AuditDecisions, hs.AuditDecisionsList)
	r.Get(endpoints.AuditDecisionByID, hs.AuditDecisionGet)
	r.Post(endpoints.AuditSearch, hs.AuditDecisionsSearch)

	// Bandit endpoints
	r.Post(endpoints.BanditPoliciesUpsert, hs.BanditPoliciesUpsert)
	r.Get(endpoints.BanditPolicies, hs.BanditPoliciesList)
	r.Post(endpoints.BanditDecide, hs.BanditDecide)
	r.Post(endpoints.BanditReward, hs.BanditReward)
	r.Post(endpoints.BanditRecommendations, hs.RecommendWithBandit)

	// Rule engine admin endpoints
	r.Post(endpoints.Rules, hs.RulesCreate)
	r.Put(endpoints.RuleByID, hs.RulesUpdate)
	r.Get(endpoints.Rules, hs.RulesList)
	r.Post(endpoints.RulesDryRun, hs.RulesDryRun)

	// Explain endpoints
	r.Post(endpoints.ExplainLLM, hs.ExplainLLM)

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
