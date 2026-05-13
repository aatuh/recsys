package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/aatuh/api-toolkit/contrib/v2/adapters/txpostgres"
	"github.com/aatuh/api-toolkit/v2/ports"
)

var tenantScopedTables = []string{
	"tenant_config_versions",
	"tenant_configs_current",
	"tenant_rule_versions",
	"tenant_rules_current",
	"audit_log",
	"cache_invalidation_events",
	"exposure_events",
	"interaction_events",
	"item_popularity_daily",
	"item_covisit_daily",
	"item_tags_current",
}

// CheckTenantRLS fails when required tenant-scoped tables do not have RLS enabled.
func CheckTenantRLS(ctx context.Context, pool ports.DatabasePool) error {
	missing, err := TenantTablesMissingRLS(ctx, txpostgres.FromCtx(ctx, pool), tenantScopedTables)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return fmt.Errorf("tenant database guardrail failed: row-level security is not enabled on %s", strings.Join(missing, ", "))
	}
	return nil
}

// TenantTablesMissingRLS returns required tables that are missing row-level security.
func TenantTablesMissingRLS(ctx context.Context, db txpostgres.DBer, tables []string) ([]string, error) {
	if len(tables) == 0 {
		return nil, nil
	}
	const q = `
SELECT expected.table_name
FROM unnest($1::text[]) AS expected(table_name)
LEFT JOIN pg_class c
	ON c.relname = expected.table_name
LEFT JOIN pg_namespace n
	ON n.oid = c.relnamespace
	AND n.nspname = 'public'
WHERE COALESCE(c.relrowsecurity, false) = false
ORDER BY expected.table_name;`
	rows, err := db.Query(ctx, q, tables)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var missing []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		missing = append(missing, table)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return missing, nil
}
