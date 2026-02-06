---
tags:
  - reference
  - config
  - recsys-service
  - developer
  - ops
---

# recsys-service configuration

## Who this is for

- Operators configuring `recsys-service` (Docker/Helm/Kubernetes)
- Developers who need the canonical env var names and defaults

## What you will get

- The canonical environment variables (names, defaults, and meanings)
- Copy/paste examples for local dev and production-like deployments
- Notes about common misconfigurations (auth, tenancy, exposure logs, artifact mode)

## Reference

### Core (server + database)

| Env var | Default | Meaning |
| --- | --- | --- |
| `API_ADDR` | `:8000` | Bind address for the HTTP server. |
| `ENV` | `development` | Runtime environment. `prod` and `production` are treated as production. |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error`. |
| `DATABASE_URL` | required | Postgres DSN. |
| `DATABASE_URL_FILE` | unset | If set, reads DSN from file and sets `DATABASE_URL`. |
| `MIGRATE_ON_START` | `false` | Run DB migrations on startup. |
| `MIGRATIONS_DIR` | `-` | Migration source. `-` means use embedded migrations. |

### Docs endpoints (local helper)

| Env var | Default | Meaning |
| --- | --- | --- |
| `DOCS_ENABLED` | `true` (non-prod) | Enables Swagger UI/docs endpoints (default is `false` in prod). |
| `FRONTEND_BASE_URL` | `http://localhost:3000` | Base URL for external links shown by the service (when applicable). |

### Auth + tenancy

| Env var | Default | Meaning |
| --- | --- | --- |
| `AUTH_REQUIRED` | `true` | Require auth for protected routes. |
| `AUTH_REQUIRE_TENANT_CLAIM` | `true` in prod | When `true`, tenant must come from a JWT claim (unless API keys). |
| `TENANT_HEADER_NAME` | `X-Org-Id` | Tenant header name (when tenant is not derived from auth). |
| `AUTH_TENANT_CLAIMS` | `tenant_id,org_id,orgId,organization_id` | CSV claim keys to search for tenant id (JWT). |
| `AUTH_ROLE_CLAIMS` | `roles,role,scope,scopes` | CSV claim keys to search for roles (JWT). |
| `AUTH_VIEWER_ROLE` | `viewer` | Admin-read role name. |
| `AUTH_OPERATOR_ROLE` | `operator` | Admin-write role name. |
| `AUTH_ADMIN_ROLE` | `admin` | Full admin role name. Set empty to disable admin role. |
| `AUTH_JWKS_ALLOWED_HOSTS` | unset | CSV allow-list of JWKS hosts (when `AUTH_ALLOW_INSECURE_JWKS=false`). |
| `AUTH_ALLOW_INSECURE_JWKS` | `false` | Allow JWKS URLs that are not in `AUTH_JWKS_ALLOWED_HOSTS`. |

Notes:

- Claim keys can be dotted (for example: `realm_access.roles`).
- When multiple tenant sources are present, they must match (see: [`reference/auth-and-tenancy.md`](../auth-and-tenancy.md)).

### JWT auth (production)

| Env var | Default | Meaning |
| --- | --- | --- |
| `JWT_AUTH_ENABLED` | `true` | Enable JWT middleware. |
| `JWT_JWKS_URL` | unset | JWKS URL for key discovery. |
| `JWT_ISSUER` | unset | Expected issuer (`iss`). |
| `JWT_AUDIENCE` | unset | Expected audience (`aud`). |
| `JWT_ALLOWED_ALGORITHMS` | unset | CSV allow-list of JWT algorithms. |
| `JWT_JWKS_REFRESH_INTERVAL` | `10m` | How often JWKS is refreshed. |
| `JWT_JWKS_REFRESH_TIMEOUT` | `5s` | Timeout for JWKS refresh. |
| `JWT_ALLOWED_CLOCK_SKEW` | `30s` | Allowed clock skew for token checks. |
| `JWT_ALLOW_DANGEROUS_DEV_BYPASSES` | `false` | Allows skip header and other dev bypasses (dangerous). |
| `JWT_SKIP_HEADER_ENABLED` | `false` | Enable a "skip auth" header for trusted proxies (dangerous). |
| `JWT_SKIP_HEADER_NAME` | unset | Header name for the skip mechanism. |
| `JWT_SKIP_TRUSTED_PROXIES` | unset | CSV trusted proxy CIDRs/hosts for skip header. |

### API keys

| Env var | Default | Meaning |
| --- | --- | --- |
| `API_KEY_ENABLED` | `false` | Enable API key auth. |
| `API_KEY_HEADER` | `X-API-Key` | API key header name. |
| `API_KEY_HASH_SECRET` | unset | Secret used to hash keys for comparison. |

### Dev header auth (local/test)

| Env var | Default | Meaning |
| --- | --- | --- |
| `DEV_AUTH_ENABLED` | `false` | Enable dev headers middleware. |
| `DEV_AUTH_USER_ID_HEADER` | `X-Dev-User-Id` | User id header. |
| `DEV_AUTH_EMAIL_HEADER` | `X-Dev-User-Email` | Email header. |
| `DEV_AUTH_FIRST_NAME_HEADER` | `X-Dev-User-First` | First name header. |
| `DEV_AUTH_LAST_NAME_HEADER` | `X-Dev-User-Last` | Last name header. |
| `DEV_AUTH_DEFAULT_LANGUAGE` | unset | Default language value when header is missing. |
| `DEV_AUTH_TENANT_HEADER` | `X-Dev-Org-Id` | Dev tenant header (used to set tenant scope in local flows). |

### CORS (Swagger UI and browser clients)

| Env var | Default | Meaning |
| --- | --- | --- |
| `CORS_ALLOWED_ORIGINS` | `*` | CSV origins. |
| `CORS_ALLOWED_METHODS` | `GET,POST,PUT,DELETE,OPTIONS` | CSV HTTP methods. |
| `CORS_ALLOWED_HEADERS` | `Accept,Authorization,Content-Type` | CSV headers. The service appends auth headers. |
| `CORS_ALLOW_CREDENTIALS` | `false` | Sets `Access-Control-Allow-Credentials`. |
| `CORS_MAX_AGE` | `300` | Preflight cache max-age (seconds). |

### Per-tenant rate limiting

| Env var | Default | Meaning |
| --- | --- | --- |
| `TENANT_RATE_LIMIT_ENABLED` | `true` | Enable tenant rate limiting. |
| `TENANT_RATE_LIMIT_CAPACITY` | `60` | Token bucket capacity. |
| `TENANT_RATE_LIMIT_REFILL_RATE` | `30` | Tokens refilled per second. |
| `TENANT_RATE_LIMIT_RETRY_AFTER` | `1s` | Retry-after duration when limited. |
| `TENANT_RATE_LIMIT_FAIL_OPEN` | `false` | When `true`, allow traffic if the limiter fails. |
| `RATE_LIMIT_SKIP_ENABLED` | `false` | Base service bypass toggle (intended for test/dev). |
| `RATE_LIMIT_SKIP_HEADER` | unset | Header used for bypass, when enabled. |
| `RATE_LIMIT_ALLOW_DANGEROUS_DEV_BYPASSES` | `false` | Allows bypass headers in non-prod (dangerous). |

### Audit logging (admin actions)

| Env var | Default | Meaning |
| --- | --- | --- |
| `AUDIT_LOG_ENABLED` | `true` if path set | Enables audit JSONL logging. |
| `AUDIT_LOG_PATH` | unset | JSONL output path. |
| `AUDIT_LOG_FSYNC` | `false` | When `true`, fsync after each write. |

### Exposure logging (evaluation input)

| Env var | Default | Meaning |
| --- | --- | --- |
| `EXPOSURE_LOG_ENABLED` | `true` if path set | Enables exposure JSONL logging. |
| `EXPOSURE_LOG_PATH` | unset | JSONL output path. |
| `EXPOSURE_LOG_FORMAT` | `service_v1` | `service_v1` (operational) or `eval_v1` (compatible with `recsys-eval`). |
| `EXPOSURE_LOG_FSYNC` | `false` | When `true`, fsync after each write. |
| `EXPOSURE_LOG_RETENTION_DAYS` | `30` | Retention (days) for cleanup jobs, when applicable. |
| `EXPOSURE_HASH_SALT` | unset | Salt for hashing/pseudonymization (also used as default experiment salt). |

See also: [Minimum instrumentation](../minimum-instrumentation.md).

### Artifact/manifest mode (pipelines)

| Env var | Default | Meaning |
| --- | --- | --- |
| `RECSYS_ARTIFACT_MODE_ENABLED` | `false` | Enables artifact/manifest reading for popularity/co-vis artifacts. |
| `RECSYS_ARTIFACT_MANIFEST_TEMPLATE` | unset | Manifest URI template (supports `{tenant}` and `{surface}`). |
| `RECSYS_ARTIFACT_MANIFEST_TTL` | `1m` | Manifest cache TTL. |
| `RECSYS_ARTIFACT_CACHE_TTL` | `1m` | Artifact blob cache TTL. |
| `RECSYS_ARTIFACT_MAX_BYTES` | `10000000` | Max artifact size (bytes). |
| `RECSYS_ARTIFACT_S3_ENDPOINT` | unset | S3/MinIO endpoint (`host:port`). |
| `RECSYS_ARTIFACT_S3_ACCESS_KEY` | unset | S3 access key. |
| `RECSYS_ARTIFACT_S3_SECRET_KEY` | unset | S3 secret key. |
| `RECSYS_ARTIFACT_S3_REGION` | unset | S3 region. |
| `RECSYS_ARTIFACT_S3_USE_SSL` | `false` | Use TLS when talking to S3/MinIO. |

Note: Tags and constraints are still read from Postgres even in artifact mode.

### Experiment assignment

| Env var | Default | Meaning |
| --- | --- | --- |
| `EXPERIMENT_ASSIGNMENT_ENABLED` | `true` | Enable deterministic assignment for experiment mode. |
| `EXPERIMENT_DEFAULT_VARIANTS` | `A,B` | CSV variant labels. |
| `EXPERIMENT_ASSIGNMENT_SALT` | `EXPOSURE_HASH_SALT` | Salt used for assignment hashing. |

### Explain/trace safeguards

| Env var | Default | Meaning |
| --- | --- | --- |
| `RECSYS_EXPLAIN_MAX_ITEMS` | `50` | Maximum items allowed in explain output. |
| `RECSYS_EXPLAIN_REQUIRE_ADMIN` | `true` | When `true`, require admin/operator scope for explain. |

### Licensing status (optional)

| Env var | Default | Meaning |
| --- | --- | --- |
| `RECSYS_LICENSE_FILE` | unset | License file path. |
| `RECSYS_LICENSE_PUBLIC_KEY` | unset | License public key (inline). |
| `RECSYS_LICENSE_PUBLIC_KEY_FILE` | unset | License public key (file path). |
| `RECSYS_LICENSE_CACHE_TTL` | `1m` | Cache TTL for license status. |

### recsys-algo defaults (ranking knobs)

These env vars set the default ranking behavior. See also:
[`recsys-algo/ranking-reference.md`](../../recsys-algo/ranking-reference.md).

| Env var | Default | Meaning |
| --- | --- | --- |
| `RECSYS_ALGO_VERSION` | `recsys-algo@local` | Algo version label shown in responses. |
| `RECSYS_ALGO_DEFAULT_NAMESPACE` | `default` | Default namespace/surface when missing. |
| `RECSYS_ALGO_MODE` | `blend` | `blend`, `popularity`, `cooc`, `implicit`, `content_sim`, `session_seq`. |
| `RECSYS_ALGO_RULES_ENABLED` | `false` | Enable pin/exclude rules. |
| `RECSYS_ALGO_RULES_REFRESH_INTERVAL` | `2s` | How often rules are refreshed. |
| `RECSYS_ALGO_RULES_MAX_PINS` | `3` | Max pinned items allowed per request. |
| `RECSYS_ALGO_HALF_LIFE_DAYS` | `30` | Popularity decay half-life. |
| `RECSYS_ALGO_COVIS_WINDOW_DAYS` | `30` | Co-vis window. |
| `RECSYS_ALGO_PROFILE_WINDOW_DAYS` | `30` | Profile signal window. |
| `RECSYS_ALGO_MMR_LAMBDA` | `0` | Diversity lambda (0 disables). |
| `RECSYS_ALGO_MAX_K` | `200` | Max `k` accepted in requests. |

### Observability (OpenTelemetry)

| Env var | Default | Meaning |
| --- | --- | --- |
| `OTEL_TRACING_ENABLED` | `false` | Enable OpenTelemetry tracing export. |
| `OTEL_SERVICE_NAME` | `recsys-service` | Service name (defaults to `api` but rewritten to `recsys-service`). |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | unset | OTLP endpoint. |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | unset | `grpc` or `http/protobuf` (depends on collector). |
| `OTEL_TRACES_SAMPLER` | unset | Overrides sampling strategy (see OTel docs). |
| `OTEL_TRACES_SAMPLER_ARG` | unset | Sampler arg. |
| `OTEL_SAMPLE_RATIO` | unset | Ratio `0..1` when sampler is not set. |
| `APP_ENV` | `ENV` | Optional environment label in traces (falls back to `ENV`). |

### Performance and caches

| Env var | Default | Meaning |
| --- | --- | --- |
| `RECSYS_BACKPRESSURE_MAX_INFLIGHT` | `0` | Max in-flight requests (0 disables). |
| `RECSYS_BACKPRESSURE_MAX_QUEUE` | `0` | Max queued requests (0 disables). |
| `RECSYS_BACKPRESSURE_WAIT_TIMEOUT` | `200ms` | Time to wait for queue capacity. |
| `RECSYS_BACKPRESSURE_RETRY_AFTER` | `1s` | Retry-after when queue is full. |
| `RECSYS_CONFIG_CACHE_TTL` | `5m` | Tenant config cache TTL. |
| `RECSYS_RULES_CACHE_TTL` | `5m` | Tenant rules cache TTL. |
| `PPROF_ENABLED` | `false` | Enables pprof endpoints (dangerous on public networks). |

## Examples

### Local dev (DB-only + dev headers)

```bash
# Required
DATABASE_URL=postgres://recsys-db:recsys@db:5432/recsys-db

# Local dev auth: dev headers
AUTH_REQUIRED=true
DEV_AUTH_ENABLED=true
AUTH_REQUIRE_TENANT_CLAIM=false
TENANT_HEADER_NAME=X-Org-Id
DEV_AUTH_TENANT_HEADER=X-Dev-Org-Id

# DB-only mode (no artifact manifest)
RECSYS_ARTIFACT_MODE_ENABLED=false

# Enable eval-compatible exposure logs
EXPOSURE_LOG_ENABLED=true
EXPOSURE_LOG_FORMAT=eval_v1
EXPOSURE_LOG_PATH=/app/tmp/exposures.eval.jsonl
```

### Production-like (JWT + strict tenant claim)

```bash
# Required
DATABASE_URL=postgres://user:pass@postgres:5432/recsys?sslmode=require
ENV=production

# Auth
AUTH_REQUIRED=true
AUTH_REQUIRE_TENANT_CLAIM=true
JWT_AUTH_ENABLED=true
JWT_JWKS_URL=https://auth.example.com/.well-known/jwks.json
JWT_ISSUER=https://auth.example.com/
JWT_AUDIENCE=recsys
```

## Read next

- Auth and tenancy rules: [`reference/auth-and-tenancy.md`](../auth-and-tenancy.md)
- Artifact vs DB-only modes: [`explanation/data-modes.md`](../../explanation/data-modes.md)
- Ranking knobs: [`recsys-algo/ranking-reference.md`](../../recsys-algo/ranking-reference.md)
