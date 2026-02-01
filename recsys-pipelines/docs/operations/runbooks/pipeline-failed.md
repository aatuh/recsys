
# Runbook: Pipeline failed

## Symptoms

- `recsys-pipelines run` exits non-zero
- job binary exits non-zero

## Immediate safety check

Publishing updates the manifest pointer last.
If publish failed mid-run, serving should still point to the previous version.

## Triage checklist

1) Identify which step failed (logs): ingest / validate / compute / publish
2) Check disk paths and permissions
3) Check raw input presence and format
4) If validation failed, inspect the reported rule
5) If limit exceeded, see `runbooks/limit-exceeded.md`

## Recovery

- Fix root cause
- Re-run the affected day only
- If publish already updated to a bad version, roll back the manifest

See `how-to/rollback-manifest.md`.
