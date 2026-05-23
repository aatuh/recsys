# Runbook: Service Not Ready

Use this when `/readyz` fails or an orchestrator keeps the API out of rotation.

## Fast triage

Set the service URL:

```bash
BASE_URL=${BASE_URL:-http://localhost:8000}
```

Check liveness separately from readiness:

```bash
curl -fsS "$BASE_URL/healthz"
curl -fsS "$BASE_URL/readyz"
curl -fsS "$BASE_URL/health" || true
```

Interpretation:

| Result | Meaning |
| --- | --- |
| `/healthz` fails | The process or basic liveness path is unhealthy. Start with container logs. |
| `/healthz` passes but `/readyz` fails | The process is up, but a readiness dependency is unhealthy. |
| `/health` fails | Overall health is unhealthy. Treat details as internal operational data. |

## Current readiness dependencies

The service registers:

- `basic` readiness in all modes.
- `database` readiness when a database pool is configured.
- Auth provider health checks when auth middleware registers them.

## Local checks

Inspect logs and database migration state:

```bash
docker compose logs --tail=100 api
cd api && make migrate-status
```

Check common production-sensitive settings:

```bash
rg "AUTH_REQUIRED|JWT_|API_KEY_|DATABASE_URL|RECSYS_ARTIFACT_|PPROF" api/.env api/.env.example
```

## Common causes

| Cause | Signal | Remediation |
| --- | --- | --- |
| Database unreachable | `/readyz` fails after startup and logs mention database ping/connectivity. | Fix `DATABASE_URL`, network policy, DNS, credentials, or database availability. |
| Migrations missing | Service starts but database-backed routes or readiness fail after schema changes. | Run the migration workflow for the environment. |
| JWT/JWKS dependency unhealthy | Auth is enabled and logs mention JWKS or allowed-host failures. | Fix `JWT_JWKS_URL`, `AUTH_JWKS_ALLOWED_HOSTS`, egress, or TLS. |
| Unsafe production config | Service exits during config validation. | Fix missing salts/secrets, S3 SSL settings, or pprof binding as reported by logs. |

## Verification

1. `/healthz` returns success.
2. `/readyz` returns success.
3. A known-good recommendation request succeeds.
4. Logs do not contain repeated readiness or dependency errors after recovery.

## Read next

- [Operations](../../operations.md)
- [Configuration Reference](../../reference/config.md)
- [Security](../../security.md)
