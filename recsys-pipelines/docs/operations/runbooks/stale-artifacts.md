
# Runbook: Stale artifacts

## Symptoms

- Manifest `updated_at` is older than expected
- Serving still uses old URIs

## Triage

1) Confirm scheduler ran
2) Confirm pipeline completed successfully
3) Check for validation failures or limit exceeded

## Recovery

- Re-run the missing day
- If inputs are missing, decide whether to publish empty artifacts or skip

See `how-to/run-backfill.md`.
