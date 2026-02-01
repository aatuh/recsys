package store

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aatuh/recsys-suite/api/internal/services/recsysvc"

	"github.com/aatuh/api-toolkit-contrib/adapters/txpostgres"
	"github.com/aatuh/api-toolkit/ports"
)

// TenantConfigStore is a Postgres-backed adapter for tenant configs.
type TenantConfigStore struct {
	Pool ports.DatabasePool
}

func NewTenantConfigStore(pool ports.DatabasePool) *TenantConfigStore {
	return &TenantConfigStore{Pool: pool}
}

// GetConfig returns the current config and version for the tenant.
func (s *TenantConfigStore) GetConfig(ctx context.Context, tenantID, surface string) (recsysvc.TenantConfig, error) {
	if strings.TrimSpace(tenantID) == "" {
		return recsysvc.TenantConfig{}, recsysvc.ErrConfigNotFound
	}
	db := txpostgres.FromCtx(ctx, s.Pool)
	const q = `
select v.config, v.etag
  from tenants t
  join tenant_configs_current c on c.tenant_id = t.id
  join tenant_config_versions v on v.id = c.config_version_id
 where t.external_id = $1 or t.id::text = $1
`
	var raw []byte
	var etag string
	if err := db.QueryRow(ctx, q, tenantID).Scan(&raw, &etag); err != nil {
		if txpostgres.IsNoRows(err) {
			return recsysvc.TenantConfig{}, recsysvc.ErrConfigNotFound
		}
		return recsysvc.TenantConfig{}, err
	}

	var payload struct {
		Weights *recsysvc.Weights `json:"weights"`
		Flags   map[string]bool   `json:"flags"`
		Algo    string            `json:"algo"`
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return recsysvc.TenantConfig{}, err
		}
	}
	return recsysvc.TenantConfig{
		TenantID: tenantID,
		Surface:  surface,
		Version:  etag,
		Weights:  payload.Weights,
		Flags:    payload.Flags,
		Algo:     strings.TrimSpace(payload.Algo),
	}, nil
}

// TenantRulesStore is a Postgres-backed adapter for tenant rules.
type TenantRulesStore struct {
	Pool ports.DatabasePool
}

func NewTenantRulesStore(pool ports.DatabasePool) *TenantRulesStore {
	return &TenantRulesStore{Pool: pool}
}

// GetRules returns the current rules and version for the tenant.
func (s *TenantRulesStore) GetRules(ctx context.Context, tenantID, surface string) (recsysvc.TenantRules, error) {
	if strings.TrimSpace(tenantID) == "" {
		return recsysvc.TenantRules{}, recsysvc.ErrRulesNotFound
	}
	db := txpostgres.FromCtx(ctx, s.Pool)
	const q = `
select v.rules, v.etag
  from tenants t
  join tenant_rules_current c on c.tenant_id = t.id
  join tenant_rule_versions v on v.id = c.rules_version_id
 where t.external_id = $1 or t.id::text = $1
`
	var raw []byte
	var etag string
	if err := db.QueryRow(ctx, q, tenantID).Scan(&raw, &etag); err != nil {
		if txpostgres.IsNoRows(err) {
			return recsysvc.TenantRules{}, recsysvc.ErrRulesNotFound
		}
		return recsysvc.TenantRules{}, err
	}
	return recsysvc.TenantRules{
		TenantID: tenantID,
		Surface:  surface,
		Version:  etag,
		Raw:      json.RawMessage(raw),
	}, nil
}
