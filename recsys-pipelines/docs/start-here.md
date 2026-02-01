
# Start here

## What this system does (non-technical)

Think of `recsys-pipelines` as a **factory**:

- It reads a stream of user activity events ("people saw item X")
- It cleans and stores them in a consistent format (canonical events)
- It computes simple recommendation building blocks (artifacts)
- It publishes those artifacts in a **versioned** and **rollbackable** way

The output artifacts are meant to be consumed by an online recommender service.

## What this system does (technical)

`recsys-pipelines` builds **deterministic, version-addressed artifacts** from
raw exposure events.

Current v1 artifact types:
- **popularity**: top-N items by exposure count
- **cooc**: item-item co-occurrence neighbors within a session

Key production properties:
- **Idempotent canonicalization**: reruns for the same day window do not
  duplicate events.
- **Atomic writes**: artifacts and pointers are written using temp+rename.
- **Validation gates**: publishing is blocked if validation fails.
- **Guardrails**: configurable limits prevent resource blowups.

## Who uses it

This repo is designed to be useful for:

- Product / PM: understand artifacts, freshness, rollback
- Engineers: run locally, add artifact types, integrate storage
- Data Engineering: define event contracts, backfills, data quality
- SRE / Platform: operate daily runs, alert on freshness, handle incidents

## How it fits in a recommendation stack

Typical stack (simplified):

- `recsys-pipelines` (this repo): offline artifact builder
- `recsys-algo`: ranking / scoring logic that consumes artifacts
- `recsys-service`: online API that serves recommendations using the algo

## Mental model (one-screen)

Raw events (jsonl)
   |
   v
Ingest + canonicalize (idempotent)
   |
   v
Validate canonical data (gates)
   |
   v
Compute artifacts (popularity, cooc)
   |
   v
Stage artifacts (optional)
   |
   v
Publish (atomic):
  - write versioned blob
  - write registry record
  - update current manifest pointer

## What you should do next

- Run locally: `tutorials/local-quickstart.md`
- Understand artifacts: `explanation/artifacts-and-versioning.md`
- Learn operations: `operations/slos-and-freshness.md`

If you are here because something broke, jump to:
- `operations/runbooks/`
