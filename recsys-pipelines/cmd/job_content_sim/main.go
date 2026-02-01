package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/catalog/csv"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/catalog/jsonl"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/clock/systemclock"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/logger/stdlogger"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/metrics/noop"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/staging"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/usecase"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/catalog"
)

func main() {
	var cfgPath, tenant, surface, segment, inputPath, format, startStr, endStr string
	flag.StringVar(&cfgPath, "config", "configs/env/local.json", "env config (json)")
	flag.StringVar(&tenant, "tenant", "", "tenant")
	flag.StringVar(&surface, "surface", "", "surface")
	flag.StringVar(&segment, "segment", "", "segment (optional)")
	flag.StringVar(&inputPath, "input", "", "catalog file (csv or jsonl)")
	flag.StringVar(&format, "format", "", "format: csv | jsonl (optional)")
	flag.StringVar(&startStr, "start", "", "start day YYYY-MM-DD")
	flag.StringVar(&endStr, "end", "", "end day YYYY-MM-DD")
	flag.Parse()

	if tenant == "" || surface == "" || inputPath == "" {
		fmt.Fprintln(os.Stderr, "tenant, surface, and input are required")
		os.Exit(2)
	}

	env, err := config.LoadEnvConfig(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(2)
	}

	reader, err := buildReader(format, inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reader error:", err)
		os.Exit(2)
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

	rt := runtime.Runtime{
		Clock:   systemclock.SystemClock{},
		Logger:  stdlogger.New(stdlogger.WithLevelInfo()),
		Metrics: noop.NoopMetrics{},
	}
	content := usecase.NewComputeContentSim(rt, reader, env.Limits.MaxItemsPerArtifact)
	stage := staging.New(env.ArtifactsDir)
	bf := usecase.NewBackfill(env.Limits.MaxDaysBackfill)

	ctx := context.Background()
	err = bf.Execute(ctx, startDay, endDay, func(ctx context.Context, w windows.Window) error {
		ref, blob, err := content.Execute(ctx, tenant, surface, segment, w)
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
		fmt.Fprintln(os.Stderr, "content_sim failed:", err)
		os.Exit(1)
	}
}

func buildReader(format, inputPath string) (catalog.Reader, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		ext := strings.ToLower(filepath.Ext(inputPath))
		switch ext {
		case ".csv":
			format = "csv"
		case ".jsonl":
			format = "jsonl"
		default:
			return nil, fmt.Errorf("unknown input format; use --format")
		}
	}
	if format == "csv" {
		return csv.New(inputPath), nil
	}
	if format == "jsonl" {
		return jsonl.New(inputPath), nil
	}
	return nil, fmt.Errorf("unsupported format: %s", format)
}
