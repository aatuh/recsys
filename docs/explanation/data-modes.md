---
tags:
  - explanation
  - artifacts
  - ops
  - developer
---

# Data modes: DB-only vs artifact/manifest

## Who this is for

- Developers choosing the simplest mode for a pilot
- Operators deciding how to ship/rollback signals safely

## What you will get

- A clear comparison of DB-only vs artifact/manifest mode
- What you need to run in each mode (at a high level)
- The common failure modes when a mode is misconfigured

## Overview

The service supports **DB-only mode** and **artifact/manifest mode**:

- **DB-only mode**: read signals directly from Postgres tables.
- **Artifact/manifest mode**: read versioned signal blobs from object storage via a manifest pointer.

DB-only is the code default. The local Docker Compose environment enables artifact mode by default to support the demo
pipeline flow.

## DB-only mode (current, recommended for MVP)

Signals are stored directly in Postgres tables and read by the service:

- `item_tags`
- `item_popularity_daily`
- `item_covisit_daily` (if enabled)

Popularity uses a decayed sum over `item_popularity_daily` with the configured
half-life, so **newer days dominate** when you seed both recent and older rows.

This is ideal for local development and popularity-only pilots.

## Artifact/manifest mode (pipelines + object store)

Pipelines can publish artifacts (popularity, co-vis, embeddings) to object
storage and update a manifest pointer. This enables atomic updates and easy
rollback, but the **service must be configured to read artifacts**.

Enable artifact mode:

- `RECSYS_ARTIFACT_MODE_ENABLED=true`
- `RECSYS_ARTIFACT_MANIFEST_TEMPLATE` (for example:
  `s3://recsys/registry/current/{tenant}/{surface}/manifest.json` or
  `file:///data/registry/current/{tenant}/{surface}/manifest.json`)

Notes:

- `{tenant}` uses the incoming tenant id (header/JWT) when available.
- `{surface}` maps to the request surface (namespace).
- Tags and constraints still read from Postgres (`item_tags`), even in artifact mode.

!!! warning
    If artifact mode is enabled but the manifest pointer is missing/stale (or the object store is unreachable), serving
    can degrade to empty/partial results. See: [`operations/failure-modes.md`](../operations/failure-modes.md).

## Recommendation

- Use **DB-only** for MVP and local testing (default today).
- Use **artifact/manifest** for production-scale artifacts once pipelines are producing artifacts and the service is
  configured to read them.

## Which mode is active?

The service runs in **DB-only mode by default**. When
`RECSYS_ARTIFACT_MODE_ENABLED=true`, the service reads popularity/co-visitation
from the artifact manifest and uses Postgres for tag metadata.

## Read next

- Minimum components by goal: [`start-here/minimum-components-by-goal.md`](../start-here/minimum-components-by-goal.md)
- Tutorial (DB-only loop): [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)
- Tutorial (artifact/manifest mode): [`tutorials/production-like-run.md`](../tutorials/production-like-run.md)
