---
diataxis: reference
tags:
  - recsys-pipelines
---
# Config reference

Config is JSON. Example: `configs/env/local.json`.

Top-level fields:

- `out_dir`: base output directory (local runs)
- `raw_events_dir`: input events directory
- `canonical_dir`: canonical output directory
- `checkpoint_dir`: checkpoint storage for incremental runs
- `raw_source`: raw ingestion source configuration
- `artifacts_dir`: staging directory (job mode and pipeline staging)
- `object_store_dir`: where published blobs are written (local fs mode)
- `object_store`: object store configuration (fs or s3/minio)
- `registry_dir`: where manifests and records are written
- `db`: optional Postgres connection for DB-backed signals

## object_store

```json
{
  "type": "fs | s3 | minio",
  "dir": ".out/objectstore",
  "s3": {
    "endpoint": "localhost:9000",
    "bucket": "recsys-artifacts",
    "access_key": "minioadmin",
    "secret_key": "minioadmin",
    "prefix": "recsys",
    "use_ssl": false
  }
}
```

## db

```json
{
  "dsn": "postgres://user:pass@localhost:5432/db?sslmode=disable",
  "auto_create_tenant": true,
  "statement_timeout_s": 5
}
```

## limits

- `max_days_backfill`
- `max_events_per_run`
- `max_sessions_per_run`
- `max_items_per_session`
- `max_distinct_items_per_run`
- `max_neighbors_per_item`
- `max_items_per_artifact`
- `min_cooc_support`
- `max_users_per_run`
- `max_items_per_user`

See `explanation/validation-and-guardrails.md`.

## raw_source

```json
{
  "type": "fs | s3 | minio | postgres | kafka",
  "dir": "testdata/events",
  "s3": {
    "endpoint": "localhost:9000",
    "bucket": "recsys-raw",
    "access_key": "minioadmin",
    "secret_key": "minioadmin",
    "prefix": "raw/events",
    "use_ssl": false
  },
  "postgres": {
    "dsn": "postgres://user:pass@localhost:5432/db?sslmode=disable",
    "tenant_table": "tenants",
    "exposure_table": "exposure_events"
  },
  "kafka": {
    "brokers": ["localhost:9092"],
    "topic": "recsys-exposures",
    "group_id": "recsys-pipelines"
  }
}
```

Note: the Kafka connector is scaffolded and returns a clear error until it is
implemented with a streaming consumer.

## Read next

- Start here: [Start here](../start-here.md)
- Validation and guardrails: [Validation and guardrails](../explanation/validation-and-guardrails.md)
- Run incremental: [How-to: Run incremental pipelines](../how-to/run-incremental.md)
- Limit exceeded runbook: [Runbook: Limit exceeded](../operations/runbooks/limit-exceeded.md)
