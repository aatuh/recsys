# API Reference

## Who this is for

Developers integrating `recsys-service`, API reviewers, and maintainers checking the OpenAPI/codegen flow.

## What you will get

- The canonical API source path.
- The serving and admin endpoint groups.
- Minimal request examples for local development.
- The codegen command that syncs service API artifacts.

## Canonical OpenAPI source

The repository keeps the canonical OpenAPI YAML at:

```text
docs/reference/api/openapi.yaml
```

`api/Makefile` reads that file through `OPENAPI_SOURCE` and writes generated service artifacts into `api/docs/`.

```bash
make codegen
```

Expected result: `api/docs/openapi.yaml`, `api/docs/openapi.json`, and generated Go spec artifacts are synchronized.

## Public serving endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/v1/recommend` | Return ranked recommendations. |
| `POST` | `/v1/recommend/validate` | Validate and normalize a recommendation request. |
| `POST` | `/v1/similar` | Return similar items for an anchor item. |
| `GET` | `/v1/license` | Return commercial license status without requiring auth. |
| `GET` | `/healthz` | Liveness probe. |
| `GET` | `/readyz` | Readiness probe. |

## Admin endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` / `PUT` | `/v1/admin/tenants/{tenant_id}/config` | Read or replace tenant config. |
| `GET` / `PUT` | `/v1/admin/tenants/{tenant_id}/rules` | Read or replace tenant rules. |
| `POST` | `/v1/admin/tenants/{tenant_id}/cache/invalidate` | Invalidate config, rules, or artifact caches. |
| `GET` | `/v1/admin/tenants/{tenant_id}/audit` | Read admin audit entries. |

Admin routes require the configured admin role when auth is enabled.

## Local recommendation example

```bash
curl -sS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Org-Id: demo' \
  -H 'X-Dev-User-Id: local-dev' \
  -H 'X-Dev-Org-Id: demo' \
  -d '{"surface":"home","k":5}'
```

Expected result: a `200` response with `items` and `meta`, or a user-safe problem response for invalid input or
authorization failures.

## Payload owners

| Payload | Source |
| --- | --- |
| Recommendation request/response | `api/src/specs/types/recsys.go` |
| Admin config/rules/audit responses | `api/src/specs/types/admin.go` |
| License status response | `api/src/specs/types/license.go` |
| Problem responses | `api/src/specs/types/problem.go` |

Do not manually duplicate every field in this page. Update the OpenAPI source and generated artifacts instead.
