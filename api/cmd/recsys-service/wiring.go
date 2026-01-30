package main

import (
	"recsys/internal/config"
	"recsys/internal/experiments"
	"recsys/internal/exposure"
	"recsys/internal/services/adminsvc"
	"recsys/internal/services/recsysvc"
	"recsys/internal/store"
	"time"

	"github.com/aatuh/api-toolkit-contrib/adapters/validation"
	"github.com/aatuh/api-toolkit/ports"
	"github.com/aatuh/recsys-algo/algorithm"
	"github.com/aatuh/recsys-algo/rules"
)

type appDeps struct {
	RecsysService       *recsysvc.Service
	AdminService        *adminsvc.Service
	Validator           ports.Validator
	OverloadRetryAfter  time.Duration
	ExposureLogger      exposure.Logger
	ExposureHasher      exposure.Hasher
	ExperimentAssigner  experiments.Assigner
	ExplainMaxItems     int
	ExplainRequireAdmin bool
	AdminRole           string
	Close               func()
}

func buildAppDeps(log ports.Logger, pool ports.DatabasePool, cfg config.Config) (appDeps, error) {
	_ = log
	_ = pool
	cacheCfg := cfg.Performance.Cache
	backpressure := cfg.Performance.Backpressure
	configStore := recsysvc.NewCachedConfigStore(store.NewTenantConfigStore(pool), cacheCfg.ConfigTTL)
	rulesStore := recsysvc.NewCachedRulesStore(store.NewTenantRulesStore(pool), cacheCfg.RulesTTL)
	var configCache adminsvc.ConfigCache
	if cache, ok := configStore.(*recsysvc.CachedConfigStore); ok {
		configCache = cache
	}
	var rulesCache adminsvc.RulesCache
	if cache, ok := rulesStore.(*recsysvc.CachedRulesStore); ok {
		rulesCache = cache
	}
	queue := recsysvc.NewBoundedQueue(backpressure.MaxInFlight, backpressure.MaxQueue, backpressure.WaitTimeout)

	algoCfg := cfg.Algo
	rulesManager := rules.NewManager(store.NewRulesManagerStore(pool), rules.ManagerOptions{
		Enabled:         algoCfg.RulesEnabled,
		RefreshInterval: algoCfg.RulesRefreshInterval,
		MaxPinSlots:     algoCfg.RulesMaxPins,
	})
	engine := recsysvc.NewAlgoEngine(
		recsysvc.AlgoEngineConfig{
			Version:          algoCfg.Version,
			DefaultNamespace: algoCfg.DefaultNamespace,
			AlgorithmConfig: algorithm.Config{
				BlendAlpha:                 algoCfg.BlendAlpha,
				BlendBeta:                  algoCfg.BlendBeta,
				BlendGamma:                 algoCfg.BlendGamma,
				ProfileBoost:               algoCfg.ProfileBoost,
				ProfileWindowDays:          algoCfg.ProfileWindowDays,
				ProfileTopNTags:            algoCfg.ProfileTopNTags,
				ProfileMinEventsForBoost:   algoCfg.ProfileMinEventsForBoost,
				ProfileColdStartMultiplier: algoCfg.ProfileColdStartMultiplier,
				ProfileStarterBlendWeight:  algoCfg.ProfileStarterBlendWeight,
				MMRLambda:                  algoCfg.MMRLambda,
				BrandCap:                   algoCfg.BrandCap,
				CategoryCap:                algoCfg.CategoryCap,
				HalfLifeDays:               algoCfg.HalfLifeDays,
				CoVisWindowDays:            algoCfg.CoVisWindowDays,
				PurchasedWindowDays:        algoCfg.PurchasedWindowDays,
				RuleExcludeEvents:          algoCfg.RuleExcludeEvents,
				ExcludeEventTypes:          algoCfg.ExcludeEventTypes,
				BrandTagPrefixes:           algoCfg.BrandTagPrefixes,
				CategoryTagPrefixes:        algoCfg.CategoryTagPrefixes,
				RulesEnabled:               algoCfg.RulesEnabled,
				PopularityFanout:           algoCfg.PopularityFanout,
				MaxK:                       algoCfg.MaxK,
				MaxFanout:                  algoCfg.MaxFanout,
				MaxExcludeIDs:              algoCfg.MaxExcludeIDs,
				MaxAnchorsInjected:         algoCfg.MaxAnchorsInjected,
				SessionLookbackEvents:      algoCfg.SessionLookbackEvents,
				SessionLookaheadMinutes:    algoCfg.SessionLookaheadMinutes,
			},
		},
		store.NewAlgoStore(pool),
		rulesManager,
	)

	adminSvc := adminsvc.New(
		store.NewAdminStore(pool),
		adminsvc.WithConfigCache(configCache),
		adminsvc.WithRulesCache(rulesCache),
		adminsvc.WithRulesManager(rulesManager, algoCfg.DefaultNamespace),
	)

	recSvc := recsysvc.NewWithOptions(
		engine,
		recsysvc.WithBackpressure(queue),
		recsysvc.WithConfigStore(configStore),
		recsysvc.WithRulesStore(rulesStore),
	)
	var exposureLogger exposure.Logger
	var closers []func()
	if cfg.Exposure.Enabled {
		logger, err := exposure.NewFileLogger(exposure.FileLoggerOptions{
			Path:          cfg.Exposure.Path,
			Format:        cfg.Exposure.Format,
			Fsync:         cfg.Exposure.Fsync,
			RetentionDays: cfg.Exposure.RetentionDays,
		})
		if err != nil {
			return appDeps{}, err
		}
		exposureLogger = logger
		closers = append(closers, func() {
			_ = logger.Close()
		})
	}
	var assigner experiments.Assigner
	if cfg.Experiment.Enabled {
		assigner = experiments.NewDeterministicAssigner(cfg.Experiment.DefaultVariants, cfg.Experiment.Salt)
	}
	closeFn := func() {
		for i := len(closers) - 1; i >= 0; i-- {
			closers[i]()
		}
	}

	return appDeps{
		RecsysService:       recSvc,
		AdminService:        adminSvc,
		Validator:           validation.NewBasicValidator(),
		OverloadRetryAfter:  backpressure.RetryAfter,
		ExposureLogger:      exposureLogger,
		ExposureHasher:      exposure.NewHasher(cfg.Exposure.HashSalt),
		ExperimentAssigner:  assigner,
		ExplainMaxItems:     cfg.Explain.MaxItems,
		ExplainRequireAdmin: cfg.Explain.RequireAdmin,
		AdminRole:           cfg.Auth.AdminRole,
		Close:               closeFn,
	}, nil
}
