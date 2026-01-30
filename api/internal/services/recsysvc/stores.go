package recsysvc

import (
	"context"
	"errors"
)

var (
	// ErrConfigNotFound indicates missing tenant config.
	ErrConfigNotFound = errors.New("tenant config not found")
	// ErrRulesNotFound indicates missing tenant rules.
	ErrRulesNotFound = errors.New("tenant rules not found")
	// ErrOverloaded indicates backpressure rejection.
	ErrOverloaded = errors.New("recsys service overloaded")
)

// ConfigStore retrieves tenant configuration.
type ConfigStore interface {
	GetConfig(ctx context.Context, tenantID, surface string) (TenantConfig, error)
}

// RulesStore retrieves tenant rules.
type RulesStore interface {
	GetRules(ctx context.Context, tenantID, surface string) (TenantRules, error)
}

// NoopConfigStore returns missing configs.
type NoopConfigStore struct{}

func (NoopConfigStore) GetConfig(ctx context.Context, tenantID, surface string) (TenantConfig, error) {
	return TenantConfig{}, ErrConfigNotFound
}

// NoopRulesStore returns missing rules.
type NoopRulesStore struct{}

func (NoopRulesStore) GetRules(ctx context.Context, tenantID, surface string) (TenantRules, error) {
	return TenantRules{}, ErrRulesNotFound
}
