
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
