# API Reference

This page is the **integration hub** for `recsys-service`:

- Full schema: [`openapi.yaml`](openapi.yaml)
- Practical examples: [`examples.md`](examples.md)
- Error handling & troubleshooting: [`errors.md`](errors.md)
- Admin/control-plane + local bootstrap: [`admin.md`](admin.md)

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

Tenant scope:

- In production, tenant context typically comes from a JWT claim (see `AUTH_TENANT_CLAIMS`).
- When tenant scope is not derived from auth, send the tenant header (default `X-Org-Id`).

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
