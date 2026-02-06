---
tags:
  - explanation
  - architecture
  - developer
  - business
  - ops
---

# How it works: architecture and data flow

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

- Suite architecture: [`explanation/suite-architecture.md`](suite-architecture.md)
- Suite context diagram: [`start-here/diagrams/suite-context.md`](../start-here/diagrams/suite-context.md)

## Online request flow (serve)

1. Your app calls `POST /v1/recommend` (see [`reference/api/api-reference.md`](../reference/api/api-reference.md)).
2. `recsys-service` builds a candidate set and ranks deterministically (see
   [`explanation/candidate-vs-ranking.md`](candidate-vs-ranking.md)).
3. Optional trace/explain data can be enabled depending on your ranking setup (see
   [`recsys-algo/concepts.md`](../recsys-algo/concepts.md)).

Scoring details: [`recsys-algo/scoring-model.md`](../recsys-algo/scoring-model.md)

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

See: [`reference/api/api-reference.md`](../reference/api/api-reference.md)

### What can (and should) make results change

- Updated signals/artifacts (new day/window, refreshed pipelines outputs, changed catalog).
- Config/rules updates (versions change).
- Ship/rollback by switching the current manifest pointer (artifact mode).
- Experiment assignment (if enabled).

### What this does not guarantee

- KPI lift (depends on your data and experimentation discipline).
- Production readiness (use [`operations/production-readiness-checklist.md`](../operations/production-readiness-checklist.md)).

Known limitations and non-goals live here: [`start-here/known-limitations.md`](../start-here/known-limitations.md)

## Logging flow (audit trail)

RecSys produces an audit trail by linking:

- **Exposure logs** (what the user saw)
- **Outcome logs** (what the user did)

The join key is `request_id` (plus stable identifiers).
See: [`explanation/exposure-logging-and-attribution.md`](exposure-logging-and-attribution.md)

## Evaluation flow (decide ship/hold/rollback)

1. Validate logs and compute join-rate.
2. Produce an offline/online report (see [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)).
3. Use the report to decide: **ship / hold / rollback**.

Deep dives live under:

- `recsys-eval`: [`recsys-eval/docs/index.md`](../recsys-eval/docs/index.md)

## Data modes (where features come from)

RecSys supports two primary serving modes:

- **DB-only mode:** simplest way to start; fewer moving parts.
- **Artifact/manifest mode:** versioned artifacts + a “current manifest pointer”.

See: [`explanation/data-modes.md`](data-modes.md)

## Ship/rollback mechanics (why it’s safe)

- Config/rules changes are explicit and auditable (admin API).
- Artifact mode allows versioned rollback by switching the current manifest pointer.
- The suite is designed so rollbacks are operationally predictable.

See:

- Operational reliability & rollback: [`start-here/operational-reliability-and-rollback.md`](../start-here/operational-reliability-and-rollback.md)
- Production-like tutorial: [`tutorials/production-like-run.md`](../tutorials/production-like-run.md)

## Read next

- Quickstart: [`tutorials/quickstart.md`](../tutorials/quickstart.md)
- Pilot plan: [`start-here/pilot-plan.md`](../start-here/pilot-plan.md)
- Capability matrix (scope and non-scope): [`explanation/capability-matrix.md`](capability-matrix.md)
- Known limitations: [`start-here/known-limitations.md`](../start-here/known-limitations.md)
