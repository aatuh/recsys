package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"recsys/internal/catalog"
	"recsys/internal/config"
	dbconfig "recsys/internal/http/db"
	"recsys/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type mode string

const (
	modeBackfill mode = "backfill"
	modeRefresh  mode = "refresh"
)

func main() {
	var (
		ns        = flag.String("namespace", "default", "namespace to process")
		rawMode   = flag.String("mode", string(modeBackfill), "mode: backfill or refresh")
		batchSize = flag.Int("batch", 200, "max items per batch")
		sinceRaw  = flag.String("since", "", "filter items updated since duration (e.g. 24h) or RFC3339 timestamp (refresh mode)")
		dryRun    = flag.Bool("dry-run", false, "log actions without persisting changes")
	)
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.Load(ctx, nil)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	orgID := cfg.Recommendation.DefaultOrgID
	if orgID == uuid.Nil {
		log.Fatal("ORG_ID must be configured")
	}

	modeValue := mode(strings.ToLower(strings.TrimSpace(*rawMode)))
	if modeValue != modeBackfill && modeValue != modeRefresh {
		log.Fatalf("invalid mode %q: expected backfill or refresh", *rawMode)
	}

	var (
		since *time.Time
	)
	if *sinceRaw != "" {
		if ts, err := time.Parse(time.RFC3339, *sinceRaw); err == nil {
			since = &ts
		} else if dur, err := time.ParseDuration(*sinceRaw); err == nil {
			ts := time.Now().UTC().Add(-dur)
			since = &ts
		} else {
			log.Fatalf("failed to parse --since=%q as duration or RFC3339 timestamp", *sinceRaw)
		}
	}

	pool, err := buildPool(ctx, cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	st := store.NewWithOptions(pool, store.Options{
		QueryTimeout:        cfg.Database.QueryTimeout,
		RetryAttempts:       cfg.Database.RetryAttempts,
		RetryInitialBackoff: cfg.Database.RetryInitialBackoff,
		RetryMaxBackoff:     cfg.Database.RetryMaxBackoff,
	})

	stats, err := run(ctx, st, runOptions{
		OrgID:      orgID,
		Namespace:  *ns,
		Mode:       modeValue,
		BatchSize:  *batchSize,
		Since:      since,
		DryRun:     *dryRun,
		LogVerbose: true,
	})
	if err != nil {
		log.Fatalf("catalog %s failed: %v", modeValue, err)
	}

	result, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Printf("%s complete:\n%s\n", strings.ToUpper(string(modeValue)), result)
}

type runOptions struct {
	OrgID      uuid.UUID
	Namespace  string
	Mode       mode
	BatchSize  int
	Since      *time.Time
	DryRun     bool
	LogVerbose bool
}

type runStats struct {
	Scanned int `json:"scanned"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
	Batches int `json:"batches"`
}

func run(ctx context.Context, st *store.Store, opts runOptions) (runStats, error) {
	isBackfill := opts.Mode == modeBackfill
	batchSize := opts.BatchSize
	if batchSize <= 0 {
		batchSize = 200
	}

	stats := runStats{}
	var (
		cursorAt  *time.Time
		cursorID  string
		maxLoops  = 10_000
		loopCount = 0
	)
	var cursorVal time.Time

	for {
		loopCount++
		if loopCount > maxLoops {
			return stats, errors.New("safety guard triggered: too many batches without completion")
		}

		rows, err := st.CatalogItems(ctx, opts.OrgID, opts.Namespace, store.CatalogQueryOptions{
			MissingOnly:     isBackfill,
			UpdatedSince:    opts.Since,
			CursorUpdatedAt: cursorAt,
			CursorItemID:    cursorID,
			Limit:           batchSize,
		})
		if err != nil {
			return stats, err
		}
		if len(rows) == 0 {
			break
		}

		stats.Batches++
		stats.Scanned += len(rows)

		var upserts []store.ItemUpsert
		for _, row := range rows {
			res, err := catalog.BuildUpsert(row, catalog.Options{
				GenerateEmbedding: true,
			})
			if err != nil {
				return stats, err
			}
			if !res.Changed {
				stats.Skipped++
				continue
			}
			if opts.DryRun && opts.LogVerbose {
				reportDryRun(row.ItemID, res.Upsert)
			}
			upserts = append(upserts, res.Upsert)
		}

		if len(upserts) > 0 && !opts.DryRun {
			if err := st.UpsertItems(ctx, opts.OrgID, opts.Namespace, upserts); err != nil {
				return stats, err
			}
			stats.Updated += len(upserts)
		} else if opts.DryRun {
			stats.Updated += len(upserts)
		}

		last := rows[len(rows)-1]
		cursorVal = last.UpdatedAt
		cursorAt = &cursorVal
		cursorID = last.ItemID

		if len(rows) < batchSize {
			break
		}
	}

	return stats, nil
}

func reportDryRun(itemID string, upsert store.ItemUpsert) {
	payload := map[string]any{"item_id": itemID}
	if upsert.Brand != nil {
		payload["brand"] = *upsert.Brand
	}
	if upsert.Category != nil {
		payload["category"] = *upsert.Category
	}
	if upsert.CategoryPath != nil {
		payload["category_path"] = *upsert.CategoryPath
	}
	if upsert.Description != nil {
		payload["description"] = *upsert.Description
	}
	if upsert.ImageURL != nil {
		payload["image_url"] = *upsert.ImageURL
	}
	if upsert.MetadataVersion != nil {
		payload["metadata_version"] = *upsert.MetadataVersion
	}
	data, _ := json.Marshal(payload)
	fmt.Fprintf(os.Stdout, "dry-run update %s: %s\n", itemID, data)
}

func buildPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	poolCfg := dbconfig.Config{
		MaxConnIdle:       cfg.Database.MaxConnIdle,
		MaxConnLifetime:   cfg.Database.MaxConnLifetime,
		HealthCheckPeriod: cfg.Database.HealthCheckPeriod,
		AcquireTimeout:    cfg.Database.AcquireTimeout,
		MinConns:          cfg.Database.MinConns,
		MaxConns:          cfg.Database.MaxConns,
	}
	return dbconfig.NewPool(ctx, cfg.Database.URL, poolCfg)
}
