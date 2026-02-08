---
diataxis: how-to
tags:
  - how-to
  - ops
  - backfill
  - recsys-pipelines
---
# How-to: Backfill pipelines safely
This guide shows how to how-to: Backfill pipelines safely in a reliable, repeatable way.


## Who this is for

- Data engineers running historical reprocessing
- SRE / on-call handling late data, broken windows, or schema changes

## Goal

Recompute historical windows without breaking “current” artifacts, while staying within guardrails.

## Quick paths

- Run a backfill: [How-to: Run a backfill safely](run-backfill.md)
- Windows and backfills (concepts): [Windows and backfills](../explanation/windows-and-backfills.md)
- Validation and guardrails: [Validation and guardrails](../explanation/validation-and-guardrails.md)
- Output layout (verify results): [Output layout (local filesystem)](../reference/output-layout.md)

## Checklist (safe default)

1. Define the backfill window and why you need it

- Start small (1–3 days) to validate assumptions.

1. Run the backfill

- Follow the canonical command patterns: [How-to: Run a backfill safely](run-backfill.md)

1. Verify before publishing “current”

- Inspect output locations and manifest pointers: [Output layout (local filesystem)](../reference/output-layout.md)

1. Watch guardrails and resource limits

- Validation failures are designed to stop bad publishes:
  [Validation and guardrails](../explanation/validation-and-guardrails.md)

## Read next

- Roll back safely: [How-to: Roll back artifacts safely](rollback-safely.md)
- Validation failed runbook: [Runbook: Validation failed](../operations/runbooks/validation-failed.md)
- Limit exceeded runbook: [Runbook: Limit exceeded](../operations/runbooks/limit-exceeded.md)
