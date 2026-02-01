
# How-to: Debug a failed pipeline run

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

```bash
# Re-run one day
./bin/recsys-pipelines run --config configs/env/local.json --tenant demo \
  --surface home --start 2026-01-01 --end 2026-01-01

# Check manifest
cat .out/registry/current/demo/home/manifest.json

# Inspect canonical files
find .out/canonical -type f | sort
```

## 4) If publish failed

Publishing is ordered so that the manifest pointer updates last.
This means serving should still point to the previous version.

See `operations/runbooks/pipeline-failed.md`.
