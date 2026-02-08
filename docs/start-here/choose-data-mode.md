---
diataxis: how-to
tags:
  - start-here
  - serving
  - pipelines
  - artifact-mode
  - db-only
---
# Choose your data mode
This guide shows how to choose your data mode in a reliable, repeatable way.


## Who this is for

- Developers and platform engineers deciding how to integrate and operate RecSys.

## What you will get

- A simple decision rule for choosing **DB-only** vs **artifact/manifest** mode
- The operational tradeoffs (rollback, freshness, complexity)

## The two modes

### DB-only mode

Use DB-only mode when you want:

- Fastest time-to-first-success (few moving parts)
- A simple integration pilot (API + tenancy + exposure logging)
- Manual or external control of data updates

You provide signals by writing directly into the serving DB tables.

Start with: [Quickstart (10 minutes)](../tutorials/quickstart.md)

### Artifact/manifest mode

Use artifact/manifest mode when you want:

- Atomic publish + rollback of signals
- Clear separation of “build” vs “serve”
- A production-like operating model (pipelines produce artifacts, the manifest pointer drives serving)

Start with: [production-like run (pipelines → object store → ship/rollback)](../tutorials/production-like-run.md)

## Decision rule (one screen)

Choose **DB-only** if:

- You’re piloting integration and do not need automated daily refresh yet
- You can tolerate manual/semi-manual data updates during the pilot

Choose **artifact/manifest** if:

- You need safe ship/rollback and repeatable daily refresh from day 1
- You want evaluation + releases to move through a controlled publish pipeline

## Read next

- Data modes in depth: [Data modes: DB-only vs artifact/manifest](../explanation/data-modes.md)
- Operational reliability: [Operational reliability and rollback](operational-reliability-and-rollback.md)
- Pilot plan: [Pilot plan (2–6 weeks)](pilot-plan.md)
