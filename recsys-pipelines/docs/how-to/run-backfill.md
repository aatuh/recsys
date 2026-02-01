
# How-to: Run a backfill safely

Backfills re-run the pipeline for a date range. This repo uses **daily UTC
windows** and includes a maximum range limit.

## Safe approach

1) Start small: one day.
2) Expand gradually.
3) Monitor validation and resource limits.

## Using the all-in-one CLI

```bash
./bin/recsys-pipelines run \
  --config configs/env/local.json \
  --tenant demo \
  --surface home \
  --start 2026-01-01 \
  --end 2026-01-07
```

## Using job mode

Run jobs for the same date range and publish last.

## Notes

- End date is inclusive.
- Canonicalization is idempotent per day partition.
- Publishing updates the manifest pointer last.
