---
tags:
  - overview
  - ops
  - developer
  - recsys-pipelines
---

# Start here

## Who this is for

- Lead developers / platform engineers evaluating the offline layer
- Data engineers operating daily runs and backfills
- SRE / on-call responding to freshness and pipeline failures
- Recommendation engineers who need to understand “what artifacts exist and when they update”

## What you will get

- A clear mental model of what `recsys-pipelines` does and where it fits in the RecSys suite
- The fastest paths to: run locally, operate daily, backfill, and roll back
- Pointers to the canonical output layout, config, and on-call runbooks

## Quick paths

<div class="grid cards" markdown>

- **[Run locally](tutorials/local-quickstart.md)**  
  10–20 min: ingest → validate → compute → publish using local config.
- **[Operate daily](how-to/operate-daily.md)**  
  What to run, what to watch, and which runbook to open first.
- **[Backfill safely](how-to/backfill-safely.md)**  
  Window selection, guardrails, and verification.
- **[Roll back safely](how-to/rollback-safely.md)**  
  Manifest rollback, safety checks, and verification.
- **[Output layout](reference/output-layout.md)**  
  Where artifacts/manifests live and what “current” means.
- **[Config reference](reference/config.md)**  
  The knobs that change behavior (sources, windows, guardrails, sinks).
- **[SLOs & freshness](operations/slos-and-freshness.md)**  
  Operational invariants and “is this stale?” reasoning.
- **[Runbooks](operations/runbooks/pipeline-failed.md)**  
  Common failures and safe remediation patterns.

</div>

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
- **implicit**: user→item scores from implicit feedback
- **content_sim**: item tags for content-based similarity
- **session_seq**: user→next-item sequence signals

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
Compute artifacts (popularity, cooc, implicit, content_sim, session_seq)
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

- Run locally: [`tutorials/local-quickstart.md`](tutorials/local-quickstart.md)
- Understand artifacts: [`explanation/artifacts-and-versioning.md`](explanation/artifacts-and-versioning.md)
- Learn operations: [`operations/slos-and-freshness.md`](operations/slos-and-freshness.md)

If you are here because something broke, jump to:

- Runbooks: [`operations/runbooks/pipeline-failed.md`](operations/runbooks/pipeline-failed.md)

## Read next

- Operate daily: [`how-to/operate-daily.md`](how-to/operate-daily.md)
- Backfills and windows: [`explanation/windows-and-backfills.md`](explanation/windows-and-backfills.md)
- Output layout: [`reference/output-layout.md`](reference/output-layout.md)
- Glossary: [`glossary.md`](glossary.md)
