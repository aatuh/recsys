
# Runbook: Pipeline failed

## Symptoms

- `recsys-pipelines run` exits non-zero
- job binary exits non-zero

## Immediate safety check

Publishing updates the manifest pointer last.
If publish failed mid-run, serving should still point to the previous version.

## Triage checklist

1) Identify which step failed (logs): ingest / validate / compute / publish
1) Check disk paths and permissions
1) Check raw input presence and format
1) If validation failed, inspect the reported rule
1) If limit exceeded, see `runbooks/limit-exceeded.md`

## Recovery

- Fix root cause
- Re-run the affected day only
- If publish already updated to a bad version, roll back the manifest

See `how-to/rollback-manifest.md`.

## Read next

- Debug failures: [`how-to/debug-failures.md`](../../how-to/debug-failures.md)
- Roll back artifacts safely: [`how-to/rollback-safely.md`](../../how-to/rollback-safely.md)
- Validation failed runbook: [`operations/runbooks/validation-failed.md`](validation-failed.md)
- SLOs and freshness: [`operations/slos-and-freshness.md`](../slos-and-freshness.md)
