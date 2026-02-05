
# Runbook: Stale artifacts

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

- SLOs and freshness: [`operations/slos-and-freshness.md`](../slos-and-freshness.md)
- Schedule pipelines: [`how-to/schedule-pipelines.md`](../../how-to/schedule-pipelines.md)
- Pipeline failed runbook: [`operations/runbooks/pipeline-failed.md`](pipeline-failed.md)
- Roll back artifacts safely: [`how-to/rollback-safely.md`](../../how-to/rollback-safely.md)
