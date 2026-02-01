package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/clock/systemclock"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/logger/stdlogger"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/metrics/noop"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/signalstore/postgres"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/factory"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/usecase"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

func main() {
	var cfgPath, tenant, surface, segment, startStr, endStr string
	flag.StringVar(&cfgPath, "config", "configs/env/local.json", "env config (json)")
	flag.StringVar(&tenant, "tenant", "", "tenant")
	flag.StringVar(&surface, "surface", "", "surface")
	flag.StringVar(&segment, "segment", "", "segment (optional)")
	flag.StringVar(&startStr, "start", "", "start day YYYY-MM-DD")
	flag.StringVar(&endStr, "end", "", "end day YYYY-MM-DD")
	flag.Parse()

	if tenant == "" || surface == "" || startStr == "" || endStr == "" {
		fmt.Fprintln(os.Stderr, "missing required flags")
		os.Exit(2)
	}

	env, err := config.LoadEnvConfig(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(2)
	}
	if env.DB.DSN == "" {
		fmt.Fprintln(os.Stderr, "db.dsn must be configured")
		os.Exit(2)
	}

	startDay, err := time.ParseInLocation("2006-01-02", startStr, time.UTC)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid start:", err)
		os.Exit(2)
	}
	endDay, err := time.ParseInLocation("2006-01-02", endStr, time.UTC)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid end:", err)
		os.Exit(2)
	}

	rt := runtime.Runtime{
		Clock:   systemclock.SystemClock{},
		Logger:  stdlogger.New(stdlogger.WithLevelInfo()),
		Metrics: noop.NoopMetrics{},
	}

	store, err := postgres.NewFromDSN(context.Background(), env.DB.DSN,
		postgres.WithCreateTenant(env.DB.AutoCreateTenant),
		postgres.WithStatementTimeout(time.Duration(env.DB.StatementTimeoutS)*time.Second),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "db error:", err)
		os.Exit(2)
	}
	defer store.Close()

	canonical := factory.BuildCanonicalStore(env)
	pop := usecase.NewComputePopularity(rt, canonical, env.Limits.MaxItemsPerArtifact, env.Limits.MaxDistinctItemsPerRun)
	cooc := usecase.NewComputeCooc(
		rt,
		canonical,
		env.Limits.MaxNeighborsPerItem,
		int64(env.Limits.MinCoocSupport),
		env.Limits.MaxItemsPerArtifact,
		env.Limits.MaxSessionsPerRun,
		env.Limits.MaxItemsPerSession,
		env.Limits.MaxDistinctItemsPerRun,
	)
	persist := usecase.NewPersistSignals(rt, store)
	bf := usecase.NewBackfill(env.Limits.MaxDaysBackfill)

	ctx := context.Background()
	err = bf.Execute(ctx, startDay, endDay, func(ctx context.Context, w windows.Window) error {
		_, popJSON, err := pop.Execute(ctx, tenant, surface, segment, w)
		if err != nil {
			return err
		}
		_, coocJSON, err := cooc.Execute(ctx, tenant, surface, segment, w)
		if err != nil {
			return err
		}
		return persist.Execute(ctx, popJSON, coocJSON)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "db signals failed:", err)
		os.Exit(1)
	}
}
