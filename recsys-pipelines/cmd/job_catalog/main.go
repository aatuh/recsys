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
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/signalstore/postgres"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/usecase"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/catalog"
)

func main() {
	var cfgPath, tenant, namespace, inputPath, format string
	flag.StringVar(&cfgPath, "config", "configs/env/local.json", "env config (json)")
	flag.StringVar(&tenant, "tenant", "", "tenant")
	flag.StringVar(&namespace, "namespace", "default", "namespace/surface")
	flag.StringVar(&inputPath, "input", "", "catalog file (csv or jsonl)")
	flag.StringVar(&format, "format", "", "format: csv | jsonl (optional)")
	flag.Parse()

	if tenant == "" || inputPath == "" {
		fmt.Fprintln(os.Stderr, "tenant and input are required")
		os.Exit(2)
	}

	env, err := config.LoadEnvConfig(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(2)
	}
	if strings.TrimSpace(env.DB.DSN) == "" {
		fmt.Fprintln(os.Stderr, "db.dsn is required for job_catalog")
		os.Exit(2)
	}

	reader, err := buildReader(format, inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reader error:", err)
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

	uc := usecase.NewImportItemTags(rt, reader, store, env.Limits.MaxItemsPerArtifact)
	if err := uc.Execute(context.Background(), tenant, namespace); err != nil {
		fmt.Fprintln(os.Stderr, "import failed:", err)
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
