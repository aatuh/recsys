---
diataxis: how-to
tags:
  - recsys-pipelines
  - runbook
---
# Runbook: Stale artifacts
This guide shows how to runbook: Stale artifacts in a reliable, repeatable way.


## Symptoms

- Manifest `updated_at` is older than expected
- Serving still uses old URIs

## Triage

1) Confirm scheduler ran
1) Confirm pipeline completed successfully
1) Check for validation failures or limit exceeded

## Recovery

- Re-run the missing day
- If inputs are missing, decide whether to publish empty artifacts or skip

See `how-to/run-backfill.md`.

## Read next

- SLOs and freshness: [SLOs and freshness](../slos-and-freshness.md)
- Schedule pipelines: [How-to: Schedule pipelines with CronJob](../../how-to/schedule-pipelines.md)
- Pipeline failed runbook: [Runbook: Pipeline failed](pipeline-failed.md)
- Roll back artifacts safely: [How-to: Roll back artifacts safely](../../how-to/rollback-safely.md)
