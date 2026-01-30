package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"recsys/internal/admin"
	"recsys/internal/services/adminsvc"

	"github.com/aatuh/api-toolkit-contrib/adapters/txpostgres"
	"github.com/aatuh/api-toolkit/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// AdminStore provides admin/control-plane persistence.
type AdminStore struct {
	Pool ports.DatabasePool
}

// NewAdminStore constructs a new AdminStore.
func NewAdminStore(pool ports.DatabasePool) *AdminStore {
	return &AdminStore{Pool: pool}
}

// ResolveTenantID resolves a tenant external ID to a UUID.
func (s *AdminStore) ResolveTenantID(ctx context.Context, tenantID string) (uuid.UUID, error) {
	if s == nil || s.Pool == nil {
		return uuid.Nil, admin.ErrTenantNotFound
	}
	return resolveTenantIDByExternal(ctx, txpostgres.FromCtx(ctx, s.Pool), tenantID)
}

// GetTenantConfig returns the current tenant config document.
func (s *AdminStore) GetTenantConfig(ctx context.Context, tenantID string) (adminsvc.TenantConfig, error) {
	db := txpostgres.FromCtx(ctx, s.Pool)
	tenantUUID, err := resolveTenantIDByExternal(ctx, db, tenantID)
	if err != nil {
		return adminsvc.TenantConfig{}, err
	}
	const q = `
select v.config, v.etag
  from tenant_configs_current c
  join tenant_config_versions v on v.id = c.config_version_id
 where c.tenant_id = $1
`
	var raw []byte
	var etag string
	if err := db.QueryRow(ctx, q, tenantUUID).Scan(&raw, &etag); err != nil {
		if txpostgres.IsNoRows(err) {
			return adminsvc.TenantConfig{}, admin.ErrConfigNotFound
		}
		return adminsvc.TenantConfig{}, err
	}
	return adminsvc.TenantConfig{
		TenantID: tenantID,
		Version:  etag,
		Raw:      json.RawMessage(raw),
	}, nil
}

// GetTenantRules returns the current tenant rules document.
func (s *AdminStore) GetTenantRules(ctx context.Context, tenantID string) (adminsvc.TenantRules, error) {
	db := txpostgres.FromCtx(ctx, s.Pool)
	tenantUUID, err := resolveTenantIDByExternal(ctx, db, tenantID)
	if err != nil {
		return adminsvc.TenantRules{}, err
	}
	const q = `
select v.rules, v.etag
  from tenant_rules_current c
  join tenant_rule_versions v on v.id = c.rules_version_id
 where c.tenant_id = $1
`
	var raw []byte
	var etag string
	if err := db.QueryRow(ctx, q, tenantUUID).Scan(&raw, &etag); err != nil {
		if txpostgres.IsNoRows(err) {
			return adminsvc.TenantRules{}, admin.ErrRulesNotFound
		}
		return adminsvc.TenantRules{}, err
	}
	return adminsvc.TenantRules{
		TenantID: tenantID,
		Version:  etag,
		Raw:      json.RawMessage(raw),
	}, nil
}

// UpdateTenantConfig updates tenant config with optimistic concurrency and audit.
func (s *AdminStore) UpdateTenantConfig(
	ctx context.Context,
	tenantID string,
	raw []byte,
	ifMatch string,
	actor adminsvc.Actor,
	meta adminsvc.RequestMeta,
) (adminsvc.TenantConfig, error) {
	manager := txpostgres.New(s.Pool)
	var result adminsvc.TenantConfig
	err := manager.WithinTx(ctx, func(txCtx context.Context) error {
		db := txpostgres.FromCtx(txCtx, s.Pool)
		tenantUUID, err := resolveTenantIDByExternal(txCtx, db, tenantID)
		if err != nil {
			return err
		}
		currentETag, beforeRaw, err := loadCurrentConfig(txCtx, db, tenantUUID, ifMatch)
		if err != nil {
			return err
		}
		actorID, actorType := normalizeActor(actor)
		const upsertQ = `
insert into tenant_configs_current (tenant_id, config_version_id, updated_by_sub)
values ($1, $2, $3)
on conflict (tenant_id)
do update set config_version_id = excluded.config_version_id,
              updated_by_sub = excluded.updated_by_sub,
              updated_at = now();
`
		const insertQ = `
insert into tenant_config_versions (tenant_id, config, created_by_sub)
values ($1, $2, $3)
on conflict (tenant_id, etag) do nothing
returning id, etag;
`
		var versionID uuid.UUID
		var etag string
		if err := db.QueryRow(txCtx, insertQ, tenantUUID, raw, actorID).Scan(&versionID, &etag); err != nil {
			if txpostgres.IsNoRows(err) {
				matchID, matchETag, err := findConfigVersion(txCtx, db, tenantUUID, raw)
				if err != nil {
					return err
				}
				if matchID == uuid.Nil {
					return errors.New("config version conflict but no matching version found")
				}
				if matchETag == currentETag {
					result = adminsvc.TenantConfig{
						TenantID: tenantID,
						Version:  matchETag,
						Raw:      json.RawMessage(raw),
					}
					return nil
				}
				if _, err := db.Exec(txCtx, upsertQ, tenantUUID, matchID, actorID); err != nil {
					return err
				}
				if err := insertAudit(txCtx, db, tenantUUID, actorID, actorType, meta, "tenant_config", matchID.String(), "config.update", beforeRaw, raw); err != nil {
					return err
				}
				result = adminsvc.TenantConfig{
					TenantID: tenantID,
					Version:  matchETag,
					Raw:      json.RawMessage(raw),
				}
				return nil
			}
			return err
		}
		if _, err := db.Exec(txCtx, upsertQ, tenantUUID, versionID, actorID); err != nil {
			return err
		}
		if err := insertAudit(txCtx, db, tenantUUID, actorID, actorType, meta, "tenant_config", versionID.String(), "config.update", beforeRaw, raw); err != nil {
			return err
		}
		result = adminsvc.TenantConfig{
			TenantID: tenantID,
			Version:  etag,
			Raw:      json.RawMessage(raw),
		}
		return nil
	})
	return result, err
}

// UpdateTenantRules updates tenant rules with optimistic concurrency and audit.
func (s *AdminStore) UpdateTenantRules(
	ctx context.Context,
	tenantID string,
	raw []byte,
	ifMatch string,
	actor adminsvc.Actor,
	meta adminsvc.RequestMeta,
) (adminsvc.TenantRules, error) {
	manager := txpostgres.New(s.Pool)
	var result adminsvc.TenantRules
	err := manager.WithinTx(ctx, func(txCtx context.Context) error {
		db := txpostgres.FromCtx(txCtx, s.Pool)
		tenantUUID, err := resolveTenantIDByExternal(txCtx, db, tenantID)
		if err != nil {
			return err
		}
		currentETag, beforeRaw, err := loadCurrentRules(txCtx, db, tenantUUID, ifMatch)
		if err != nil {
			return err
		}
		actorID, actorType := normalizeActor(actor)
		const upsertQ = `
insert into tenant_rules_current (tenant_id, rules_version_id, updated_by_sub)
values ($1, $2, $3)
on conflict (tenant_id)
do update set rules_version_id = excluded.rules_version_id,
              updated_by_sub = excluded.updated_by_sub,
              updated_at = now();
`
		const insertQ = `
insert into tenant_rule_versions (tenant_id, rules, created_by_sub)
values ($1, $2, $3)
on conflict (tenant_id, etag) do nothing
returning id, etag;
`
		var versionID uuid.UUID
		var etag string
		if err := db.QueryRow(txCtx, insertQ, tenantUUID, raw, actorID).Scan(&versionID, &etag); err != nil {
			if txpostgres.IsNoRows(err) {
				matchID, matchETag, err := findRulesVersion(txCtx, db, tenantUUID, raw)
				if err != nil {
					return err
				}
				if matchID == uuid.Nil {
					return errors.New("rules version conflict but no matching version found")
				}
				if matchETag == currentETag {
					result = adminsvc.TenantRules{
						TenantID: tenantID,
						Version:  matchETag,
						Raw:      json.RawMessage(raw),
					}
					return nil
				}
				if _, err := db.Exec(txCtx, upsertQ, tenantUUID, matchID, actorID); err != nil {
					return err
				}
				if err := insertAudit(txCtx, db, tenantUUID, actorID, actorType, meta, "tenant_rules", matchID.String(), "rules.update", beforeRaw, raw); err != nil {
					return err
				}
				result = adminsvc.TenantRules{
					TenantID: tenantID,
					Version:  matchETag,
					Raw:      json.RawMessage(raw),
				}
				return nil
			}
			return err
		}
		if _, err := db.Exec(txCtx, upsertQ, tenantUUID, versionID, actorID); err != nil {
			return err
		}
		if err := insertAudit(txCtx, db, tenantUUID, actorID, actorType, meta, "tenant_rules", versionID.String(), "rules.update", beforeRaw, raw); err != nil {
			return err
		}
		result = adminsvc.TenantRules{
			TenantID: tenantID,
			Version:  etag,
			Raw:      json.RawMessage(raw),
		}
		return nil
	})
	return result, err
}

// InsertCacheInvalidation records a cache invalidation request.
func (s *AdminStore) InsertCacheInvalidation(ctx context.Context, event adminsvc.CacheInvalidationEvent) error {
	if s == nil || s.Pool == nil {
		return nil
	}
	db := txpostgres.FromCtx(ctx, s.Pool)
	actorID, _ := normalizeActor(adminsvc.Actor{ID: event.ActorID})
	status := strings.TrimSpace(event.Status)
	if status == "" {
		status = "requested"
	}
	const q = `
insert into cache_invalidation_events (
  tenant_id,
  request_id,
  requested_by_sub,
  targets,
  surface,
  status,
  applied_at,
  applied_by,
  error_detail
) values (
  $1, $2, $3, $4, $5, $6::cache_invalidation_status,
  case when $6::cache_invalidation_status = 'applied'::cache_invalidation_status then now() else null end,
  case when $6::cache_invalidation_status = 'applied'::cache_invalidation_status then $3 else null end,
  $7
);
`
	_, err := db.Exec(ctx, q, event.TenantID, event.RequestID, actorID, event.Targets, event.Surface, status, event.ErrorDetail)
	if err != nil && (isUndefinedRelation(err) || isUndefinedObject(err)) {
		return nil
	}
	return err
}

func resolveTenantIDByExternal(ctx context.Context, db txpostgres.DBer, tenantID string) (uuid.UUID, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return uuid.Nil, admin.ErrTenantNotFound
	}
	const q = `select id from tenants where external_id = $1 or id::text = $1`
	var id uuid.UUID
	if err := db.QueryRow(ctx, q, tenantID).Scan(&id); err != nil {
		if txpostgres.IsNoRows(err) {
			return uuid.Nil, admin.ErrTenantNotFound
		}
		return uuid.Nil, err
	}
	return id, nil
}

func loadCurrentConfig(ctx context.Context, db txpostgres.DBer, tenantUUID uuid.UUID, ifMatch string) (string, []byte, error) {
	ifMatch = normalizeETag(ifMatch)
	const q = `
select v.etag, v.config
  from tenant_configs_current c
  join tenant_config_versions v on v.id = c.config_version_id
 where c.tenant_id = $1
 for update;
`
	var etag string
	var raw []byte
	err := db.QueryRow(ctx, q, tenantUUID).Scan(&etag, &raw)
	if err != nil {
		if txpostgres.IsNoRows(err) {
			if ifMatch != "" {
				return "", nil, admin.ErrVersionMismatch
			}
			return "", nil, nil
		}
		return "", nil, err
	}
	if ifMatch != "" && etag != ifMatch {
		return etag, raw, admin.ErrVersionMismatch
	}
	return etag, raw, nil
}

func loadCurrentRules(ctx context.Context, db txpostgres.DBer, tenantUUID uuid.UUID, ifMatch string) (string, []byte, error) {
	ifMatch = normalizeETag(ifMatch)
	const q = `
select v.etag, v.rules
  from tenant_rules_current c
  join tenant_rule_versions v on v.id = c.rules_version_id
 where c.tenant_id = $1
 for update;
`
	var etag string
	var raw []byte
	err := db.QueryRow(ctx, q, tenantUUID).Scan(&etag, &raw)
	if err != nil {
		if txpostgres.IsNoRows(err) {
			if ifMatch != "" {
				return "", nil, admin.ErrVersionMismatch
			}
			return "", nil, nil
		}
		return "", nil, err
	}
	if ifMatch != "" && etag != ifMatch {
		return etag, raw, admin.ErrVersionMismatch
	}
	return etag, raw, nil
}

func findConfigVersion(ctx context.Context, db txpostgres.DBer, tenantUUID uuid.UUID, raw []byte) (uuid.UUID, string, error) {
	const q = `
select id, etag
  from tenant_config_versions
 where tenant_id = $1
   and config = $2::jsonb
 limit 1;
`
	var id uuid.UUID
	var etag string
	if err := db.QueryRow(ctx, q, tenantUUID, raw).Scan(&id, &etag); err != nil {
		if txpostgres.IsNoRows(err) {
			return uuid.Nil, "", nil
		}
		return uuid.Nil, "", err
	}
	return id, etag, nil
}

func findRulesVersion(ctx context.Context, db txpostgres.DBer, tenantUUID uuid.UUID, raw []byte) (uuid.UUID, string, error) {
	const q = `
select id, etag
  from tenant_rule_versions
 where tenant_id = $1
   and rules = $2::jsonb
 limit 1;
`
	var id uuid.UUID
	var etag string
	if err := db.QueryRow(ctx, q, tenantUUID, raw).Scan(&id, &etag); err != nil {
		if txpostgres.IsNoRows(err) {
			return uuid.Nil, "", nil
		}
		return uuid.Nil, "", err
	}
	return id, etag, nil
}

func isUndefinedRelation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "42P01"
	}
	return false
}

func isUndefinedObject(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "42704"
	}
	return false
}

func normalizeETag(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"")
	return strings.TrimSpace(value)
}

func normalizeActor(actor adminsvc.Actor) (string, string) {
	actorID := strings.TrimSpace(actor.ID)
	actorType := strings.TrimSpace(actor.Type)
	if actorID == "" {
		actorID = "system"
	}
	if actorType == "" {
		actorType = "user"
		if actorID == "system" {
			actorType = "system"
		}
	}
	return actorID, actorType
}

func insertAudit(
	ctx context.Context,
	db txpostgres.DBer,
	tenantID uuid.UUID,
	actorID string,
	actorType string,
	meta adminsvc.RequestMeta,
	entityType string,
	entityID string,
	action string,
	beforeRaw []byte,
	afterRaw []byte,
) error {
	if actorID == "" {
		return nil
	}
	var requestUUID *uuid.UUID
	if meta.RequestID != "" {
		if id, err := uuid.Parse(meta.RequestID); err == nil {
			requestUUID = &id
		}
	}
	const q = `
insert into audit_log (
  tenant_id,
  actor_sub,
  actor_type,
  action,
  entity_type,
  entity_id,
  request_id,
  ip,
  user_agent,
  before_state,
  after_state
) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
`
	_, err := db.Exec(ctx, q, tenantID, actorID, actorType, action, entityType, entityID, requestUUID, meta.IP, meta.UserAgent, beforeRaw, afterRaw)
	return err
}

var _ adminsvc.Store = (*AdminStore)(nil)
