
# Learning path: SRE / On-call

## Goals

- Know what "healthy" looks like
- Detect stale artifacts
- Triage failures quickly
- Roll back safely

## Read in this order

1) `operations/slos-and-freshness.md`
2) `operations/runbooks/pipeline-failed.md`
3) `operations/runbooks/stale-artifacts.md`
4) `operations/runbooks/limit-exceeded.md`
5) `how-to/rollback-manifest.md`

## What to alert on

- No successful publish within expected window (freshness)
- Validation failures
- Limit exceeded errors (resource protection)
