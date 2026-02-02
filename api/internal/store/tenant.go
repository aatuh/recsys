package store

import (
	"context"
	"strings"

	"github.com/aatuh/api-toolkit/contrib/v2/adapters/txpostgres"
	"github.com/aatuh/api-toolkit/v2/authorization"
	"github.com/aatuh/api-toolkit/v2/ports"
	"github.com/google/uuid"
)

func resolveTenantID(ctx context.Context, pool ports.DatabasePool, orgID uuid.UUID) (uuid.UUID, error) {
	if pool == nil {
		return orgID, nil
	}
	externalID := ""
	if ctx != nil {
		if v, ok := authorization.TenantIDFromContext(ctx); ok {
			externalID = strings.TrimSpace(v)
		}
	}
	if externalID != "" {
		db := txpostgres.FromCtx(ctx, pool)
		const q = `select id from tenants where external_id = $1 or id::text = $1`
		var id uuid.UUID
		if err := db.QueryRow(ctx, q, externalID).Scan(&id); err != nil {
			if txpostgres.IsNoRows(err) {
				return orgID, nil
			}
			return uuid.Nil, err
		}
		return id, nil
	}
	return orgID, nil
}
