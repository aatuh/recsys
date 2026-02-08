---
diataxis: reference
tags:
  - reference
  - config
  - recsys-pipelines
  - developer
  - ops
---
# recsys-pipelines configuration
This page is the canonical reference for recsys-pipelines configuration.


## Who this is for

- Data/ML engineers operating `recsys-pipelines` locally or in CI
- Operators wiring pipelines outputs into `recsys-service` artifact/manifest mode

## What you will get

- The JSON config shape passed via `recsys-pipelines run --config <path>`
- The core knobs you typically set (sources, object store, registry, limits)
- Copy/paste config examples for local filesystem and MinIO/S3

## Reference

The CLI expects a **JSON** config file:

```bash
recsys-pipelines run --config configs/env/local.json --tenant demo --surface home --start 2026-01-01 --end 2026-01-07
```

### Top-level keys

| Key | Meaning |
| --- | --- |
| `out_dir` | Base output directory (local runs). |
| `raw_events_dir` | Input events directory (filesystem mode). |
| `canonical_dir` | Canonical output directory. |
| `checkpoint_dir` | Checkpoint storage for incremental runs. |
| `raw_source` | Raw ingestion source configuration. |
| `artifacts_dir` | Staging directory (job mode and pipeline staging). |
| `object_store_dir` | Where published blobs are written (local fs mode). |
| `object_store` | Object store configuration (`fs`, `s3`, or `minio`). |
| `registry_dir` | Where manifests and records are written (filesystem adapter). |
| `db` | Optional Postgres connection for DB-backed signals. |
| `limits` | Guardrails and caps (artifact sizes, fanout, backfill windows). |

### `raw_source`

`raw_source.type` supports:

- `fs`
- `s3`
- `minio`
- `postgres`
- `kafka`

Note: the Kafka connector is scaffolded and returns a clear error until it is implemented with a streaming consumer.

### Artifacts produced (v1)

- `popularity`
- `cooc`
- `implicit` (collaborative)
- `content_sim`
- `session_seq`

## Examples

### Local filesystem (recommended for first runs)

```json
{
  "out_dir": ".out",
  "raw_events_dir": "testdata/events",
  "canonical_dir": ".out/canonical",
  "checkpoint_dir": ".out/checkpoints",
  "artifacts_dir": ".out/artifacts",
  "object_store_dir": ".out/objectstore",
  "registry_dir": ".out/registry",
  "raw_source": { "type": "fs", "dir": "testdata/events" },
  "object_store": { "type": "fs", "dir": ".out/objectstore" },
  "limits": {
    "max_days_backfill": 30,
    "max_events_per_run": 100000,
    "max_sessions_per_run": 100000,
    "max_items_per_session": 100,
    "max_distinct_items_per_run": 50000,
    "max_neighbors_per_item": 200,
    "max_items_per_artifact": 50000,
    "min_cooc_support": 2,
    "max_users_per_run": 100000,
    "max_items_per_user": 1000
  }
}
```

### MinIO/S3 object store (artifact publish)

```json
{
  "object_store": {
    "type": "minio",
    "s3": {
      "endpoint": "localhost:9000",
      "bucket": "recsys-artifacts",
      "access_key": "minioadmin",
      "secret_key": "minioadmin",
      "prefix": "recsys",
      "use_ssl": false
    }
  }
}
```

## Read next

- CLI usage and exit codes: [CLI: recsys-pipelines](../cli/recsys-pipelines.md)
- Operate pipelines (ship/rollback): [How-to: operate recsys-pipelines](../../how-to/operate-pipelines.md)
- Pipelines module config reference (full field list): [Config reference](../../recsys-pipelines/docs/reference/config.md)
