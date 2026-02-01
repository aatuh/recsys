package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	areg "github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/artifactregistry/fs"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/clock/systemclock"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/logger/stdlogger"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/metrics/noop"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/validator/builtin"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/factory"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/staging"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/usecase"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
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
	if env.ArtifactsDir == "" {
		fmt.Fprintln(os.Stderr, "artifacts_dir must be configured for job_publish")
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

	canonical := factory.BuildCanonicalStore(env)
	validator := builtin.New(canonical, builtin.Options{
		MaxPopularityItems: env.Limits.MaxItemsPerArtifact,
		MaxCoocRows:        env.Limits.MaxItemsPerArtifact,
		MaxCoocNeighbors:   env.Limits.MaxNeighborsPerItem,
	})

	store, err := factory.BuildObjectStore(env)
	if err != nil {
		fmt.Fprintln(os.Stderr, "object store error:", err)
		os.Exit(2)
	}
	registry := areg.New(env.RegistryDir)
	publisher := usecase.NewPublishArtifacts(rt, store, registry, validator)
	stage := staging.New(env.ArtifactsDir)
	bf := usecase.NewBackfill(env.Limits.MaxDaysBackfill)

	ctx := context.Background()
	err = bf.Execute(ctx, startDay, endDay, func(ctx context.Context, w windows.Window) error {
		popKey := artifacts.Key{
			Tenant:  tenant,
			Surface: surface,
			Segment: segment,
			Type:    artifacts.TypePopularity,
		}
		coocKey := artifacts.Key{
			Tenant:  tenant,
			Surface: surface,
			Segment: segment,
			Type:    artifacts.TypeCooc,
		}

		popRef, popJSON, popOK, err := stage.LoadCurrent(ctx, popKey, w)
		if err != nil {
			return err
		}
		coocRef, coocJSON, coocOK, err := stage.LoadCurrent(ctx, coocKey, w)
		if err != nil {
			return err
		}
		if !popOK && !coocOK {
			rt.Logger.Info(ctx, "publish: no staged artifacts for window")
			return nil
		}

		in := usecase.PublishInput{Tenant: tenant, Surface: surface}
		if popOK {
			in.Popularity = &usecase.ArtifactBlob{Ref: popRef, JSON: popJSON}
		}
		if coocOK {
			in.Cooc = &usecase.ArtifactBlob{Ref: coocRef, JSON: coocJSON}
		}
		return publisher.Execute(ctx, in)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "publish failed:", err)
		os.Exit(1)
	}
}
