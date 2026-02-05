---
tags:
  - how-to
  - ops
  - backfill
  - recsys-pipelines
---

# How-to: Backfill pipelines safely

## Who this is for

- Data engineers running historical reprocessing
- SRE / on-call handling late data, broken windows, or schema changes

## Goal

Recompute historical windows without breaking “current” artifacts, while staying within guardrails.

## Quick paths

- Run a backfill: [`how-to/run-backfill.md`](run-backfill.md)
- Windows and backfills (concepts): [`explanation/windows-and-backfills.md`](../explanation/windows-and-backfills.md)
- Validation and guardrails: [`explanation/validation-and-guardrails.md`](../explanation/validation-and-guardrails.md)
- Output layout (verify results): [`reference/output-layout.md`](../reference/output-layout.md)

## Checklist (safe default)

1. Define the backfill window and why you need it

- Start small (1–3 days) to validate assumptions.

1. Run the backfill

- Follow the canonical command patterns: [`how-to/run-backfill.md`](run-backfill.md)

1. Verify before publishing “current”

- Inspect output locations and manifest pointers: [`reference/output-layout.md`](../reference/output-layout.md)

1. Watch guardrails and resource limits

- Validation failures are designed to stop bad publishes:
  [`explanation/validation-and-guardrails.md`](../explanation/validation-and-guardrails.md)

## Read next

- Roll back safely: [`how-to/rollback-safely.md`](rollback-safely.md)
- Validation failed runbook: [`operations/runbooks/validation-failed.md`](../operations/runbooks/validation-failed.md)
- Limit exceeded runbook: [`operations/runbooks/limit-exceeded.md`](../operations/runbooks/limit-exceeded.md)
