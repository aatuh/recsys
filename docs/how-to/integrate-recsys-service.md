# How-to: integrate recsys-service into an application

## Who this is for

Backend and platform engineers integrating `recsys-service` into a product (webshop, content feed, etc.).

## Goal

Call `POST /v1/recommend`, render the list, and log outcomes with correct attribution so evaluation and troubleshooting
work.

## Prereqs

- A small stable set of [surfaces](../project/glossary.md#surface), e.g. `home`, `pdp`, `cart`
- Stable pseudonymous identifiers (`user_id` and/or `session_id`)
- Admin bootstrap completed for the tenant (tenant + config + rules)

## Steps

1) Define your surfaces (home, pdp, checkout) and keep names stable.
2) Send stable pseudonymous user/session identifiers.
3) Call `POST /v1/recommend` and render the ranked list.
4) Log outcomes (click/purchase) linked by `request_id`.
5) Use `POST /v1/recommend/validate` during development to catch warnings early.
6) Handle failures: empty list fallback; respect `429 Retry-After` under load.

Notes:

- `surface` also acts as the signal/rules [namespace](../project/glossary.md#namespace).
- For local MVPs, a `default` namespace fallback is available (see `explanation/surface-namespaces.md`).
- Admin bootstrap (tenant + config + rules) is required before first use:

  see `reference/api/admin.md`.
- If you want a domain-specific mental model, start with the cookbooks:
  - [`how-to/integration-cookbooks/index.md`](integration-cookbooks/index.md)

## Verify

Validate one request shape:

```bash
curl -fsS http://localhost:8000/v1/recommend/validate \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: integ-req-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1"}}'
```

Then call recommend and ensure you capture `meta.request_id` (or supply `X-Request-Id`) and propagate it into outcomes.

## Pitfalls

### Tenant headers (local dev)

- When `DEV_AUTH_ENABLED=true`, send **both**:
  - `X-Dev-Org-Id` (dev auth tenant context)
  - `X-Org-Id` (tenant scope enforced by middleware)
- In JWT mode, a bearer token with a tenant claim is sufficient (see `AUTH_TENANT_CLAIMS`).
- To use a single header locally, set `DEV_AUTH_TENANT_HEADER=X-Org-Id`.

## Read next

- Exposure logging & attribution: [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)
- Admin bootstrap (tenant + config + rules): [`reference/api/admin.md`](../reference/api/admin.md)
- Integration cookbooks: [`how-to/integration-cookbooks/index.md`](integration-cookbooks/index.md)
