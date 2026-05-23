# Architecture

RecSys is organized as a suite rather than one monolithic service. The serving path is online and low-latency; the
pipeline and evaluation paths are offline and auditable.

## Components

| Component | Path | Role |
| --- | --- | --- |
| recsys-service | `api/` | HTTP API, admin routes, auth/tenancy, migrations, exposure logging, artifact loading, and OpenAPI artifacts. |
| recsys-algo | `recsys-algo/` | Pure ranking logic: candidates, scoring, personalization, rules, diversity, and deterministic options. |
| recsys-pipelines | `recsys-pipelines/` | Batch jobs for ingesting events, computing signals, validating quality, and publishing manifests/artifacts. |
| recsys-eval | `recsys-eval/` | Evaluation CLI and schemas for offline regression, experiment analysis, OPE, interleaving, and decisions. |

## Runtime flow

```text
client
  -> recsys-service /v1/recommend
  -> tenant config + rules + optional artifact manifest
  -> recsys-algo ranking
  -> recommendation response + metadata
  -> exposure log for evaluation and audit
```

The service reads tenant scope from auth claims or configured tenant headers. Admin routes manage tenant config and
rules. Exposure logging is optional, but production evaluation needs stable request IDs and exposure/outcome joins.

## Offline flow

```text
raw events/catalog
  -> recsys-pipelines jobs
  -> versioned artifacts + manifest
  -> recsys-service artifact mode
  -> exposure/outcome logs
  -> recsys-eval reports and ship/hold decisions
```

The artifact manifest is the rollout boundary. A published manifest can be advanced, rolled back, or held while the
service keeps serving the last known-good artifact version.

## Dependency direction

The repository follows a ports/adapters shape:

- Domain and ranking logic stays inside `recsys-algo`, `recsys-eval/internal/domain`, and
  `recsys-pipelines/internal/domain`.
- Application use cases depend on ports, not concrete infrastructure.
- Adapters implement HTTP, database, filesystem, S3, JSONL, Postgres, reporting, and clock boundaries.
- The API service wires adapters at the edge in `api/cmd/recsys-service/`.

## Trust boundaries

| Boundary | Guardrail |
| --- | --- |
| HTTP input | Validation, problem responses, rate limits, auth middleware, tenant checks. |
| Tenant data | Tenant header/claim enforcement and optional database RLS requirement. |
| Files and artifacts | Size limits, path confinement, manifest TTLs, and optional S3 SSL enforcement in production. |
| Explanations and traces | `RECSYS_EXPLAIN_REQUIRE_ADMIN=true` by default. |
| Production secrets | Required salts/secrets are enforced by config validation when production toggles are enabled. |

## What this documentation does not claim

- It does not claim managed hosting.
- It does not claim automatic model training or KPI lift.
- It does not claim legal compliance beyond the documented technical posture and commercial terms.
- It does not replace a production readiness review for a specific deployment.
