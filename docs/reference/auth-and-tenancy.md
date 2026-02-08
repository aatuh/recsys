---
diataxis: reference
tags:
  - reference
  - api
  - security
  - developer
  - recsys-service
---
# Auth and tenancy reference
This page is the canonical reference for Auth and tenancy reference.


## Who this is for

- Developers integrating `recsys-service`
- Operators securing production deployments

## What you will get

- the supported auth modes (dev headers, JWT, API keys)
- how tenant scope is determined (claims vs headers)
- the practical “which headers do I send?” answers

## Concepts (2 minutes)

- **Tenant**: a logically separate recommendation domain
  (see [Glossary](../project/glossary.md#tenant)).
- **Surface**: the recommendation surface/placement
  (see [Glossary](../project/glossary.md#surface)).
- **Tenant scope**: which tenant a request is allowed to operate on.

In general:

- production: tenant scope should come from trusted auth (JWT/API key)
- local dev: tenant scope often comes from headers (dev headers + `X-Org-Id`)

## Auth modes

### Dev headers (local/test)

When `DEV_AUTH_ENABLED=true`, you can authenticate using headers.
The canonical env var list and defaults live in:
[recsys-service configuration](config/recsys-service.md).

Common headers:

- `X-Dev-User-Id`: who is calling (any stable string for local dev)
- `X-Dev-Org-Id`: dev-auth tenant context
- `X-Org-Id`: tenant scope (enforced by middleware unless tenant scope comes from a claim/key)

Tip: if you want to use only one tenant header locally, set:

- `DEV_AUTH_TENANT_HEADER=X-Org-Id`

Production guidance:

- disable dev headers (`DEV_AUTH_ENABLED=false`)

Example (local dev request):

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: example-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":3,"user":{"user_id":"u_1"}}'
```

### JWT (production)

When `JWT_AUTH_ENABLED=true`, send:

- `Authorization: Bearer <token>`

Tenant scope is derived from a claim configured via:

- `AUTH_TENANT_CLAIMS` (comma-separated claim keys)

Admin role checks use:

- `AUTH_VIEWER_ROLE`, `AUTH_OPERATOR_ROLE`, `AUTH_ADMIN_ROLE`
- `AUTH_ROLE_CLAIMS` (where roles/scopes are read from)

### API keys (production)

When `API_KEY_ENABLED=true`, send:

- `X-API-Key: <key>` (or the header configured by `API_KEY_HEADER`)

## Practical guidance (what to send)

### `POST /v1/recommend` (most integrations)

- Always send:
  - `Content-Type: application/json`
  - `X-Request-Id: <id>` (recommended; otherwise the service generates one)
  - `surface` (in the JSON body)
  - a pseudonymous `user_id` and/or `session_id` (in the JSON body)
- Tenant scope:
  - JWT/API key mode: tenant scope should come from trusted auth.
  - Dev header mode: send `X-Org-Id` (and `X-Dev-Org-Id` unless you changed `DEV_AUTH_TENANT_HEADER`).

### Admin endpoints (config/rules/audit)

Admin endpoints are control-plane: restrict network access and require roles in production.

- Docs: [Admin API + local bootstrap (recsys-service)](api/admin.md)
- Security checklist: [Security, privacy, and compliance (overview)](../start-here/security-privacy-compliance.md)

## Read next

- Integration guide: [How-to: integrate recsys-service into an application](../how-to/integrate-recsys-service.md)
- API reference: [API Reference](api/api-reference.md)
