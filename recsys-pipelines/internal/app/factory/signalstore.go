package factory

import (
	"context"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/signalstore/postgres"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/signalstore"
)

// BuildSignalStore constructs a signal store if DB config is provided.
func BuildSignalStore(cfg config.EnvConfig) (signalstore.Store, func(), error) {
	if cfg.DB.DSN == "" {
		return nil, nil, nil
	}
	store, err := postgres.NewFromDSN(context.Background(), cfg.DB.DSN,
		postgres.WithCreateTenant(cfg.DB.AutoCreateTenant),
		postgres.WithStatementTimeout(time.Duration(cfg.DB.StatementTimeoutS)*time.Second),
	)
	if err != nil {
		return nil, nil, err
	}
	return store, store.Close, nil
}
