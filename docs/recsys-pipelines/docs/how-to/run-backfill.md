
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

- Backfill safely: [`how-to/backfill-safely.md`](backfill-safely.md)
- Windows and backfills (concepts): [`explanation/windows-and-backfills.md`](../explanation/windows-and-backfills.md)
- Validation and guardrails: [`explanation/validation-and-guardrails.md`](../explanation/validation-and-guardrails.md)
- Output layout: [`reference/output-layout.md`](../reference/output-layout.md)
- Limit exceeded runbook: [`operations/runbooks/limit-exceeded.md`](../operations/runbooks/limit-exceeded.md)
