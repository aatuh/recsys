package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"recsys/internal/config"
	"recsys/internal/eval/blend"
	dbconfig "recsys/internal/http/db"
	"recsys/internal/http/handlers"
	"recsys/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

type candidateFile struct {
	Candidates []blend.CandidateConfig `yaml:"configs"`
}

func main() {
	var (
		namespace  = flag.String("namespace", "default", "namespace to evaluate")
		limit      = flag.Int("limit", 200, "max user samples")
		minEvents  = flag.Int("min-events", 5, "minimum events per user")
		k          = flag.Int("k", 20, "recommendation length")
		lookback   = flag.Duration("lookback", 30*24*time.Hour, "event lookback window")
		configPath = flag.String("configs", "", "path to YAML file with blend configs")
	)
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.Load(ctx, nil)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	orgID := cfg.Recommendation.DefaultOrgID
	if orgID == uuid.Nil {
		log.Fatal("recommendation default org id is required (ORG_ID)")
	}

	pool, err := buildPool(ctx, cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	storeOpts := store.Options{
		QueryTimeout:        cfg.Database.QueryTimeout,
		RetryAttempts:       cfg.Database.RetryAttempts,
		RetryInitialBackoff: cfg.Database.RetryInitialBackoff,
		RetryMaxBackoff:     cfg.Database.RetryMaxBackoff,
	}
	st := store.NewWithOptions(pool, storeOpts)

	recCfg := handlers.RecommendationConfig{
		HalfLifeDays:        cfg.Recommendation.HalfLifeDays,
		CoVisWindowDays:     cfg.Recommendation.CoVisWindowDays,
		PopularityFanout:    cfg.Recommendation.PopularityFanout,
		MMRLambda:           cfg.Recommendation.MMRLambda,
		BrandCap:            cfg.Recommendation.BrandCap,
		CategoryCap:         cfg.Recommendation.CategoryCap,
		RuleExcludeEvents:   cfg.Recommendation.RuleExcludeEvents,
		ExcludeEventTypes:   cfg.Recommendation.ExcludeEventTypes,
		BrandTagPrefixes:    cfg.Recommendation.BrandTagPrefixes,
		CategoryTagPrefixes: cfg.Recommendation.CategoryTagPrefixes,
		PurchasedWindowDays: cfg.Recommendation.PurchasedWindowDays,
		ProfileWindowDays:   cfg.Recommendation.Profile.WindowDays,
		ProfileBoost:        cfg.Recommendation.Profile.Boost,
		ProfileTopNTags:     cfg.Recommendation.Profile.TopNTags,
		BlendAlpha:          cfg.Recommendation.Blend.Alpha,
		BlendBeta:           cfg.Recommendation.Blend.Beta,
		BlendGamma:          cfg.Recommendation.Blend.Gamma,
		RulesEnabled:        cfg.Rules.Enabled,
	}
	baseCfg := recCfg.BaseConfig()

	candidates, err := loadCandidates(*configPath)
	if err != nil {
		log.Fatalf("load configs: %v", err)
	}

	h := blend.Harness{
		Store:      st,
		BaseConfig: baseCfg,
		OrgID:      orgID,
		Namespace:  *namespace,
		K:          *k,
		Limit:      *limit,
		MinEvents:  *minEvents,
		Lookback:   *lookback,
		Candidates: candidates,
	}

	results, err := h.Run(ctx)
	if err != nil {
		log.Fatalf("run evaluation: %v", err)
	}

	fmt.Printf("Blend evaluation for namespace=%s (samples=%d, k=%d)\n", *namespace, h.Limit, *k)
	fmt.Printf("%-20s %6s %6s %7s %7s %7s %8s %8s\n", "config", "hit@", "mrr", "avgRank", "coverage", "listsz", "hits", "fail")
	for _, res := range results {
		hit := fmtFloat(res.HitRate)
		mrr := fmtFloat(res.MRR)
		avgRank := fmtFloat(res.AvgRank)
		coverage := fmtFloat(res.Coverage)
		listSize := fmtFloat(res.AvgListLength)
		fmt.Printf("%-20s %6s %6s %7s %7s %7s %4d/%-3d %4d\n",
			res.Name, hit, mrr, avgRank, coverage, listSize, res.Hits, res.Total, res.Failures)
	}
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

func loadCandidates(path string) ([]blend.CandidateConfig, error) {
	if path == "" {
		return defaultCandidates(), nil
	}
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	var file candidateFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if len(file.Candidates) == 0 {
		return nil, errors.New("configs file contained no candidates")
	}
	return file.Candidates, nil
}

func defaultCandidates() []blend.CandidateConfig {
	return []blend.CandidateConfig{
		{Name: "baseline", Alpha: 0.3, Beta: 0.5, Gamma: 0.2},
		{Name: "pop-heavy", Alpha: 0.6, Beta: 0.3, Gamma: 0.1},
		{Name: "embed-heavy", Alpha: 0.2, Beta: 0.2, Gamma: 0.6},
	}
}

func fmtFloat(v float64) string {
	if v == 0 {
		return "0.000"
	}
	return fmt.Sprintf("%.3f", v)
}
