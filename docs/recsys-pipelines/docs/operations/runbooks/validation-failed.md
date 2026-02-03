
# Runbook: Validation failed

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
