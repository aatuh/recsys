---
diataxis: how-to
tags:
  - how-to
  - ops
  - recsys-pipelines
---
# How-to: Operate pipelines daily
This guide shows how to how-to: Operate pipelines daily in a reliable, repeatable way.


## Who this is for

- SRE / on-call running `recsys-pipelines` on a schedule
- Data engineers responsible for freshness and correctness

## Goal

Run pipelines predictably, detect staleness early, and respond to failures using the right runbook.

## Quick paths

- Schedule runs: [How-to: Schedule pipelines with CronJob](schedule-pipelines.md)
- Incremental runs: [How-to: Run incremental pipelines](run-incremental.md)
- Debug failures: [How-to: Debug a failed pipeline run](debug-failures.md)
- SLOs and freshness: [SLOs and freshness](../operations/slos-and-freshness.md)
- Runbooks:
  - Pipeline failed: [Runbook: Pipeline failed](../operations/runbooks/pipeline-failed.md)
  - Validation failed: [Runbook: Validation failed](../operations/runbooks/validation-failed.md)
  - Stale artifacts: [Runbook: Stale artifacts](../operations/runbooks/stale-artifacts.md)
  - Limit exceeded: [Runbook: Limit exceeded](../operations/runbooks/limit-exceeded.md)

## Daily checklist (practical)

1. Confirm the expected schedule and windowing

- If you run nightly/daily: verify `--start/--end` semantics and UTC windows.
- If you run incremental: ensure `checkpoint_dir` is stable across runs.

1. Run and publish (or confirm the scheduler did)

- Primary: `recsys-pipelines run ... --incremental`  
  See: [How-to: Run incremental pipelines](run-incremental.md)

1. Validate outputs and “current” pointers

- Check expected files/paths: [Output layout (local filesystem)](../reference/output-layout.md)
- Confirm manifest pointer updated only when validation passes.

1. Watch freshness/SLOs

- Use the invariants and expected freshness windows: [SLOs and freshness](../operations/slos-and-freshness.md)

1. When something fails, open the closest runbook first

- Pipeline failed: [Runbook: Pipeline failed](../operations/runbooks/pipeline-failed.md)
- Validation failed: [Runbook: Validation failed](../operations/runbooks/validation-failed.md)
- Stale artifacts: [Runbook: Stale artifacts](../operations/runbooks/stale-artifacts.md)
- Limit exceeded: [Runbook: Limit exceeded](../operations/runbooks/limit-exceeded.md)

## Read next

- Start here: [Start here](../start-here.md)
- Config reference: [Config reference](../reference/config.md)
- Exit codes: [Exit codes](../reference/exit-codes.md)
