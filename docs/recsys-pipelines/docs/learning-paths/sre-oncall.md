
# Learning path: SRE / On-call

## Goals

- Know what "healthy" looks like
- Detect stale artifacts
- Triage failures quickly
- Roll back safely

## Read in this order

1) `operations/slos-and-freshness.md`
1) `operations/runbooks/pipeline-failed.md`
1) `operations/runbooks/stale-artifacts.md`
1) `operations/runbooks/limit-exceeded.md`
1) `how-to/rollback-manifest.md`

## What to alert on

- No successful publish within expected window (freshness)
- Validation failures
- Limit exceeded errors (resource protection)
