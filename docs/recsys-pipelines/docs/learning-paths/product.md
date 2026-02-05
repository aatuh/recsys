
# Learning path: Product / PM

## What you care about

- What artifacts exist and what they mean
- How freshness works (daily windows)
- How to roll back if something goes wrong
- What "data quality" means in practice

## Read in this order

1) `start-here.md`
1) `explanation/artifacts-and-versioning.md`
1) `operations/slos-and-freshness.md`
1) `operations/runbooks/pipeline-failed.md`

## Key concepts

- Artifacts are versioned and rollbackable because production needs safe

  recovery.

- Manifest pointers are updated last so serving never points to missing blobs.
- Validation gates exist to prevent bad artifacts from reaching users.

## Practical questions (and where answered)

- "How often do recommendations update?"
  - `operations/slos-and-freshness.md`
- "Can we revert to yesterday's artifact?"
  - `how-to/rollback-manifest.md`
- "What if data is missing for a day?"
  - `explanation/data-lifecycle.md`

## Read next

- Start here: [`start-here.md`](../start-here.md)
- Artifacts and versioning: [`explanation/artifacts-and-versioning.md`](../explanation/artifacts-and-versioning.md)
- SLOs and freshness: [`operations/slos-and-freshness.md`](../operations/slos-and-freshness.md)
- Roll back artifacts safely: [`how-to/rollback-safely.md`](../how-to/rollback-safely.md)
