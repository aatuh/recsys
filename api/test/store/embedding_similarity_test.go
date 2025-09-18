package store_test

import (
	"context"
	"testing"
	"time"

	"recsys/internal/store"
	"recsys/shared/util"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// getTestPool connects to the DB used by other integration tests.
func getTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := util.MustGetEnv("DATABASE_URL")
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

// ensurePgVector makes sure the extension exists for the current DB.
// This is safe to run under non-superuser if the role owns the DB.
func ensurePgVector(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;")
	if err != nil {
		t.Fatalf("create extension vector: %v", err)
	}
}

// clearNamespace deletes test data for isolation.
func clearNamespace(t *testing.T, pool *pgxpool.Pool, org uuid.UUID,
	ns string,
) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`DELETE FROM items WHERE org_id=$1 AND namespace=$2`, org, ns)
	if err != nil {
		t.Fatalf("clear items: %v", err)
	}
}

// Test that SimilarByEmbeddingTopK returns nearest neighbors by cosine
// distance, ignores items without embeddings, and filters unavailable.
func TestSimilarByEmbeddingTopK_Basics(t *testing.T) {
	pool := getTestPool(t)
	ensurePgVector(t, pool)
	s := store.New(pool)

	org := uuid.New()
	ns := "embedtest_" + time.Now().UTC().Format("150405")

	clearNamespace(t, pool, org, ns)

	// Embedding dimension must match store.EmbeddingDims (384).
	// For simplicity, use sparse vectors with only a few non-zeros.
	vUnit := func(i int) []float64 {
		v := make([]float64, store.EmbeddingDims)
		if i >= 0 && i < len(v) {
			v[i] = 1.0
		}
		return v
	}
	vMix := func(i, j int, wi, wj float64) []float64 {
		v := make([]float64, store.EmbeddingDims)
		if i >= 0 && i < len(v) {
			v[i] = wi
		}
		if j >= 0 && j < len(v) {
			v[j] = wj
		}
		return v
	}

	items := []store.ItemUpsert{
		// Anchor A: unit on dim 0
		{ItemID: "A", Available: true, Embedding: util.Ptr(vUnit(0))},
		// B: very similar to A (mostly on dim 0)
		{ItemID: "B", Available: true, Embedding: util.Ptr(vMix(0, 1, 0.98, 0.02))},
		// C: orthogonal-ish (on dim 1)
		{ItemID: "C", Available: true, Embedding: util.Ptr(vUnit(1))},
		// D: missing embedding, should be ignored
		{ItemID: "D", Available: true},
		// E: similar to A but unavailable, must be filtered out
		{ItemID: "E", Available: false, Embedding: util.Ptr(vMix(0, 2, 0.9, 0.1))},
	}

	if err := s.UpsertItems(context.Background(), org, ns, items); err != nil {
		t.Fatalf("UpsertItems: %v", err)
	}

	got, err := s.SimilarByEmbeddingTopK(context.Background(), org, ns, "A",
		10)
	if err != nil {
		t.Fatalf("SimilarByEmbeddingTopK: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected some neighbors, got 0")
	}

	// Expect B before C. D has no embedding, E is unavailable.
	wantOrder := []string{"B", "C"}
	if len(got) < len(wantOrder) {
		t.Fatalf("want at least %d results, got %d: %+v",
			len(wantOrder), len(got), got)
	}
	for i, id := range wantOrder {
		if got[i].ItemID != id {
			t.Fatalf("at %d want %q got %q", i, id, got[i].ItemID)
		}
	}
	// Scores are 1 - cosine_distance. Identical vectors give ~1.0.
	if got[0].Score <= got[1].Score {
		// Already ordered ASC by distance, so Score should descend.
		t.Fatalf("expected score[0] > score[1], got %.6f <= %.6f",
			got[0].Score, got[1].Score)
	}
}
