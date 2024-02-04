# Error handling & troubleshooting API calls

## Who this is for

Integrators and on-call engineers who need to diagnose failed requests quickly and safely.

## What you will get

- The error format (`application/problem+json`)
- What the common HTTP status codes mean in this service
- Endpoint-specific “what to do next”

## Error format (Problem Details)

Errors use [RFC 7807](https://www.rfc-editor.org/rfc/rfc7807) with content type `application/problem+json`.

Typical fields:

- `type`, `title`, `status` (always present)
- `detail` (human-readable explanation; should be safe to show to end users)
- `code` (machine-readable error category, when available)
- `request_id` (for correlation)

Example:

```json
{
  "type": "about:blank",
  "title": "validation failed",
  "status": 422,
  "detail": "surface is required",
  "code": "VALIDATION_FAILED",
  "request_id": "a2e38779dfbe/nmiLt982Mq-000004"
}
```

Operational tip: include `request_id` in client logs and support tickets. It is the fastest way to find the matching
server log line.

## Common status codes (what to do)

- **400**: invalid JSON or wrong content type
  - Check you send `Content-Type: application/json` and valid JSON.
- **401/403**: authentication/authorization/tenant scope failure
  - Check auth headers, tenant header (`X-Org-Id`), and role requirements for admin endpoints.
- **404** (admin endpoints): tenant not found
  - Tenant creation is DB-only today; see [`reference/api/admin.md`](admin.md) and
    [`start-here/known-limitations.md`](../../start-here/known-limitations.md).
- **409** (admin `PUT`): optimistic concurrency conflict
  - Fetch the latest resource, take its ETag, and retry with `If-Match`.
- **422**: validation failure (semantically invalid request)
  - Call `POST /v1/recommend/validate` to see the normalized request + warnings.
- **429**: rate limited
  - Back off and retry. If this is unexpected, review per-tenant rate limits.
- **503**: not ready / overloaded
  - Check `GET /readyz` and verify dependencies (DB, artifact store if enabled).
- **500**: internal error
  - Use `request_id` to locate server logs; follow the relevant runbook.

## Endpoint notes

### `POST /v1/recommend` (and `POST /v1/similar`)

Expected error responses (see OpenAPI):

- `400`, `401`, `403`, `422`, `429`, `500`, `503`

What to do first:

1. Call `POST /v1/recommend/validate` with the same payload to surface normalization and warnings.
2. Confirm tenant scope (JWT claims or `X-Org-Id` header).
3. If you see empty results, use the “empty recs” runbook:
   [`operations/runbooks/empty-recs.md`](../../operations/runbooks/empty-recs.md)

### `POST /v1/recommend/validate`

This endpoint is your fastest “is my request shape sane?” tool.

Expected error responses:

- `400`, `401`, `403`, `422`, `429`

If you get `422`, fix the request payload before calling `/v1/recommend`.

### Admin endpoints (`/v1/admin/...`)

These endpoints are for operators (config, rules, cache invalidation, audit).

Common pitfalls:

- `401/403`: missing operator/admin privileges (or dev auth is not enabled)
- `409` on `PUT`: you updated with a stale version; retry with the latest `If-Match`
- `404`: tenant does not exist (bootstrap is DB-first today)

See: [`reference/api/admin.md`](admin.md)

### Health endpoints (`GET /healthz`, `GET /readyz`)

- `/healthz` is a liveness probe (is the process up?)
- `/readyz` is a readiness probe (are dependencies reachable?)

If `/readyz` returns `503`, use the “service not ready” runbook:
[`operations/runbooks/service-not-ready.md`](../../operations/runbooks/service-not-ready.md)
