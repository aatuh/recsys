---
tags:
  - reference
  - database
  - developer
  - ops
  - recsys-service
---

# Database schema

## Who this is for

- Operators provisioning Postgres for `recsys-service`
- Developers debugging "empty recs" and data wiring in DB-only mode

## What you will get

- A high-level map of the tables used by the serving stack
- Which tables are high-volume and how they scale (partitioning guidance)
- Copy/paste queries to inspect the schema quickly

## Reference

`recsys-service` uses Postgres for:

- tenant registry (`tenants`)
- versioned tenant config and rules (plus "current pointers")
- exposure logging (for evaluation and debugging)
- audit logging (admin actions)
- signal tables (DB-only mode)

### Tenants

- `tenants`: one row per tenant/org. `external_id` typically matches the tenant claim or header value.

### Config and rules (versioned + current pointers)

- `tenant_config_versions`: append-only config versions (JSONB) with an ETag hash
- `tenant_configs_current`: pointer to the current config version
- `tenant_rule_versions`: append-only rule set versions (JSONB) with an ETag hash
- `tenant_rules_current`: pointer to the current rules version

### Exposure and interaction logs (high volume)

- `exposure_events`: partitioned by `occurred_at` (RANGE partitions) for scale
- `interaction_events`: user actions tied to `request_id` (used for analysis/debugging)

Operational note: keep partitions created ahead of time so inserts never fail due to missing partitions.

### DB-only signal tables (serving reads these directly)

- `item_tags`: item metadata/tags by `(tenant_id, namespace, item_id)`
- `item_popularity_daily`: daily popularity scores by `(tenant_id, namespace, item_id, day)`
- `item_covisit_daily`: daily co-visitation by `(tenant_id, namespace, item_id, neighbor_id, day)`

## Examples

List tables:

```bash
docker exec -i recsys-db psql -U recsys-db -d recsys-db -c "\\dt"
```

Inspect a table:

```bash
docker exec -i recsys-db psql -U recsys-db -d recsys-db -c "\\d+ tenants"
```

Check exposure table partitions:

```bash
docker exec -i recsys-db psql -U recsys-db -d recsys-db -c "select relname from pg_class where relname like 'exposure_events%';"
```

## Read next

- Migrations (safe upgrade): [`reference/database/migrations.md`](migrations.md)
- DB-only seed examples: [`reference/database/db-only-seeding.md`](db-only-seeding.md)
- Empty recs runbook: [`operations/runbooks/empty-recs.md`](../../operations/runbooks/empty-recs.md)
