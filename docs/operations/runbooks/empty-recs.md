# Runbook: Empty recs

## Symptoms

- `POST /v1/recommend` returns `200 OK` but `items` is empty (`[]`)
- The issue is tenant/surface specific (some tenants/surfaces work, others don’t)

## Decision tree (fast path)

```mermaid
flowchart TD
  A[items[] is empty] --> B{warnings[] explains why?}
  B -->|CANDIDATES_INCLUDE_EMPTY| C[Check candidates.include_ids / allow-lists]
  B -->|CONSTRAINTS_FILTERED| D[Relax constraints (tags, caps, price)]
  B -->|SIGNAL_UNAVAILABLE / SIGNAL_PARTIAL| E{Artifact mode enabled?}
  B -->|No warnings| F{config_version or rules_version empty?}

  F -->|Yes| G[Bootstrap tenant config/rules + invalidate caches]
  F -->|No| E

  E -->|Yes| H[Verify current manifest + artifact URIs are readable]
  E -->|No| I[Verify DB signals exist (item_popularity_daily, item_tags)]
```

## Quick triage (copy/paste)

Set:

```bash
BASE_URL=${BASE_URL:-http://localhost:8000}
TENANT_ID=demo
```

1) Validate the request shape (normalization + warnings):

   ```bash
   curl -fsS "$BASE_URL/v1/recommend/validate" \
     -H "Content-Type: application/json" \
     -H "X-Org-Id: $TENANT_ID" \
     -d '{"surface":"home","k":10,"user":{"user_id":"debug-user-1"}}'
   ```

2) Call recommend and inspect `meta` + `warnings`:

   ```bash
   curl -fsS "$BASE_URL/v1/recommend" \
     -H "Content-Type: application/json" \
     -H "X-Org-Id: $TENANT_ID" \
     -d '{"surface":"home","k":10,"user":{"user_id":"debug-user-1"}}'
   ```

If your environment requires auth headers, see [Admin API + local bootstrap](../../reference/api/admin.md).

## What to look for in the response

### 1) `meta.config_version` and `meta.rules_version`

If either is empty, the tenant may be missing config/rules.

- Confirm with:
  - `GET /v1/admin/tenants/{tenant_id}/config`
  - `GET /v1/admin/tenants/{tenant_id}/rules`

If config/rules were just updated, invalidate caches:

- `POST /v1/admin/tenants/{tenant_id}/cache/invalidate` with `{"targets":["config","rules"]}`.

### 2) `warnings[]`

Common warning codes that explain “empty recs”:

- `CANDIDATES_INCLUDE_EMPTY`: `candidates.include_ids` filtered everything out
- `CONSTRAINTS_FILTERED`: tag constraints filtered most/all results
- `SIGNAL_UNAVAILABLE` / `SIGNAL_PARTIAL`: one or more signals are missing or incomplete

## Likely causes (and checks)

### A) No popularity signal available (most common)

DB-only mode:

- Verify `item_popularity_daily` has rows for your tenant and surface (namespace):

  ```sql
  select count(*) from item_popularity_daily
  where tenant_id = (select id from tenants where external_id = 'demo')
    and namespace = 'home';
  ```

Artifact/manifest mode (`RECSYS_ARTIFACT_MODE_ENABLED=true`):

- Verify the manifest exists and points to readable artifacts for the tenant/surface.
- If artifacts were just published, invalidate caches:
  - `POST /v1/admin/tenants/{tenant_id}/cache/invalidate` with `{"targets":["popularity"]}`.

See [Data modes](../../explanation/data-modes.md) for how DB-only vs artifact mode works.

### B) Constraints or allow-lists filtered everything

Checks:

- If you set `constraints.required_tags`, make sure items actually carry those tags.
- If you set `constraints.forbidden_tags` or `constraints.max_per_tag`, relax them temporarily.
- If you set `candidates.include_ids`, confirm those item IDs exist in your catalog/signals.

DB-only mode tag lookup depends on `item_tags`:

- Confirm the items you expect are tagged under the same `namespace` (surface).

### C) Rules blocked everything

Checks:

- `GET /v1/admin/tenants/{tenant_id}/rules` and look for high-priority `block` rules.
- Verify the response `meta.rules_version` matches what you think is active.

Safe remediation:

- Roll back rules to a last-known-good version (see [Runbook: Roll back config/rules](rollback-config-rules.md)).

## If you need a fast “is the system alive?” sanity check

Temporarily remove constraints/allow-lists and request a small `k`:

```bash
curl -fsS "$BASE_URL/v1/recommend" \
  -H "Content-Type: application/json" \
  -H "X-Org-Id: $TENANT_ID" \
  -d '{"surface":"home","k":5,"user":{"user_id":"debug-user-1"},"constraints":null,"candidates":null}'
```
