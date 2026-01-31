# How-to: integrate recsys-service into an application

1) Define your surfaces (home, pdp, checkout) and keep names stable
2) Send stable pseudonymous user/session identifiers
3) Call POST /v1/recommend and render the list
4) Log outcomes (click/purchase) linked by request_id
5) Use /v1/recommend/validate during development
6) Handle failures: empty list fallback, respect 429 Retry-After

Notes:
- `surface` also acts as the **namespace** for signals and rules.
- For local MVPs, a `default` namespace fallback is available (see `explanation/surface-namespaces.md`).
- Admin bootstrap (tenant + config + rules) is required before first use:
  see `reference/api/admin.md`.

Tenant headers (local dev):
- When `DEV_AUTH_ENABLED=true`, send **both**:
  - `X-Dev-Org-Id` (dev auth tenant context)
  - `X-Org-Id` (tenant scope enforced by middleware)
- In JWT mode, a bearer token with a tenant claim is sufficient (see `AUTH_TENANT_CLAIMS`).
- To use a single header locally, set `DEV_AUTH_TENANT_HEADER=X-Org-Id`.
