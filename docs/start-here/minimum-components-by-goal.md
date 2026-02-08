---
diataxis: reference
tags:
  - overview
  - quickstart
  - developer
  - ops
---
# Minimum components by goal
This page is the canonical reference for Minimum components by goal.


## Who this is for

- Developers and operators choosing the smallest runnable setup for a pilot or integration.

## What you will get

- a decision table: goal → what to run → what stores you need
- links to the shortest tutorial for each setup

## Decision table

| Goal | What you run | Required stores | Start here |
| --- | --- | --- | --- |
| Serve (DB-only) | `recsys-service` + PG | PG | [`Quickstart`](../tutorials/quickstart.md) |
| Serve + eval report | + `recsys-eval` | PG + logs | [`Local end-to-end`](../tutorials/local-end-to-end.md) |
| Artifact mode | + S3 + `recsys-pipelines` | PG + S3 | [`Prod-like run`](../tutorials/production-like-run.md) |

Notes:

- `PG` = Postgres.
- `S3` = an S3-compatible bucket (MinIO works for local dev).
- `recsys-algo` is the deterministic ranking core used by `recsys-service` (you don’t run it separately).
- Artifact/manifest mode still uses Postgres for tenants/config/rules and tag metadata.

## Read next

- Auth & tenancy: [Auth and tenancy reference](../reference/auth-and-tenancy.md)
- Data modes: [Data modes: DB-only vs artifact/manifest](../explanation/data-modes.md)
- Integrate into an app: [How-to: integrate recsys-service into an application](../how-to/integrate-recsys-service.md)
