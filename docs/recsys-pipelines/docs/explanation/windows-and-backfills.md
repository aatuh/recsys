
# Windows and backfills

## Window semantics

This repo uses daily windows in UTC:

- A day window is [00:00, 24:00) UTC
- Backfills iterate start..end (inclusive end date)

## Why daily windows

- Easy operational model
- Simple freshness SLOs
- Deterministic partitioning

## Backfill safety

Backfills should be safe because:

- canonical partitions are replaced idempotently
- publishing updates the pointer last

Still, you should backfill gradually and watch limits.

See `how-to/run-backfill.md`.

## Read next

- Backfill safely: [`how-to/backfill-safely.md`](../how-to/backfill-safely.md)
- Run a backfill: [`how-to/run-backfill.md`](../how-to/run-backfill.md)
- Validation and guardrails: [`explanation/validation-and-guardrails.md`](validation-and-guardrails.md)
- Limit exceeded runbook: [`operations/runbooks/limit-exceeded.md`](../operations/runbooks/limit-exceeded.md)
- Glossary: [`glossary.md`](../glossary.md)
