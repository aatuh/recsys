package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"recsys/internal/catalog"
	"recsys/internal/config"
	dbconfig "recsys/internal/http/db"
	"recsys/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type runStats struct {
	Namespace      string `json:"namespace"`
	ItemsProcessed int    `json:"items_processed"`
	ItemsUpserted  int    `json:"items_upserted"`
	UsersProcessed int    `json:"users_processed"`
	UsersUpserted  int    `json:"users_upserted"`
	DryRun         bool   `json:"dry_run"`
}

func main() {
	var (
		ns       = flag.String("namespace", "default", "namespace to process")
		sinceRaw = flag.String("since", "", "only consider events on/after this RFC3339 timestamp or duration (e.g. 30d)")
		dryRun   = flag.Bool("dry-run", false, "compute factors without writing them")
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

	var since *time.Time
	if raw := strings.TrimSpace(*sinceRaw); raw != "" {
		if ts, err := time.Parse(time.RFC3339, raw); err == nil {
			since = &ts
		} else if dur, err := time.ParseDuration(raw); err == nil {
			ts := time.Now().UTC().Add(-dur)
			since = &ts
		} else {
			log.Fatalf("failed to parse --since=%q as RFC3339 timestamp or duration", raw)
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

	stats, err := run(ctx, st, runConfig{
		Pool:      pool,
		OrgID:     orgID,
		Namespace: strings.TrimSpace(*ns),
		Since:     since,
		DryRun:    *dryRun,
	})
	if err != nil {
		log.Fatalf("collaborative factors failed: %v", err)
	}

	result, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Printf("collab factors complete:\n%s\n", result)
}

type runConfig struct {
	Pool      *pgxpool.Pool
	OrgID     uuid.UUID
	Namespace string
	Since     *time.Time
	DryRun    bool
}

func run(ctx context.Context, st *store.Store, cfg runConfig) (runStats, error) {
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	stats := runStats{Namespace: cfg.Namespace, DryRun: cfg.DryRun}

	itemVectors, itemUpserts, err := loadItemFactors(ctx, cfg.Pool, cfg.OrgID, cfg.Namespace)
	if err != nil {
		return stats, fmt.Errorf("load item factors: %w", err)
	}
	stats.ItemsProcessed = len(itemUpserts)

	if !cfg.DryRun && len(itemUpserts) > 0 {
		if err := st.UpsertItemFactors(ctx, cfg.OrgID, cfg.Namespace, itemUpserts); err != nil {
			return stats, fmt.Errorf("upsert item factors: %w", err)
		}
	}
	stats.ItemsUpserted = len(itemUpserts)

	userUpserts, err := buildUserFactors(ctx, cfg.Pool, cfg.OrgID, cfg.Namespace, cfg.Since, itemVectors)
	if err != nil {
		return stats, fmt.Errorf("build user factors: %w", err)
	}
	stats.UsersProcessed = len(userUpserts)

	if !cfg.DryRun && len(userUpserts) > 0 {
		if err := st.UpsertUserFactors(ctx, cfg.OrgID, cfg.Namespace, userUpserts); err != nil {
			return stats, fmt.Errorf("upsert user factors: %w", err)
		}
	}
	stats.UsersUpserted = len(userUpserts)

	return stats, nil
}

func loadItemFactors(ctx context.Context, pool *pgxpool.Pool, orgID uuid.UUID, namespace string) (map[string][]float64, []store.ItemFactorUpsert, error) {
	rows, err := pool.Query(ctx, `
        SELECT item_id,
               embedding::text,
               tags,
               brand,
               category,
               description
        FROM items
        WHERE org_id = $1
          AND namespace = $2
    `, orgID, namespace)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	itemVectors := make(map[string][]float64)
	upserts := make([]store.ItemFactorUpsert, 0, 1024)

	for rows.Next() {
		var (
			itemID      string
			embedding   pgtype.Text
			tags        []string
			brand       pgtype.Text
			category    pgtype.Text
			description pgtype.Text
		)
		if err := rows.Scan(&itemID, &embedding, &tags, &brand, &category, &description); err != nil {
			return nil, nil, err
		}

		var vec []float64
		if embedding.Valid && strings.TrimSpace(embedding.String) != "" {
			parsed, err := parseVectorLiteral(embedding.String)
			if err != nil {
				return nil, nil, fmt.Errorf("item %s: parse embedding: %w", itemID, err)
			}
			vec = parsed
		} else {
			textParts := []string{itemID}
			if brand.Valid {
				textParts = append(textParts, brand.String)
			}
			if category.Valid {
				textParts = append(textParts, category.String)
			}
			if len(tags) > 0 {
				textParts = append(textParts, strings.Join(tags, " "))
			}
			if description.Valid {
				textParts = append(textParts, description.String)
			}
			vec = catalog.DeterministicEmbeddingFromText(strings.Join(textParts, " "))
			if len(vec) == 0 {
				vec = catalog.DeterministicEmbeddingFromText(itemID)
			}
		}

		if len(vec) != store.EmbeddingDims {
			return nil, nil, fmt.Errorf("item %s: expected vector length %d, got %d", itemID, store.EmbeddingDims, len(vec))
		}

		copied := make([]float64, len(vec))
		copy(copied, vec)
		itemVectors[itemID] = copied
		upserts = append(upserts, store.ItemFactorUpsert{
			ItemID:  itemID,
			Factors: copied,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return itemVectors, upserts, nil
}

func buildUserFactors(ctx context.Context, pool *pgxpool.Pool, orgID uuid.UUID, namespace string, since *time.Time, itemVectors map[string][]float64) ([]store.UserFactorUpsert, error) {
	rows, err := pool.Query(ctx, `
        SELECT user_id,
               item_id
        FROM events
        WHERE org_id = $1
          AND namespace = $2
          AND item_id IS NOT NULL
          AND ($3::timestamptz IS NULL OR ts >= $3)
    `, orgID, namespace, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type accumulator struct {
		sum   []float64
		count int
	}

	acc := make(map[string]*accumulator)

	for rows.Next() {
		var (
			userID string
			itemID string
		)
		if err := rows.Scan(&userID, &itemID); err != nil {
			return nil, err
		}
		vec, ok := itemVectors[itemID]
		if !ok {
			continue
		}

		entry := acc[userID]
		if entry == nil {
			entry = &accumulator{
				sum:   make([]float64, len(vec)),
				count: 0,
			}
			acc[userID] = entry
		}
		for i, v := range vec {
			entry.sum[i] += v
		}
		entry.count++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	upserts := make([]store.UserFactorUpsert, 0, len(acc))
	for userID, entry := range acc {
		if entry.count == 0 {
			continue
		}
		factors := make([]float64, len(entry.sum))
		denom := float64(entry.count)
		for i, v := range entry.sum {
			factors[i] = v / denom
		}
		upserts = append(upserts, store.UserFactorUpsert{
			UserID:  userID,
			Factors: factors,
		})
	}
	return upserts, nil
}

func parseVectorLiteral(s string) ([]float64, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return nil, fmt.Errorf("empty vector literal")
	}
	if s[0] == '[' && s[len(s)-1] == ']' {
		s = s[1 : len(s)-1]
	}
	if strings.TrimSpace(s) == "" {
		return nil, fmt.Errorf("empty vector literal")
	}
	parts := strings.Split(s, ",")
	vec := make([]float64, 0, len(parts))
	for _, part := range parts {
		val := strings.TrimSpace(part)
		if val == "" {
			continue
		}
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, err
		}
		vec = append(vec, f)
	}
	return vec, nil
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
