package shared

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// MustPool creates a new pool connection and returns it.
func MustPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(getDSN())
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)

	ctxPing, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	require.NoError(t, pool.Ping(ctxPing))

	t.Cleanup(func() { pool.Close() })
	return pool
}

// MustOrgID returns the org ID from the environment or a default value.
func MustOrgID(t *testing.T) uuid.UUID {
	t.Helper()
	idStr := os.Getenv("ORG_ID")
	if idStr == "" {
		idStr = "00000000-0000-0000-0000-000000000001"
	}
	id, err := uuid.Parse(idStr)
	require.NoError(t, err)
	return id
}

// CleanTables truncates mutable tables between tests.
func CleanTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `TRUNCATE TABLE events, items, users RESTART IDENTITY`)
	require.NoError(t, err)
}

// getDSN reads DATABASE_URL or falls back to the compose default.
func getDSN() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://recsys:recsys@db:5432/recsys?sslmode=disable"
}
