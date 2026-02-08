---
diataxis: how-to
tags:
  - recsys-pipelines
---
# How-to: Debug a failed pipeline run
This guide shows how to how-to: Debug a failed pipeline run in a reliable, repeatable way.


## 1) Identify the step

Look at logs for one of:

- ingest
- validate
- popularity
- cooc
- publish

## 2) Common root causes

- Input files missing or wrong path
- Bad JSON in raw event files
- Validation fails (out-of-window timestamps, too many events)
- Resource limit exceeded (sessions/items)
- Disk permission errors

## 3) Useful commands

Re-run one day:

```bash
./bin/recsys-pipelines run --config configs/env/local.json --tenant demo \
  --surface home --start 2026-01-01 --end 2026-01-01
```

Check manifest:

```bash
cat .out/registry/current/demo/home/manifest.json
```

Inspect canonical files:

```bash
find .out/canonical -type f | sort
```

## 4) If publish failed

Publishing is ordered so that the manifest pointer updates last.
This means serving should still point to the previous version.

See `operations/runbooks/pipeline-failed.md`.

## Read next

- Operate pipelines daily: [How-to: Operate pipelines daily](operate-daily.md)
- Pipeline failed runbook: [Runbook: Pipeline failed](../operations/runbooks/pipeline-failed.md)
- Validation failed runbook: [Runbook: Validation failed](../operations/runbooks/validation-failed.md)
- Limit exceeded runbook: [Runbook: Limit exceeded](../operations/runbooks/limit-exceeded.md)
- Output layout (verify “current”): [Output layout (local filesystem)](../reference/output-layout.md)
