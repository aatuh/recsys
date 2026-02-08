---
diataxis: how-to
tags:
  - how-to
  - integration
  - api
  - developer
  - recsys-service
---
# How-to: integrate recsys-service into an application
This guide shows how to how-to: integrate recsys-service into an application in a reliable, repeatable way.


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
  - [Integration cookbooks (map RecSys to your domain)](integration-cookbooks/index.md)

## Minimal integration example (one surface)

This is the smallest pattern that supports evaluation and troubleshooting.

### Language-agnostic pseudocode (request_id propagation)

```text
requestId = newRequestId()
resp = POST /v1/recommend(tenant, surface, user, requestId)
render(resp.items)
logExposure(requestId, tenant, surface, itemsShown)

onUserAction(itemId, eventType):
  logOutcome(requestId, itemId, eventType)
```

The one rule that unlocks evaluation and debugging is: **the same `request_id` must flow through recommend → exposure → outcome**.


### 1) Serve (request)

Generate one `request_id` per rendered list, then call recommend:

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: req-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":10,"user":{"user_id":"u_1","session_id":"s_1"}}'
```

Capture `meta.request_id` (or the header value you supplied) and propagate it to both exposures and outcomes.

### 2) Log exposures (what was shown)

Emit one exposure record per rendered list (schema: `exposure.v1`):

```json
{
  "request_id": "req-1",
  "items": [
    { "item_id": "item_1", "rank": 1 }
  ],
  "context": { "tenant_id": "demo", "surface": "home" }
}
```

### 3) Log outcomes (what the user did)

Emit one outcome record per action you measure (schema: `outcome.v1`):

```json
{
  "request_id": "req-1",
  "item_id": "item_1",
  "event_type": "click"
}
```

For the canonical schemas and required fields, see: [Data contracts](../reference/data-contracts/index.md)

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
- Auth and tenancy reference: [Auth and tenancy reference](../reference/auth-and-tenancy.md)

## Troubleshooting

If something looks wrong (empty lists, auth issues, logs not joining), use:

- [Troubleshooting for integrators](troubleshooting-integration.md)

## Read next

- Integration checklist: [How-to: Integration checklist (one surface)](integration-checklist.md)
- Troubleshoot empty lists, auth, or join issues: [How-to: troubleshooting for integrators](troubleshooting-integration.md)
- Exposure logging & attribution: [Exposure logging and attribution](../explanation/exposure-logging-and-attribution.md)
- Minimum instrumentation spec (what to log for credible eval): [Minimum instrumentation spec (for credible evaluation)](../reference/minimum-instrumentation.md)
- Auth & tenancy: [Auth and tenancy reference](../reference/auth-and-tenancy.md)
