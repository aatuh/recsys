---
diataxis: explanation
tags:
  - explanation
  - architecture
  - developer
  - business
  - ops
---
# How it works: architecture and data flow
This page explains How it works: architecture and data flow and how it fits into the RecSys suite.


## Who this is for

- Stakeholders and engineers who want a clear mental model of what runs where, and how auditability is produced.

## What you will get

- which components exist in the suite
- how a request turns into a recommendation
- how audit artifacts (logs → reports → decisions) are produced
- how ship/rollback works at a high level

!!! info "Key terms"
    - **[Tenant](../project/glossary.md#tenant)**: a configuration + data isolation boundary.
    - **[Surface](../project/glossary.md#surface)**: where recommendations are shown (home, PDP, cart, ...).
    - **[Artifact](../project/glossary.md#artifact)**: an immutable, versioned blob produced offline and consumed online.
    - **[Manifest](../project/glossary.md#manifest)**: maps artifact types to artifact URIs for a `(tenant, surface)` pair.
    - **[Exposure log](../project/glossary.md#exposure-log)**: what was shown (audit trail + evaluation input).

## One-screen mental model

RecSys separates concerns across four modules:

- **Serving (online, deterministic):** `recsys-service`
- **Ranking logic (deterministic):** `recsys-algo`
- **Computation (offline, versioned outputs):** `recsys-pipelines`
- **Evaluation (analysis + decisions):** `recsys-eval`

See also:

- Suite architecture: [Suite architecture](suite-architecture.md)
- Suite context diagram: [Suite Context](../start-here/diagrams/suite-context.md)

## Online request flow (serve)

1. Your app calls `POST /v1/recommend` (see [API Reference](../reference/api/api-reference.md)).
2. `recsys-service` builds a candidate set and ranks deterministically (see
   [Candidate generation vs ranking](candidate-vs-ranking.md)).
3. Optional trace/explain data can be enabled depending on your ranking setup (see
   [Concepts](../recsys-algo/concepts.md)).

Scoring details: [Scoring model specification (recsys-algo)](../recsys-algo/scoring-model.md)

## Determinism and auditability contract

RecSys is designed so you can answer two questions reliably:

- What did we serve (and what changed)?
- Can we reproduce and evaluate decisions from logs later?

### What “deterministic” means here (operational definition)

For a given tenant, `POST /v1/recommend` is deterministic when the following inputs are the same:

- request payload (including `surface`, identifiers, and any per-request overrides)
- `recsys-service` / `recsys-algo` version
- serving inputs (DB rows in DB-only mode, or the same artifacts + manifest in artifact mode)
- tenant config and rules

You can verify you are comparing like-for-like by recording:

- `meta.request_id` (or the `X-Request-Id` you provided)
- `meta.config_version` and `meta.rules_version` (ETags)

See: [API Reference](../reference/api/api-reference.md)

### What can (and should) make results change

- Updated signals/artifacts (new day/window, refreshed pipelines outputs, changed catalog).
- Config/rules updates (versions change).
- Ship/rollback by switching the current manifest pointer (artifact mode).
- Experiment assignment (if enabled).

### What this does not guarantee

- KPI lift (depends on your data and experimentation discipline).
- Production readiness (use [Production readiness checklist (RecSys suite)](../operations/production-readiness-checklist.md)).

Known limitations and non-goals live here: [Known limitations and non-goals (current)](../start-here/known-limitations.md)

## Logging flow (audit trail)

RecSys produces an audit trail by linking:

- **Exposure logs** (what the user saw)
- **Outcome logs** (what the user did)

The join key is `request_id` (plus stable identifiers).
See: [Exposure logging and attribution](exposure-logging-and-attribution.md)

## Evaluation flow (decide ship/hold/rollback)

1. Validate logs and compute join-rate.
2. Produce an offline/online report (see [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)).
3. Use the report to decide: **ship / hold / rollback**.

Deep dives live under:

- `recsys-eval`: [recsys-eval docs](../recsys-eval/docs/index.md)

## Data modes (where features come from)

RecSys supports two primary serving modes:

- **DB-only mode:** simplest way to start; fewer moving parts.
- **Artifact/manifest mode:** versioned artifacts + a “current manifest pointer”.

See: [Data modes: DB-only vs artifact/manifest](data-modes.md)

## Ship/rollback mechanics (why it’s safe)

- Config/rules changes are explicit and auditable (admin API).
- Artifact mode allows versioned rollback by switching the current manifest pointer.
- The suite is designed so rollbacks are operationally predictable.

See:

- Operational reliability & rollback: [Operational reliability and rollback](../start-here/operational-reliability-and-rollback.md)
- Production-like tutorial: [production-like run (pipelines → object store → ship/rollback)](../tutorials/production-like-run.md)

## Read next

- Quickstart (minimal): [Tutorial: Quickstart (minimal)](../tutorials/quickstart-minimal.md)
- Quickstart (full validation): [Tutorial: Quickstart (full validation)](../tutorials/quickstart.md)
- Pilot plan: [Pilot plan (2–6 weeks)](../start-here/pilot-plan.md)
- Capability matrix (scope and non-scope): [Capability matrix (scope and non-scope)](capability-matrix.md)
- Known limitations: [Known limitations and non-goals (current)](../start-here/known-limitations.md)
