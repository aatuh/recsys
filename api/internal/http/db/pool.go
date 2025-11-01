package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config controls pgxpool settings.
type Config struct {
	MaxConnIdle time.Duration
	MinConns    int32
	MaxConns    int32
}

// DefaultConfig returns recommended defaults.
func DefaultConfig() Config {
	return Config{
		MaxConnIdle: 90 * time.Second,
		MinConns:    0,
		MaxConns:    10,
	}
}

// NewPool initialises a pgx connection pool with the provided settings.
func NewPool(ctx context.Context, dsn string, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	if cfg.MaxConnIdle > 0 {
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdle
	}
	if cfg.MinConns >= 0 {
		poolCfg.MinConns = cfg.MinConns
	}
	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
