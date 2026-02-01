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
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/staging"
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

	env, err := config.LoadEnvConfig(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	rt := runtime.Runtime{
		Clock:   systemclock.SystemClock{},
		Logger:  stdlogger.New(stdlogger.WithLevelInfo()),
		Metrics: noop.NoopMetrics{},
	}

	startDay, err := time.ParseInLocation("2006-01-02", startStr, time.UTC)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	endDay, err := time.ParseInLocation("2006-01-02", endStr, time.UTC)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	canonical := canon.NewFSCanonicalStore(env.CanonicalDir)
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
	stage := staging.New(env.ArtifactsDir)
	bf := usecase.NewBackfill(env.Limits.MaxDaysBackfill)

	ctx := context.Background()
	err = bf.Execute(ctx, startDay, endDay, func(ctx context.Context, w windows.Window) error {
		ref, blob, err := cooc.Execute(ctx, tenant, surface, segment, w)
		if err != nil {
			return err
		}
		if env.ArtifactsDir == "" {
			return nil
		}
		_, err = stage.Put(ctx, ref, blob)
		return err
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "cooc failed:", err)
		os.Exit(1)
	}
}
