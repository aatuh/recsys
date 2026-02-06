---
tags:
  - reference
  - api
  - developer
  - recsys-service
---

# API Reference

This page is the **integration hub** for `recsys-service`:

- Full schema: [`openapi.yaml`](openapi.yaml)
- Practical examples: [`examples.md`](examples.md)
- Error handling & troubleshooting: [`errors.md`](errors.md)
- Admin/control-plane + local bootstrap: [`admin.md`](admin.md)
- Auth & tenancy: [`reference/auth-and-tenancy.md`](../auth-and-tenancy.md)

## Who this is for

- Developers integrating the serving API (`POST /v1/recommend`, `POST /v1/similar`)
- Operators validating auth/tenancy and admin/control-plane access

## What you will get

- The base URL and versioning conventions
- A minimal request/response example you can run locally
- Pointers to the canonical schema, examples, and error semantics

## Base URL and versioning

- All API endpoints are under `/v1`.
- Health endpoints:
  - `/healthz` (liveness)
  - `/readyz` (readiness)
  - `/health/detailed` (debugging)

## Auth and tenancy (high level)

The service supports:

- `Authorization: Bearer <token>` (JWT)
- `X-API-Key: <key>` (API keys)

Local development can also use dev headers (see [`admin.md`](admin.md)).
For details (headers, claims, roles), see: [`reference/auth-and-tenancy.md`](../auth-and-tenancy.md).

Tenant scope:

- In production, tenant context typically comes from a JWT claim (see `AUTH_TENANT_CLAIMS`).
- When tenant scope is not derived from auth, send the tenant header (default `X-Org-Id`).

## Hello world (minimal request/response)

If you want the fastest path to a non-empty response, start with:

- Tutorial: [`tutorials/quickstart.md`](../../tutorials/quickstart.md)

Example request (local dev headers):

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: hello-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":3,"user":{"user_id":"u_1"}}'
```

Example response shape (abbreviated):

```json
{
  "items": [{ "item_id": "item_1", "rank": 1, "score": 0.12 }],
  "meta": { "tenant_id": "demo", "surface": "home", "request_id": "hello-1" },
  "warnings": []
}
```

## Request/response conventions

- Requests and responses use JSON.
- Recommendation responses include `meta`:
  - `config_version`, `rules_version` (ETags derived from JSON payloads)
  - `algo_version` (effective algorithm version label)
  - `request_id` (for support and log correlation)
- Responses may include non-fatal `warnings[]` (for example: missing signals or filtered candidates).

## Versioning

- Endpoints are versioned via the `/v1` path prefix.
- When you integrate, treat the response body as the contract (don’t depend on field ordering).

## Retries and idempotency

- `POST /v1/recommend` and `POST /v1/similar` are read-only, but **may have side effects** (for example: exposure logging).
- If you implement retries, ensure your integration does not accidentally double-count events downstream.

## Errors

Errors use [Problem Details](https://www.rfc-editor.org/rfc/rfc7807) with content type `application/problem+json`.

Common status codes you should handle:

- `400` invalid JSON
- `401/403` auth or scope failure
- `409` optimistic concurrency conflict (`If-Match` mismatch)
- `422` validation failure (semantically invalid request)
- `429` rate limit
- `503` overloaded or not ready

## Swagger UI

<swagger-ui src="./openapi.yaml"/>
