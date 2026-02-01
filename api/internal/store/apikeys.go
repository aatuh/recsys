package store

import (
	"context"
	"time"

	"github.com/aatuh/recsys-suite/api/internal/auth"

	"github.com/aatuh/api-toolkit-contrib/adapters/txpostgres"
	"github.com/aatuh/api-toolkit/ports"
)

// APIKeyStore resolves tenant-scoped API keys from Postgres.
type APIKeyStore struct {
	Pool ports.DatabasePool
}

// NewAPIKeyStore constructs an APIKeyStore.
func NewAPIKeyStore(pool ports.DatabasePool) *APIKeyStore {
	return &APIKeyStore{Pool: pool}
}

// Lookup resolves the API key metadata from its hash.
func (s *APIKeyStore) Lookup(ctx context.Context, hash string) (auth.APIKey, error) {
	if s == nil || s.Pool == nil {
		return auth.APIKey{}, auth.ErrAPIKeyNotFound
	}
	db := txpostgres.FromCtx(ctx, s.Pool)
	const q = `
select k.id::text,
       k.tenant_id::text,
       t.external_id,
       coalesce(k.name, ''),
       k.roles,
       k.expires_at,
       k.revoked_at
  from tenant_api_keys k
  join tenants t on t.id = k.tenant_id
 where k.key_hash = $1
`
	var key auth.APIKey
	var roles []string
	var expiresAt *time.Time
	var revokedAt *time.Time
	if err := db.QueryRow(ctx, q, hash).Scan(
		&key.ID,
		&key.TenantID,
		&key.TenantExternalID,
		&key.Name,
		&roles,
		&expiresAt,
		&revokedAt,
	); err != nil {
		if txpostgres.IsNoRows(err) {
			return auth.APIKey{}, auth.ErrAPIKeyNotFound
		}
		return auth.APIKey{}, err
	}
	if revokedAt != nil {
		return auth.APIKey{}, auth.ErrAPIKeyRevoked
	}
	if expiresAt != nil && expiresAt.Before(time.Now().UTC()) {
		return auth.APIKey{}, auth.ErrAPIKeyExpired
	}
	key.Roles = roles
	return key, nil
}
