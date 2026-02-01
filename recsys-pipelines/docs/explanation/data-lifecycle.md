
# Data lifecycle

## Stages

1) Raw events
- Input is JSON Lines files (jsonl)
- Schema: `schemas/events/exposure.v1.json`

2) Canonical events
- Stored per day (UTC) per tenant/surface
- Written idempotently (replace per partition)

3) Validation
- Canonical is validated before any artifacts are computed/published

4) Artifact compute
- popularity: counts by item
- cooc: session-level co-occurrence

5) Staging (optional)
- Compute jobs can stage artifacts to `artifacts_dir`

6) Publish
- Versioned blob written to object store
- Registry record written
- Current manifest pointer updated last

## Why canonicalization exists

- Raw data is messy (missing fields, inconsistent formatting)
- Canonical events define a stable boundary

## Why validation gates exist

If you publish a bad artifact, you can degrade user experience immediately.
Validation prevents "bad data" from reaching serving.
