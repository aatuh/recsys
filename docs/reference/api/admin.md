# Admin API + local bootstrap (recsys-service)

This page documents the **admin/control-plane** endpoints and the minimum
bootstrap steps required to call `/v1/recommend` and `/v1/similar`.

Why this exists:

- The OpenAPI file (`reference/api/openapi.yaml`) documents the HTTP surface.
- This page adds **bootstrap guidance**, examples, and operational notes for admin/control-plane usage.

## 0) Prereqs

- Postgres is running and migrations are applied.
- recsys-service is running and reachable (e.g. `http://localhost:8000`).

## 1) Create a tenant row (DB bootstrap)

Admin endpoints require a tenant record in `tenants`. There is no admin API to
create tenants yet, so insert directly in Postgres:

```sql
insert into tenants (external_id, name)
values ('demo', 'Demo Tenant')
on conflict (external_id) do nothing;
```

Notes:

- `external_id` should match the tenant/org claim in your JWT, or the dev tenant

  header value (see below).

- You can also use the tenant UUID in admin paths; `external_id` is preferred.

## 2) Auth + tenancy (local dev)

Local dev can use **dev headers** instead of JWT:

Set in `.env`:

```bash
AUTH_REQUIRED=true
AUTH_REQUIRE_TENANT_CLAIM=false
DEV_AUTH_ENABLED=true
DEV_AUTH_USER_ID_HEADER=X-Dev-User-Id
DEV_AUTH_TENANT_HEADER=X-Dev-Org-Id
AUTH_ADMIN_ROLE=   # empty to disable admin role checks locally
```

Then send headers on every request:

```text
X-Dev-User-Id: dev-user-1
X-Dev-Org-Id: demo
X-Org-Id: demo   # must match tenant scope
```

Why two tenant headers?

- `X-Dev-Org-Id` is used to **derive tenant context** in local/dev mode.
- `X-Org-Id` is the tenant header enforced by the tenant middleware.

Tip (single header in local dev):

- Set `DEV_AUTH_TENANT_HEADER=X-Org-Id` to use **one** header for both dev auth

  and tenant scope.

If you prefer JWT:

- Provide a bearer token with a tenant claim (see `AUTH_TENANT_CLAIMS`).
- Include an admin role (default `admin`) under a role claim (see

  `AUTH_ROLE_CLAIMS`) to access admin endpoints.

RBAC roles (JWT or API keys):

- `viewer`: read-only admin access (GET config/rules/audit).
- `operator`: config/rules updates + cache invalidation.
- `admin`: full admin access (includes operator + viewer).

## 3) Admin endpoints

All admin endpoints are under `/v1/admin`.

The `tenant_id` path parameter accepts **external_id or UUID**.

### GET /v1/admin/tenants/{tenant_id}/config

Returns current tenant config and `config_version`.

### PUT /v1/admin/tenants/{tenant_id}/config

Updates tenant config (optimistic concurrency).

Headers:

- `If-Match`: optional. Use the `config_version` from the latest GET response.

  Omit for the first insert.

Payload (minimal example):

```json
{
  "weights": { "pop": 0.7, "cooc": 0.2, "emb": 0.1 },
  "flags": { "enable_rules": true },
  "limits": { "max_k": 50, "max_exclude_ids": 200 }
}
```

Validation notes:

- weights must be non‑negative
- limits must be non‑negative

Common config keys (current behavior):

- `weights.pop`, `weights.cooc`, `weights.emb`
- `flags` (boolean map, free-form)
- `limits.max_k`, `limits.max_exclude_ids`

### GET /v1/admin/tenants/{tenant_id}/rules

Returns current tenant rules and `rules_version`.

### PUT /v1/admin/tenants/{tenant_id}/rules

Updates tenant rules (optimistic concurrency).

Headers:

- `If-Match`: optional. Use the `rules_version` from the latest GET response.

  Omit for the first insert.

Payload (minimal example array):

```json
[
  {
    "action": "pin",
    "target_type": "item",
    "item_ids": ["item_1", "item_2"],
    "surface": "home",
    "priority": 10
  },
  {
    "action": "block",
    "target_type": "tag",
    "target_key": "brand:nike"
  }
]
```

Supported actions: `pin`, `boost`, `block` (aliases allowed: `promote`, `exclude`, etc).
Supported targets: `item`, `tag`, `brand`, `category`.

Common rule keys (current parser):

- `action`: `pin` | `boost` | `block`
- `target_type`: `item` | `tag` | `brand` | `category`
- `target_key`: string (for tag/brand/category)
- `item_ids`: array of item ids (for item targeting)
- `namespace` (optional), `surface` (optional), `segment` (optional)
- `priority` (int), `enabled` (bool)
- `valid_from`, `valid_until` (RFC3339 timestamps)
- `boost_value` (number, when `action=boost`)
- `max_pins` (int, when `action=pin`)

### POST /v1/admin/tenants/{tenant_id}/cache/invalidate

Payload:

```json
{ "targets": ["config", "rules"], "surface": "home" }
```

Valid targets: `config`, `rules`, `popularity`.

Notes:

- `surface` is optional. If provided, invalidation is scoped to that surface.
- `popularity` invalidates artifact/manifest caches (no‑op in DB‑only mode).

### GET /v1/admin/tenants/{tenant_id}/audit

Returns recent audit log entries for admin actions (write operations).

Query parameters:

- `limit` (optional, default 100, max 200)
- `before` (optional, RFC3339 timestamp for pagination)
- `before_id` (optional, numeric id for pagination tie‑break)

Example:

```http
GET /v1/admin/tenants/demo/audit?limit=50
```

Response includes `entries` and optional `next_before`/`next_before_id` cursor.

## 4) Call the API

Once config and rules exist, you can call `/v1/recommend` or `/v1/similar`
using the same tenant headers/token.

See also:

- `reference/api/examples/admin-config.http`
- `tutorials/local-end-to-end.md`
