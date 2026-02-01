package root

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	areg "github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/artifactregistry/fs"
	checkpointfs "github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/checkpoint/fs"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/clock/systemclock"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/logger/stdlogger"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/metrics/noop"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/validator/builtin"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/factory"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/usecase"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/workflow"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

func Main() int {
	if len(os.Args) < 2 {
		usage()
		return 2
	}
	switch os.Args[1] {
	case "run":
		return run(os.Args[2:])
	case "version":
		fmt.Println("dev")
		return 0
	case "-h", "--help", "help":
		usage()
		return 0
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", os.Args[1])
		usage()
		return 2
	}
}

func usage() {
	fmt.Print(`recsys-pipelines

Usage:
  recsys-pipelines run --config <path> --tenant <t> --surface <s> --start YYYY-MM-DD --end YYYY-MM-DD
  recsys-pipelines run --config <path> --tenant <t> --surface <s> --end YYYY-MM-DD --incremental

Notes:
  The scaffold uses filesystem adapters for local development.
	`)
}

func run(args []string) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var cfgPath, tenant, surface, segment, startStr, endStr string
	var incremental bool
	fs.StringVar(&cfgPath, "config", "configs/env/local.json", "env config (json)")
	fs.StringVar(&tenant, "tenant", "", "tenant")
	fs.StringVar(&surface, "surface", "", "surface")
	fs.StringVar(&segment, "segment", "", "segment (optional)")
	fs.StringVar(&startStr, "start", "", "start day YYYY-MM-DD")
	fs.StringVar(&endStr, "end", "", "end day YYYY-MM-DD")
	fs.BoolVar(&incremental, "incremental", false, "use checkpointed start day")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if tenant == "" || surface == "" || endStr == "" || (!incremental && startStr == "") {
		fmt.Fprintln(os.Stderr, "missing required flags")
		return 2
	}

	var (
		startDay time.Time
		err      error
	)
	if startStr != "" {
		startDay, err = time.ParseInLocation("2006-01-02", startStr, time.UTC)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid start:", err)
			return 2
		}
	}
	endDay, err := time.ParseInLocation("2006-01-02", endStr, time.UTC)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid end:", err)
		return 2
	}

	env, err := config.LoadEnvConfig(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		return 2
	}

	rt := runtime.Runtime{
		Clock:   systemclock.SystemClock{},
		Logger:  stdlogger.New(stdlogger.WithLevelInfo()),
		Metrics: noop.NoopMetrics{},
	}

	raw, rawClose, err := factory.BuildRawSource(env)
	if err != nil {
		fmt.Fprintln(os.Stderr, "raw source error:", err)
		return 2
	}
	if rawClose != nil {
		defer rawClose()
	}
	canonical := factory.BuildCanonicalStore(env)
	store, err := factory.BuildObjectStore(env)
	if err != nil {
		fmt.Fprintln(os.Stderr, "object store error:", err)
		return 2
	}
	registry := areg.New(env.RegistryDir)
	validator := builtin.New(canonical, builtin.Options{
		MinEvents:           0,
		MaxEvents:           env.Limits.MaxEventsPerRun,
		MaxDistinctItems:    env.Limits.MaxDistinctItemsPerRun,
		MaxDistinctSessions: env.Limits.MaxSessionsPerRun,
		MaxPopularityItems:  env.Limits.MaxItemsPerArtifact,
		MaxCoocRows:         env.Limits.MaxItemsPerArtifact,
		MaxCoocNeighbors:    env.Limits.MaxNeighborsPerItem,
	})
	signalStore, signalClose, err := factory.BuildSignalStore(env)
	if err != nil {
		fmt.Fprintln(os.Stderr, "signal store error:", err)
		return 2
	}
	if signalClose != nil {
		defer signalClose()
	}

	ingest := usecase.NewIngestEvents(rt, raw, canonical, env.Limits.MaxEventsPerRun)
	validateUC := usecase.NewValidateQuality(rt, validator)
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
	publish := usecase.NewPublishArtifacts(rt, store, registry, validator)
	var persistSignals *usecase.PersistSignals
	if signalStore != nil {
		persistSignals = usecase.NewPersistSignals(rt, signalStore)
	}

	pipe := &workflow.Pipeline{
		RT:           rt,
		ArtifactsDir: env.ArtifactsDir,
		Ingest:       ingest,
		Validate:     validateUC,
		Pop:          pop,
		Cooc:         cooc,
		Signals:      persistSignals,
		Publish:      publish,
	}

	bf := usecase.NewBackfill(env.Limits.MaxDaysBackfill)
	checkpoints := checkpointfs.New(env.CheckpointDir)
	if incremental {
		last, ok, err := checkpoints.GetLastIngested(context.Background(), tenant, surface)
		if err != nil {
			fmt.Fprintln(os.Stderr, "checkpoint error:", err)
			return 2
		}
		if !ok && startStr == "" {
			fmt.Fprintln(os.Stderr, "checkpoint missing; provide --start for first run")
			return 2
		}
		if ok && startStr == "" {
			startDay = last.Add(24 * time.Hour)
		}
	}

	ctx := context.Background()
	err = bf.Execute(ctx, startDay, endDay, func(ctx context.Context, w windows.Window) error {
		if err := pipe.RunDay(ctx, tenant, surface, segment, w); err != nil {
			return err
		}
		return checkpoints.SetLastIngested(ctx, tenant, surface, w.Start)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "run failed:", err)
		return 1
	}
	return 0
}
