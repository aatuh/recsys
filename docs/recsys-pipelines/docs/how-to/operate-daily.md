---
tags:
  - how-to
  - ops
  - recsys-pipelines
---

# How-to: Operate pipelines daily

## Who this is for

- SRE / on-call running `recsys-pipelines` on a schedule
- Data engineers responsible for freshness and correctness

## Goal

Run pipelines predictably, detect staleness early, and respond to failures using the right runbook.

## Quick paths

- Schedule runs: [`how-to/schedule-pipelines.md`](schedule-pipelines.md)
- Incremental runs: [`how-to/run-incremental.md`](run-incremental.md)
- Debug failures: [`how-to/debug-failures.md`](debug-failures.md)
- SLOs and freshness: [`operations/slos-and-freshness.md`](../operations/slos-and-freshness.md)
- Runbooks:
  - Pipeline failed: [`operations/runbooks/pipeline-failed.md`](../operations/runbooks/pipeline-failed.md)
  - Validation failed: [`operations/runbooks/validation-failed.md`](../operations/runbooks/validation-failed.md)
  - Stale artifacts: [`operations/runbooks/stale-artifacts.md`](../operations/runbooks/stale-artifacts.md)
  - Limit exceeded: [`operations/runbooks/limit-exceeded.md`](../operations/runbooks/limit-exceeded.md)

## Daily checklist (practical)

1. Confirm the expected schedule and windowing

- If you run nightly/daily: verify `--start/--end` semantics and UTC windows.
- If you run incremental: ensure `checkpoint_dir` is stable across runs.

1. Run and publish (or confirm the scheduler did)

- Primary: `recsys-pipelines run ... --incremental`  
  See: [`how-to/run-incremental.md`](run-incremental.md)

1. Validate outputs and “current” pointers

- Check expected files/paths: [`reference/output-layout.md`](../reference/output-layout.md)
- Confirm manifest pointer updated only when validation passes.

1. Watch freshness/SLOs

- Use the invariants and expected freshness windows: [`operations/slos-and-freshness.md`](../operations/slos-and-freshness.md)

1. When something fails, open the closest runbook first

- Pipeline failed: [`operations/runbooks/pipeline-failed.md`](../operations/runbooks/pipeline-failed.md)
- Validation failed: [`operations/runbooks/validation-failed.md`](../operations/runbooks/validation-failed.md)
- Stale artifacts: [`operations/runbooks/stale-artifacts.md`](../operations/runbooks/stale-artifacts.md)
- Limit exceeded: [`operations/runbooks/limit-exceeded.md`](../operations/runbooks/limit-exceeded.md)

## Read next

- Start here: [`start-here.md`](../start-here.md)
- Config reference: [`reference/config.md`](../reference/config.md)
- Exit codes: [`reference/exit-codes.md`](../reference/exit-codes.md)
