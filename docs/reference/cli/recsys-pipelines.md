---
diataxis: reference
tags:
  - reference
  - cli
  - recsys-pipelines
  - developer
  - ops
---
# CLI: recsys-pipelines
This page is the canonical reference for CLI: recsys-pipelines.


## Who this is for

- Engineers running `recsys-pipelines` locally or in CI
- Operators publishing artifacts + manifests for `recsys-service` (artifact mode)

## What you will get

- The canonical `recsys-pipelines` command, flags, and exit codes
- Copy/paste examples for local runs and incremental backfills

## Build/install

From repo root:

```bash
(cd recsys-pipelines && make build)
```

Binary:

- `recsys-pipelines/bin/recsys-pipelines`

## Commands

### `recsys-pipelines run`

Runs the end-to-end pipeline for a date window.

Required flags:

- `--tenant <id>`
- `--surface <name>`
- `--end YYYY-MM-DD`
- `--start YYYY-MM-DD` (required unless `--incremental` is set)

Common optional flags:

- `--config <path.json>`: env config (JSON). Default: `configs/env/local.json`
- `--segment <name>`: optional segment label
- `--incremental`: uses the last checkpointed day as the start

Output (with the default local config):

- Manifest: `recsys-pipelines/.out/registry/current/<tenant>/<surface>/manifest.json`

Notes:

- The CLI currently expects **JSON** config files.
- The first `--incremental` run needs a checkpoint. If no checkpoint exists, pass `--start` once.

### `recsys-pipelines version`

Prints the CLI version.

## Exit codes

- `0`: success
- `1`: pipeline run failed
- `2`: usage error (unknown command, flag parse error, missing/invalid flags, config errors)

## Examples

### Local: run one day

```bash
(cd recsys-pipelines && ./bin/recsys-pipelines run \
  --config configs/env/local.json \
  --tenant demo \
  --surface home \
  --start 2026-01-01 \
  --end 2026-01-01)
```

### CI: incremental daily run

```bash
(cd recsys-pipelines && ./bin/recsys-pipelines run \
  --config configs/env/local.json \
  --tenant demo \
  --surface home \
  --end 2026-01-31 \
  --incremental)
```

## Read next

- Artifact mode workflow: [production-like run (pipelines → object store → ship/rollback)](../../tutorials/production-like-run.md)
- Operate pipelines: [How-to: operate recsys-pipelines](../../how-to/operate-pipelines.md)
- Runbooks (pipeline failed): [Runbook: Pipeline failed](../../recsys-pipelines/docs/operations/runbooks/pipeline-failed.md)
