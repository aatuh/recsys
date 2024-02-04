# Runbook: Roll back config/rules

This runbook covers **rolling back tenant config and rules** in `recsys-service`.

If you need to roll back the **artifact manifest** (artifact mode), see:
[Roll back the manifest](../../recsys-pipelines/docs/how-to/rollback-manifest.md).

## Symptoms

- A recent config/rules change caused empty recs, regressions, or elevated errors
- You need to revert a tenant to the last-known-good behavior quickly

## What “version” means here

`config_version` and `rules_version` are **ETags** derived from the JSON payload
(SHA-256 of the document). Rolling back to an earlier document means setting the
current pointer back to that exact JSON payload.

## Preferred rollback path (audited, API-only)

This path is safe and leaves an audit trail.

### 1) Capture the current state

- `GET /v1/admin/tenants/{tenant_id}/config` → record `config_version` (and `ETag` header)
- `GET /v1/admin/tenants/{tenant_id}/rules` → record `rules_version` (and `ETag` header)

Also capture a failing `POST /v1/recommend` response `meta` for the incident record.

### 2) Find the last-known-good payload (Audit Log)

Use:

- `GET /v1/admin/tenants/{tenant_id}/audit?limit=50`

Look for:

- `action=config.update` (config changes)
- `action=rules.update` (rules changes)

For the entry you want to revert, copy the JSON from:

- `before_state` (the previous payload)

### 3) Re-apply the previous payload

PUT the previous payload back:

- `PUT /v1/admin/tenants/{tenant_id}/config` with body = `before_state`
- `PUT /v1/admin/tenants/{tenant_id}/rules` with body = `before_state`

Use `If-Match` (recommended) to avoid overwriting a concurrent update:

- `If-Match: <current config_version>` / `If-Match: <current rules_version>`

If you get `409 RECSYS_VERSION_MISMATCH`, re-run step (1) and try again.

### 4) Invalidate caches

- `POST /v1/admin/tenants/{tenant_id}/cache/invalidate` with:

  ```json
  { "targets": ["config", "rules", "popularity"] }
  ```

Notes:

- `popularity` is relevant in artifact/manifest mode (no-op in DB-only mode).

### 5) Verify

1) Call `POST /v1/recommend` and confirm `meta.config_version` and
   `meta.rules_version` match the expected rolled-back versions.
2) Watch key metrics (error rate, latency, empty-recs rate) for recovery.

## Break-glass rollback (direct DB pointer update)

Only use this if the admin API is unavailable and you have DBA-level access.

!!! warning
    Direct DB changes can cause outages. Prefer the audited API rollback path above.

High-level steps:

1) Identify the tenant UUID:

   ```sql
   select id from tenants where external_id = 'demo';
   ```

2) Pick the target version IDs (for example, the most recent two versions):

   ```sql
   select id, etag, created_at
     from tenant_config_versions
    where tenant_id = :tenant_uuid
    order by created_at desc
    limit 5;
   ```

3) Update the current pointer in a transaction:

   ```sql
   begin;
   update tenant_configs_current
      set config_version_id = :target_config_version_id,
          updated_by_sub = 'break-glass',
          updated_at = now()
    where tenant_id = :tenant_uuid;
   commit;
   ```

Repeat the same pattern for `tenant_rules_current`.
