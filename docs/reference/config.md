# Configuration Reference

## Who this is for

Operators, backend developers, and reviewers configuring local, staging, or production RecSys deployments.

## What you will get

- The files that own default configuration.
- Production-sensitive environment variables.
- Local commands for creating env files.
- A compact map of service, pipeline, and evaluation config sources.

## Source files

| Area | Source |
| --- | --- |
| API local defaults | `api/.env.example` |
| API test defaults | `api/.env.test.example` |
| API config loader | `api/internal/config/config.go` |
| Evaluation configs | `recsys-eval/configs/eval/` and `recsys-eval/configs/examples/` |
| Pipeline config | `recsys-pipelines/configs/env/local.json` |
| Docker services | `docker-compose.yml` |

## Local env setup

```bash
make env
make test-env
```

Expected result: `api/.env` and `api/.env.test` are created only if they are missing.

## Production-sensitive service settings

| Variable | Why it matters |
| --- | --- |
| `AUTH_REQUIRED` | Should remain `true` for protected recommendation and admin routes. |
| `JWT_AUTH_ENABLED` / `API_KEY_ENABLED` | Choose production auth mode. |
| `DEV_AUTH_ENABLED` | Development only; disable in production. |
| `AUTH_REQUIRE_TENANT_CLAIM` | Enforces tenant scoping from auth claims when appropriate. |
| `API_KEY_HASH_SECRET` | Required in production when API key auth is enabled. |
| `EXPOSURE_HASH_SALT` | Required in production when exposure logging is enabled. |
| `EXPERIMENT_ASSIGNMENT_SALT` | Required in production when experiment assignment is enabled. |
| `EXPERIMENT_CONFIG_JSON` | Optional JSON lifecycle config for experiment traffic allocation, variants, and active windows. |
| `CORS_ALLOWED_ORIGINS` | Restrict browser origins for web clients and Swagger UI. |
| `RECSYS_ARTIFACT_S3_USE_SSL` | Must be true in production when S3 artifact mode is configured. |
| `PPROF_ENABLED` | Only allowed on loopback bindings. |

## Artifact mode

Artifact mode reads a manifest and artifact blobs instead of relying only on in-memory/default data:

```bash
RECSYS_ARTIFACT_MODE_ENABLED=true
RECSYS_ARTIFACT_MANIFEST_TEMPLATE=s3://recsys-artifacts/registry/current/{tenant}/{surface}/manifest.json
RECSYS_ARTIFACT_MANIFEST_TTL=1m
RECSYS_ARTIFACT_CACHE_TTL=1m
```

Use a rollback-ready manifest process before enabling artifact mode for production traffic.

## Pipeline artifacts

`recsys-pipelines` accepts an `artifact_kinds` array in its JSON config:

```json
{
  "artifact_kinds": ["popularity", "cooc", "implicit", "content_sim", "session_seq"],
  "catalog": {
    "path": "../examples/data/ecommerce-mini/catalog.csv",
    "format": "csv"
  }
}
```

Supported artifact kinds are `popularity`, `cooc`, `implicit`, `content_sim`, and `session_seq`. If `artifact_kinds` is
omitted, the default remains `popularity` plus `cooc`. `content_sim` requires `catalog.path`; `catalog.format` can be
`csv` or `jsonl`, and is inferred from the file extension when omitted.

## Experiment lifecycle

`EXPERIMENT_CONFIG_JSON` can restrict deterministic assignment by experiment ID, surface, traffic percentage, and active
window:

```json
[
  {
    "id": "home-ranker-v2",
    "enabled": true,
    "surface": "home",
    "traffic_percent": 25,
    "variants": ["A", "B"],
    "starts_at": "2026-01-01T00:00:00Z",
    "ends_at": "2026-02-01T00:00:00Z"
  }
]
```

If no matching definition exists, the assignment behavior remains backward compatible and uses
`EXPERIMENT_DEFAULT_VARIANTS`.

## Configuration validation

The API fails fast on unsafe production combinations, including missing production salts/secrets for enabled features,
unsafe S3 SSL settings in production artifact mode, and pprof on non-loopback bindings.
