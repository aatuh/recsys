package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/github.com/jackc/pgx/otelpgx"
	"go.opentelemetry.io/otel"
)

// Config controls pgxpool settings.
type Config struct {
	MaxConnIdle       time.Duration
	MaxConnLifetime   time.Duration
	HealthCheckPeriod time.Duration
	AcquireTimeout    time.Duration
	MinConns          int32
	MaxConns          int32
}

// DefaultConfig returns recommended defaults.
func DefaultConfig() Config {
	return Config{
		MaxConnIdle:       90 * time.Second,
		MaxConnLifetime:   0,
		HealthCheckPeriod: 30 * time.Second,
		AcquireTimeout:    5 * time.Second,
		MinConns:          0,
		MaxConns:          10,
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
	if cfg.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.HealthCheckPeriod > 0 {
		poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	}
	if poolCfg.ConnConfig.RuntimeParams == nil {
		poolCfg.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolCfg.ConnConfig.RuntimeParams["application_name"] = "recsys-api"
	poolCfg.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithTracerProvider(otel.GetTracerProvider()))
	if cfg.AcquireTimeout > 0 {
		poolCfg.ConnConfig.ConnectTimeout = cfg.AcquireTimeout
	}

	poolCtx := ctx
	var cancel context.CancelFunc
	if cfg.AcquireTimeout > 0 {
		poolCtx, cancel = context.WithTimeout(ctx, cfg.AcquireTimeout)
	}
	pool, err := pgxpool.NewWithConfig(poolCtx, poolCfg)
	if cancel != nil {
		cancel()
	}
	if err != nil {
		return nil, err
	}
	pingCtx := ctx
	var pingCancel context.CancelFunc
	if cfg.AcquireTimeout > 0 {
		pingCtx, pingCancel = context.WithTimeout(ctx, cfg.AcquireTimeout)
	}
	if pingCancel != nil {
		defer pingCancel()
	}
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
