# Runbook: Database migration issues

## Symptoms

- Service crashes or fails to start during deploy/upgrade
- Requests fail with schema errors (missing table/column, constraint mismatch)
- Migration job exits non-zero or reports failed migrations

## Decision tree (fast path)

```mermaid
flowchart TD
  A[Upgrade failing] --> B{DB reachable?}
  B -->|No| C[Fix connectivity/creds/network policy]
  B -->|Yes| D{Failed migrations present?}
  D -->|Yes| E[Investigate the failed migration and roll forward]
  D -->|No| F{Migrations behind the code?}
  F -->|Yes| G[Apply migrations (up)]
  F -->|No| H[Check app logs for non-migration causes]
```

## Quick triage (copy/paste)

Local dev:

```bash
cd api
make migrate-preflight
make migrate-status
```

Production:

- Run your migration job/container and capture output.
- If you have DB access, run the same migrate commands in a controlled job environment.

## Safe remediations

- Prefer **roll forward**:
  - fix the migration and deploy a new migration
  - re-run `up`
- Avoid `down` in production unless you have an explicit rollback plan and DB backups.

See the migration policy and commands: [`reference/database/migrations.md`](../../reference/database/migrations.md)

## Verification

- `migrate status` shows no failed migrations.
- Service starts cleanly and `/readyz` becomes `200 OK`.

## Read next

- Service readiness runbook: [`operations/runbooks/service-not-ready.md`](service-not-ready.md)
- Migrations reference: [`reference/database/migrations.md`](../../reference/database/migrations.md)
