package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"recsys/internal/audit"
	"recsys/internal/config"
	"recsys/internal/explain"
	"recsys/internal/http/common"
	"recsys/internal/http/db"
	"recsys/internal/http/handlers"
	httpmiddleware "recsys/internal/http/middleware"
	"recsys/internal/migrator"
	"recsys/internal/observability"
	policymetrics "recsys/internal/observability/policy"
	"recsys/internal/rules"
	"recsys/internal/services/datamanagement"
	"recsys/internal/services/ingestion"
	manualsvc "recsys/internal/services/manual"
	"recsys/internal/services/recommendation"
	"recsys/internal/store"
	"recsys/specs/endpoints"
)

// Options configures the application composition root.
type Options struct {
	Config config.Config
	Logger *zap.Logger
}

// App wires up the API service dependencies and lifecycle.
type App struct {
	cfg              config.Config
	logger           *zap.Logger
	pool             *pgxpool.Pool
	decisionRecorder audit.Recorder
	server           *http.Server
	errorMetrics     *httpmiddleware.ErrorMetrics

	closers       []func(context.Context) error
	metricsCancel context.CancelFunc
	metricsWG     sync.WaitGroup
}

// New constructs the application graph based on the supplied options.
func New(ctx context.Context, opts Options) (*App, error) {
	cfg := opts.Config

	debugCfg := common.NewDebugConfig(cfg.Debug.Environment, cfg.Debug.AppDebug)

	logger := opts.Logger
	if logger == nil {
		var err error
		logger, err = buildLogger(debugCfg)
		if err != nil {
			return nil, fmt.Errorf("logger: %w", err)
		}
	}
	common.SetDebugConfig(debugCfg)

	dbCfg := db.Config{
		MaxConnIdle:       cfg.Database.MaxConnIdle,
		MaxConnLifetime:   cfg.Database.MaxConnLifetime,
		HealthCheckPeriod: cfg.Database.HealthCheckPeriod,
		AcquireTimeout:    cfg.Database.AcquireTimeout,
		MinConns:          cfg.Database.MinConns,
		MaxConns:          cfg.Database.MaxConns,
	}
	pool, err := db.NewPool(ctx, cfg.Database.URL, dbCfg)
	if err != nil {
		return nil, fmt.Errorf("db pool: %w", err)
	}

	var closers []func(context.Context) error
	closers = append(closers, func(ctx context.Context) error {
		_ = ctx
		pool.Close()
		return nil
	})

	if cfg.Migrations.RunOnStart {
		if err := runMigrations(ctx, cfg.Migrations, cfg.Database.URL, logger); err != nil {
			pool.Close()
			return nil, err
		}
	}

	st := store.NewWithOptions(pool, store.Options{
		QueryTimeout:        cfg.Database.QueryTimeout,
		RetryAttempts:       cfg.Database.RetryAttempts,
		RetryInitialBackoff: cfg.Database.RetryInitialBackoff,
		RetryMaxBackoff:     cfg.Database.RetryMaxBackoff,
	})
	ingestionSvc := ingestion.New(st)
	dataSvc := datamanagement.New(st)
	rulesManager := rules.NewManager(st, rules.ManagerOptions{
		RefreshInterval: cfg.Rules.CacheRefresh,
		MaxPinSlots:     cfg.Rules.MaxPinSlots,
		Enabled:         cfg.Rules.Enabled,
	})
	recommendationSvc := recommendation.New(st, rulesManager)
	if len(cfg.Recommendation.BlendOverrides) > 0 {
		entries := make(map[string]recommendation.ResolvedBlendConfig, len(cfg.Recommendation.BlendOverrides))
		now := time.Now().UTC()
		for ns, weights := range cfg.Recommendation.BlendOverrides {
			entries[ns] = recommendation.ResolvedBlendConfig{
				Namespace: ns,
				Alpha:     weights.Alpha,
				Beta:      weights.Beta,
				Gamma:     weights.Gamma,
				Source:    "static_config",
				UpdatedAt: now,
			}
		}
		if resolver := recommendation.NewStaticBlendResolver(entries); resolver != nil {
			recommendationSvc = recommendationSvc.WithBlendResolver(resolver)
		}
	}

	explainService := newExplainService(cfg.Explain, st, logger)

	writerCfg := audit.WriterConfig{
		Enabled:           cfg.Audit.DecisionTrace.Enabled,
		QueueSize:         cfg.Audit.DecisionTrace.QueueSize,
		BatchSize:         cfg.Audit.DecisionTrace.BatchSize,
		FlushInterval:     cfg.Audit.DecisionTrace.FlushInterval,
		SampleDefaultRate: cfg.Audit.DecisionTrace.SampleDefault,
		NamespaceRates:    cfg.Audit.DecisionTrace.NamespaceSample,
	}

	decisionRecorder := audit.NewWriter(ctx, st, logger, writerCfg)
	closers = append(closers, func(parent context.Context) error {
		closeCtx, cancel := context.WithTimeout(parent, 5*time.Second)
		defer cancel()
		if err := decisionRecorder.Close(closeCtx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			logger.Warn("decision recorder close", zap.Error(err))
			return err
		}
		return nil
	})

	tracer := handlers.NewDecisionTracer(decisionRecorder, logger, cfg.Audit.DecisionTrace.Salt, cfg.Rules.AuditSample)

	recConfig := handlers.RecommendationConfig{
		HalfLifeDays:               cfg.Recommendation.HalfLifeDays,
		CoVisWindowDays:            cfg.Recommendation.CoVisWindowDays,
		PopularityFanout:           cfg.Recommendation.PopularityFanout,
		MMRLambda:                  cfg.Recommendation.MMRLambda,
		BrandCap:                   cfg.Recommendation.BrandCap,
		CategoryCap:                cfg.Recommendation.CategoryCap,
		RuleExcludeEvents:          cfg.Recommendation.RuleExcludeEvents,
		ExcludeEventTypes:          cfg.Recommendation.ExcludeEventTypes,
		BrandTagPrefixes:           cfg.Recommendation.BrandTagPrefixes,
		CategoryTagPrefixes:        cfg.Recommendation.CategoryTagPrefixes,
		PurchasedWindowDays:        cfg.Recommendation.PurchasedWindowDays,
	ProfileWindowDays:          cfg.Recommendation.Profile.WindowDays,
	ProfileBoost:               cfg.Recommendation.Profile.Boost,
	ProfileTopNTags:            cfg.Recommendation.Profile.TopNTags,
	ProfileMinEventsForBoost:   cfg.Recommendation.Profile.MinEventsForBoost,
	ProfileColdStartMultiplier: cfg.Recommendation.Profile.ColdStartMultiplier,
	MMRPresets:                 cfg.Recommendation.MMRPresets,
	BlendAlpha:                 cfg.Recommendation.Blend.Alpha,
		BlendBeta:                  cfg.Recommendation.Blend.Beta,
		BlendGamma:                 cfg.Recommendation.Blend.Gamma,
		RulesEnabled:               cfg.Rules.Enabled,
		BanditExperiment: handlers.BanditExperimentConfig{
			Enabled:        cfg.Recommendation.BanditExperiment.Enabled,
			HoldoutPercent: cfg.Recommendation.BanditExperiment.HoldoutPercent,
			Label:          cfg.Recommendation.BanditExperiment.Label,
			Surfaces:       make(map[string]struct{}),
		},
	}
	for _, surface := range cfg.Recommendation.BanditExperiment.Surfaces {
		recConfig.BanditExperiment.Surfaces[surface] = struct{}{}
	}

	ingHandler := handlers.NewIngestionHandler(ingestionSvc, cfg.Recommendation.DefaultOrgID, logger)
	dataHandler := handlers.NewDataManagementHandler(dataSvc, cfg.Recommendation.DefaultOrgID, logger)
	segmentsHandler := handlers.NewSegmentsHandler(st, cfg.Recommendation.DefaultOrgID)
	banditHandler := handlers.NewBanditHandler(st, recommendationSvc, recConfig, tracer, cfg.Recommendation.DefaultOrgID, cfg.Recommendation.BanditAlgo, logger)
	rulesHandler := handlers.NewRulesHandler(st, rulesManager, cfg.Recommendation.DefaultOrgID, cfg.Recommendation.BrandTagPrefixes, cfg.Recommendation.CategoryTagPrefixes)
	manualSvc := manualsvc.New(st)
	manualHandler := handlers.NewManualOverridesHandler(manualSvc, rulesManager, cfg.Recommendation.DefaultOrgID)
	eventTypesHandler := handlers.NewEventTypesHandler(st, cfg.Recommendation.DefaultOrgID)
	explainHandler := handlers.NewExplainHandler(explainService, cfg.Recommendation.DefaultOrgID)
	auditHandler := handlers.NewAuditHandler(st, cfg.Recommendation.DefaultOrgID)

	obs, err := observability.Setup(ctx, cfg.Observability, logger)
	if err != nil {
		for _, closer := range closers {
			_ = closer(context.Background())
		}
		return nil, fmt.Errorf("observability: %w", err)
	}
	if obs != nil && obs.Shutdown != nil {
		closers = append(closers, obs.Shutdown)
	}

	var policyMetrics *policymetrics.Metrics
	if obs != nil {
		policyMetrics = obs.PolicyMetrics
	}

	recoHandler := handlers.NewRecommendationHandler(recommendationSvc, st, recConfig, tracer, cfg.Recommendation.DefaultOrgID, logger, policyMetrics)

	router := chi.NewRouter()
	if obs != nil {
		if obs.TraceMiddleware != nil {
			router.Use(obs.TraceMiddleware)
		}
		if obs.MetricsMiddleware != nil {
			router.Use(obs.MetricsMiddleware)
		}
	}
	router.Use(httpmiddleware.CORS(httpmiddleware.CORSOptions{
		AllowedOrigins:   cfg.HTTP.CORS.AllowedOrigins,
		AllowCredentials: cfg.HTTP.CORS.AllowCredentials,
	}))
	router.Use(middleware.RequestID, middleware.RealIP)
	router.Use(httpmiddleware.RequestLogger(logger))
	router.Use(httpmiddleware.JSONRecovererWithLogger(logger))
	router.Use(httpmiddleware.ErrorLogger(logger))

	errorMetrics := httpmiddleware.NewErrorMetrics()
	router.Use(httpmiddleware.ErrorMetricsMiddleware(errorMetrics))

	router.Get(endpoints.Health, func(w http.ResponseWriter, r *http.Request) {
		type workerHealth struct {
			Status  string `json:"status"`
			Message string `json:"message,omitempty"`
		}
		health := struct {
			Status  string                  `json:"status"`
			Workers map[string]workerHealth `json:"workers,omitempty"`
		}{
			Status: "ok",
		}

		if err := decisionRecorder.Healthy(); err != nil {
			health.Status = "degraded"
			if health.Workers == nil {
				health.Workers = make(map[string]workerHealth, 1)
			}
			health.Workers["decision_recorder"] = workerHealth{
				Status:  "error",
				Message: err.Error(),
			}
		}

		code := http.StatusOK
		if health.Status != "ok" {
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(health)
	})
	if obs != nil && cfg.Observability.MetricsEnabled && obs.MetricsHandler != nil && cfg.Observability.MetricsPath != "" {
		router.Handle(cfg.Observability.MetricsPath, obs.MetricsHandler)
	}

	protected := chi.NewRouter()
	protected.Use(httpmiddleware.RequireOrgID())

	if cfg.Auth.Enabled {
		keyAccess := make(map[string]httpmiddleware.APIKeyAccess, len(cfg.Auth.APIKeys))
		for key, accessCfg := range cfg.Auth.APIKeys {
			entry := httpmiddleware.APIKeyAccess{AllowAll: accessCfg.AllowAll}
			if !accessCfg.AllowAll && len(accessCfg.OrgIDs) > 0 {
				entry.OrgIDs = make(map[uuid.UUID]struct{}, len(accessCfg.OrgIDs))
				for _, id := range accessCfg.OrgIDs {
					entry.OrgIDs[id] = struct{}{}
				}
			}
			keyAccess[key] = entry
		}
		authorizer := httpmiddleware.NewAPIKeyAuthorizer(keyAccess, logger)
		protected.Use(authorizer.Middleware)
	}

	if cfg.Auth.RateLimit.Enabled {
		if limiter := httpmiddleware.NewRateLimiter(cfg.Auth.RateLimit.RequestsPerMinute, cfg.Auth.RateLimit.Burst, logger); limiter != nil {
			protected.Use(limiter.Middleware)
		}
	}
	protected.Options("/*", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	protected.Post(endpoints.ItemsUpsert, ingHandler.ItemsUpsert)
	protected.Post(endpoints.UsersUpsert, ingHandler.UsersUpsert)
	protected.Post(endpoints.EventsBatch, ingHandler.EventsBatch)
	protected.Post(endpoints.Recommendations, recoHandler.Recommend)
	protected.Get(endpoints.ItemsSimilar, recoHandler.ItemSimilar)
	protected.Post(endpoints.EventTypesUpsert, eventTypesHandler.EventTypesUpsert)
	protected.Get(endpoints.EventTypes, eventTypesHandler.EventTypesList)

	protected.Get(endpoints.UsersList, dataHandler.ListUsers)
	protected.Get(endpoints.ItemsList, dataHandler.ListItems)
	protected.Get(endpoints.EventsList, dataHandler.ListEvents)
	protected.Post(endpoints.UsersDelete, dataHandler.DeleteUsers)
	protected.Post(endpoints.ItemsDelete, dataHandler.DeleteItems)
	protected.Post(endpoints.EventsDelete, dataHandler.DeleteEvents)

	protected.Get(endpoints.SegmentProfiles, segmentsHandler.SegmentProfilesList)
	protected.Post(endpoints.SegmentProfilesUpsert, segmentsHandler.SegmentProfilesUpsert)
	protected.Post(endpoints.SegmentProfilesDelete, segmentsHandler.SegmentProfilesDelete)
	protected.Get(endpoints.Segments, segmentsHandler.SegmentsList)
	protected.Post(endpoints.SegmentsUpsert, segmentsHandler.SegmentsUpsert)
	protected.Post(endpoints.SegmentsDelete, segmentsHandler.SegmentsDelete)
	protected.Post(endpoints.SegmentDryRun, segmentsHandler.SegmentsDryRun)

	protected.Get(endpoints.AuditDecisions, auditHandler.AuditDecisionsList)
	protected.Get(endpoints.AuditDecisionByID, auditHandler.AuditDecisionGet)
	protected.Post(endpoints.AuditSearch, auditHandler.AuditDecisionsSearch)

	protected.Post(endpoints.BanditPoliciesUpsert, banditHandler.BanditPoliciesUpsert)
	protected.Get(endpoints.BanditPolicies, banditHandler.BanditPoliciesList)
	protected.Post(endpoints.BanditDecide, banditHandler.BanditDecide)
	protected.Post(endpoints.BanditReward, banditHandler.BanditReward)
	protected.Post(endpoints.BanditRecommendations, banditHandler.RecommendWithBandit)

	protected.Post(endpoints.Rules, rulesHandler.RulesCreate)
	protected.Put(endpoints.RuleByID, rulesHandler.RulesUpdate)
	protected.Get(endpoints.Rules, rulesHandler.RulesList)
	protected.Post(endpoints.RulesDryRun, rulesHandler.RulesDryRun)
	protected.Post(endpoints.ManualOverrides, manualHandler.ManualOverrideCreate)
	protected.Get(endpoints.ManualOverrides, manualHandler.ManualOverrideList)
	protected.Post(endpoints.ManualOverrideCancel, manualHandler.ManualOverrideCancel)
	protected.Get(endpoints.RecommendationPresets, recoHandler.RecommendationPresets)

	protected.Post(endpoints.ExplainLLM, explainHandler.ExplainLLM)

	router.Mount("/", protected)

	router.Options("/*", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	server := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}

	return &App{
		cfg:              cfg,
		logger:           logger,
		pool:             pool,
		decisionRecorder: decisionRecorder,
		server:           server,
		errorMetrics:     errorMetrics,
		closers:          closers,
	}, nil
}

// Run starts the HTTP server and blocks until the context is done or the server exits.
func (a *App) Run(ctx context.Context) error {
	a.startMetrics()

	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("api listening", zap.String("addr", a.server.Addr))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		close(errCh)
	}()

	var runErr error
	select {
	case <-ctx.Done():
		runErr = ctx.Err()
	case err := <-errCh:
		runErr = err
	}

	shutdownErr := a.Shutdown(context.Background())
	if runErr != nil && !errors.Is(runErr, context.Canceled) && !errors.Is(runErr, context.DeadlineExceeded) {
		return runErr
	}
	if shutdownErr != nil {
		return shutdownErr
	}
	return nil
}

// Shutdown gracefully stops the HTTP server and releases resources.
func (a *App) Shutdown(ctx context.Context) error {
	a.stopMetrics()

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	for _, closer := range a.closers {
		if err := closer(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Logger exposes the application's logger instance.
func (a *App) Logger() *zap.Logger {
	return a.logger
}

func (a *App) startMetrics() {
	if a.errorMetrics == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.metricsCancel = cancel
	a.metricsWG.Add(1)
	go func() {
		defer a.metricsWG.Done()
		httpmiddleware.LogErrorMetrics(ctx, a.logger, a.errorMetrics, 5*time.Minute)
	}()
}

func (a *App) stopMetrics() {
	if a.metricsCancel != nil {
		a.metricsCancel()
		a.metricsWG.Wait()
		a.metricsCancel = nil
	}
}

func buildLogger(debugCfg common.DebugConfig) (*zap.Logger, error) {
	if debugCfg.IsDebug() {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

func runMigrations(ctx context.Context, cfg config.MigrationConfig, dsn string, logger *zap.Logger) error {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open sql db: %w", err)
	}
	defer sqlDB.Close()

	runner := migrator.New(sqlDB, nil, migrator.Options{
		MigrationsDir: cfg.Dir,
		Logger: func(format string, args ...any) {
			logger.Sugar().Infof(format, args...)
		},
	})
	if err := runner.Up(ctx); err != nil {
		return fmt.Errorf("migrate on start: %w", err)
	}
	return nil
}

func newExplainService(cfg config.ExplainConfig, st *store.Store, logger *zap.Logger) *explain.Service {
	svc := &explain.Service{
		Collector: &explain.Collector{Store: st},
		Client:    explain.NullClient{},
		Config: explain.Config{
			Enabled:       cfg.Enabled,
			Provider:      cfg.Provider,
			ModelPrimary:  cfg.ModelPrimary,
			ModelEscalate: cfg.ModelEscalate,
			Timeout:       cfg.Timeout,
			MaxTokens:     cfg.MaxTokens,
			CacheTTL:      30 * time.Minute,
			CircuitBreaker: explain.CircuitBreakerConfig{
				Enabled:           cfg.CircuitBreaker.Enabled,
				FailureThreshold:  cfg.CircuitBreaker.FailureThreshold,
				ResetAfter:        cfg.CircuitBreaker.ResetAfter,
				HalfOpenSuccesses: cfg.CircuitBreaker.HalfOpenSuccesses,
			},
		},
		Logger: logger,
	}

	if !cfg.Enabled {
		svc.Client = explain.WithCircuitBreaker(svc.Client, svc.Config.CircuitBreaker, logger)
		return svc
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 6 * time.Second
	}
	httpClient := &http.Client{Timeout: timeout}

	registry := explain.DefaultProviderRegistry()
	client, err := registry.Build(cfg.Provider, explain.ProviderOptions{
		APIKey:     cfg.APIKey,
		BaseURL:    cfg.BaseURL,
		HTTPClient: httpClient,
		Logger:     logger,
	})
	if err != nil {
		logger.Warn("explain provider init failed", zap.Error(err))
		svc.Config.Enabled = false
	} else {
		svc.Client = client
	}

	svc.Client = explain.WithCircuitBreaker(svc.Client, svc.Config.CircuitBreaker, logger)
	return svc
}
