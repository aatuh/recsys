---
tags:
  - overview
  - developer
  - ops
  - artifacts
---

# Choose your data mode (DB-only vs artifact/manifest)

## Who this is for

- Lead developers choosing a serving mode for a pilot or first integration
- Operators deciding how you will ship and roll back offline signals

## What you will get

- A decision guide (when to choose each mode)
- The exact tutorial to follow next
- A quick “how to confirm it’s active” checklist

## The decision (in 60 seconds)

Start with **DB-only mode** unless you already run offline pipelines + object storage and you need an atomic
ship/rollback lever.

Choose **artifact/manifest mode** when you need to publish versioned signals (popularity/co-vis/embeddings) and ship
or roll back by swapping a manifest pointer.

## Comparison

| Mode | Choose this when | Main tradeoff |
| --- | --- | --- |
| DB-only | Fastest path to first success | Signals live in Postgres; no atomic ship/rollback lever |
| Artifact/manifest | Versioned artifacts + ship/rollback by pointer | More moving parts (pipelines + object store) |

## Start here (exact tutorials)

- DB-only mode:
  - [Quickstart (10 minutes)](../tutorials/quickstart.md)
  - [Local end-to-end](../tutorials/local-end-to-end.md)
- Artifact/manifest mode:
  - [Production-like run](../tutorials/production-like-run.md)

For the minimal runnable stacks by goal, see:

- [Minimum components by goal](minimum-components-by-goal.md)

## Minimum configuration knobs

- DB-only mode (default):
  - `RECSYS_ARTIFACT_MODE_ENABLED=false`
- Artifact/manifest mode:
  - `RECSYS_ARTIFACT_MODE_ENABLED=true`
  - `RECSYS_ARTIFACT_MANIFEST_TEMPLATE=...` (points at your “current manifest” convention)

## How to confirm which mode is active

- Check the running environment:

  ```bash
  docker compose exec -T api sh -c 'printenv | grep -E "^RECSYS_ARTIFACT_MODE_ENABLED="'
  ```

- If artifact/manifest mode is enabled, confirm the service actually loads a manifest:

  ```bash
  docker compose logs --tail 200 api | grep -i "artifact manifest loaded"
  ```

If artifact mode is enabled but the manifest is missing or stale, serving can degrade to empty/partial results. See:

- [Failure modes & diagnostics](../operations/failure-modes.md)
- Runbook: [Stale manifest](../operations/runbooks/stale-manifest.md)

## Read next

- Data modes (details): [Data modes](../explanation/data-modes.md)
- Production readiness checklist:
  [Production readiness checklist](../operations/production-readiness-checklist.md)
- Operate pipelines (artifact mode): [How-to: operate recsys-pipelines](../how-to/operate-pipelines.md)
