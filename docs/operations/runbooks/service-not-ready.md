---
diataxis: how-to
tags:
  - runbook
  - ops
  - recsys-service
---
# Runbook: Service not ready
This guide shows how to runbook: Service not ready in a reliable, repeatable way.


## Symptoms

- `/readyz` returns `503 Service Unavailable`
- Kubernetes marks the pod `Ready=False` (readiness probe failing)
- `/healthz` may still return `200 OK` (process is running but not ready)

## Quick triage (copy/paste)

Set:

```bash
BASE_URL=${BASE_URL:-http://localhost:8000}
```

Then run:

```bash
curl -fsS "$BASE_URL/healthz"
curl -fsS "$BASE_URL/readyz" || true
curl -fsS "$BASE_URL/health/detailed"
```

Interpretation:

- `healthz` failing means the process is not healthy (check container logs first).
- `readyz` failing means at least one **readiness dependency** is unhealthy.
- `health/detailed` tells you *which* check(s) are unhealthy and why.

!!! warning
    When sharing `/health/detailed` output (tickets, Slack), redact secrets and internal hostnames.

## What readiness checks exist

In `recsys-service`, readiness includes:

- `basic` (always healthy)
- `database` (Postgres ping) when `DATABASE_URL` is configured
- auth provider checks (for example: JWKS fetch) when JWT auth is enabled

## Common causes and safe remediations

### Check `database` is unhealthy

Likely causes:

- Postgres is down or not reachable from the service network
- `DATABASE_URL` points to the wrong host/db or has invalid credentials
- network policy / DNS / TLS issues

Checks:

- Local dev (docker compose): `cd api && make migrate-status`
- From a pod/container: run a simple query (for example `psql "$DATABASE_URL" -c 'select 1'`)

Safe remediations:

- Fix connectivity/credentials (secret wiring, DNS, firewall, DB availability)
- Apply migrations if they are behind:
  - Local dev: `cd api && make migrate-up`
  - Production: run your migration job (see [Database migrations](../../reference/database/migrations.md))
  - If migrations fail: see [Runbook: Database migration issues](db-migration-issues.md)

### Auth/JWKS check is unhealthy (JWT enabled)

Likely causes:

- `JWT_JWKS_URL` is unreachable from the pod (egress/DNS)
- the JWKS host is not allowlisted (`AUTH_JWKS_ALLOWED_HOSTS`)
- TLS/certificate problems

Checks:

- Verify configuration: `JWT_JWKS_URL`, `AUTH_JWKS_ALLOWED_HOSTS`
- From the same network as the service, check reachability:
  - `curl -fsS "$JWT_JWKS_URL" >/dev/null`

Safe remediations:

- Fix egress/DNS/TLS for the JWKS URL
- Update `AUTH_JWKS_ALLOWED_HOSTS` to match the JWKS hostname

!!! warning
    Do not use insecure JWKS settings (`AUTH_ALLOW_INSECURE_JWKS=true`) in production.

## If `/readyz` is OK but traffic still fails

Readiness only covers baseline dependencies. If `/readyz` is `200` but:

- `/v1/recommend` returns empty → see [Runbook: Empty recs](empty-recs.md)
- you are in artifact/manifest mode and requests fail → verify `RECSYS_ARTIFACT_*` config and manifest publishing (see
  [Data modes](../../explanation/data-modes.md))

## Read next

- Failure modes & diagnostics: [Failure modes & diagnostics](../failure-modes.md)
- Database migration issues: [Runbook: Database migration issues](db-migration-issues.md)
- Operations index: [Operations](../index.md)
