package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/clock/systemclock"
	canon "github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/datasource/files"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/logger/stdlogger"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/metrics/noop"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/validator/builtin"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/usecase"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

func main() {
	var cfgPath, tenant, surface, startStr, endStr string
	flag.StringVar(&cfgPath, "config", "configs/env/local.json", "env config (json)")
	flag.StringVar(&tenant, "tenant", "", "tenant")
	flag.StringVar(&surface, "surface", "", "surface")
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

	canonical := canon.NewFSCanonicalStore(env.CanonicalDir)
	validator := builtin.New(canonical, builtin.Options{
		MinEvents:           0,
		MaxEvents:           env.Limits.MaxEventsPerRun,
		MaxDistinctItems:    env.Limits.MaxDistinctItemsPerRun,
		MaxDistinctSessions: env.Limits.MaxSessionsPerRun,
	})
	validateUC := usecase.NewValidateQuality(rt, validator)
	bf := usecase.NewBackfill(env.Limits.MaxDaysBackfill)

	ctx := context.Background()
	err = bf.Execute(ctx, startDay, endDay, func(ctx context.Context, w windows.Window) error {
		return validateUC.Execute(ctx, tenant, surface, w)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "validate failed:", err)
		os.Exit(1)
	}
}
