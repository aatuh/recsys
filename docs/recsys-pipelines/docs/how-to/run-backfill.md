---
diataxis: how-to
tags:
  - recsys-pipelines
---
# How-to: Run a backfill safely

Backfills re-run the pipeline for a date range. This repo uses **daily UTC
windows** and includes a maximum range limit.

## Safe approach

1) Start small: one day.
1) Expand gradually.
1) Monitor validation and resource limits.

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

## Read next

- Backfill safely: [How-to: Backfill pipelines safely](backfill-safely.md)
- Windows and backfills (concepts): [Windows and backfills](../explanation/windows-and-backfills.md)
- Validation and guardrails: [Validation and guardrails](../explanation/validation-and-guardrails.md)
- Output layout: [Output layout (local filesystem)](../reference/output-layout.md)
- Limit exceeded runbook: [Runbook: Limit exceeded](../operations/runbooks/limit-exceeded.md)
