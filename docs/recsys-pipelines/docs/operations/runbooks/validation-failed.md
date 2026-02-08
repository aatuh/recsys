---
diataxis: how-to
tags:
  - recsys-pipelines
  - runbook
---
# Runbook: Validation failed
This guide shows how to runbook: Validation failed in a reliable, repeatable way.


## Symptoms

- Pipeline stops before publish
- Error indicates validation failure

## Common causes

- Events outside the window (timestamp issues)
- Bad JSON / schema mismatch
- Unexpected spike/drop in event volume

## Recovery

- Fix data at source if possible
- Re-run the affected day
- If needed, roll back serving to previous artifacts

See `explanation/data-lifecycle.md`.

## Read next

- Data lifecycle: [Data lifecycle](../../explanation/data-lifecycle.md)
- Run a backfill: [How-to: Run a backfill safely](../../how-to/run-backfill.md)
- Limit exceeded runbook: [Runbook: Limit exceeded](limit-exceeded.md)
- SLOs and freshness: [SLOs and freshness](../slos-and-freshness.md)
