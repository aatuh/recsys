---
diataxis: reference
tags:
  - reference
  - data-contracts
  - developer
---
# Data contracts
This page is the canonical reference for Data contracts.


## Who this is for

- Lead developers and data engineers implementing logging, pipelines, and data validation
- Analysts and recommendation engineers running `recsys-eval`
- Operators who need to reason about “what was served” vs “what was clicked” vs “what artifact version is
  live”

## What you will get

- The contract types used across the suite (serving, evaluation, pipelines)
- Minimal examples you can copy/paste
- Where the canonical schemas live and how they are versioned

## Overview: three contract families

- **Evaluation events (for `recsys-eval`)**
  - Purpose: measure quality (offline regression, experiments).
  - Join key: `request_id` (exposures ↔ outcomes ↔ assignments).
  - Details + examples: [`Eval events`](eval-events.md)
  - Join semantics: [`Event join logic`](join-logic.md)
  - Minimum instrumentation spec: [`Minimum instrumentation`](../minimum-instrumentation.md)

- **Serving logs (what the service emitted)**
  - Purpose: auditable “what was served” record.
  - Canonical schema: [Exposure schema (JSON)](exposures.schema.json)

- **Pipelines + artifacts (what pipelines consume/publish)**
  - Purpose: convert interactions into versioned artifacts and a manifest pointer.
  - Interaction schema: [Interactions schema (JSON)](interactions.schema.json)
  - Manifest schema: [Manifest schema (JSON)](artifacts/manifest.schema.json)

## Evaluation events (recsys-eval): what you must be able to produce

If your goal is “measure lift” or “decide what to ship”, implement these:

- `exposure.v1` (what you showed)
- `outcome.v1` (what the user did)
- `assignment.v1` (optional; experiment bucket)

Minimal JSONL examples (one object per line):

```json
{"request_id":"req-1","user_id":"u_1","ts":"2026-02-05T10:00:00Z","items":[{"item_id":"item_1","rank":1},{"item_id":"item_2","rank":2}],"context":{"tenant_id":"demo","surface":"home"}}
{"request_id":"req-1","user_id":"u_1","item_id":"item_2","event_type":"click","ts":"2026-02-05T10:00:02Z"}
{"experiment_id":"exp-1","variant":"A","request_id":"req-1","user_id":"u_1","ts":"2026-02-05T10:00:00Z","context":{"tenant_id":"demo","surface":"home"}}
```

Validation:

```bash
recsys-eval validate --schema exposure.v1 --input exposures.jsonl
recsys-eval validate --schema outcome.v1 --input outcomes.jsonl
recsys-eval validate --schema assignment.v1 --input assignments.jsonl
```

Tip: `recsys-service` can emit eval-compatible exposures directly. See “Service exposure logs vs eval schema” in
[`Eval events`](eval-events.md).

## Serving exposure events (service-native): what was actually served

This event shape is useful for auditability, debugging, and building derived datasets. It is **not** the same as the
`recsys-eval` exposure schema (which is stricter and optimized for evaluation).

Canonical schema: [Exposure schema (JSON)](exposures.schema.json)

Minimal example:

```json
{
  "schema_version": "exposure.v1",
  "occurred_at": "2026-02-05T10:00:00Z",
  "tenant_id": "demo",
  "request_id": "00000000-0000-0000-0000-000000000000",
  "surface": "home",
  "segment": "default",
  "served": [{ "item_id": "item_1", "rank": 1, "score": 0.12 }]
}
```

## Interaction events (pipelines): what happened in the product

This is the minimal “something happened” record used by pipelines.

Canonical schema: [Interactions schema (JSON)](interactions.schema.json)

Minimal example:

```json
{
  "schema_version": "interaction.v1",
  "occurred_at": "2026-02-05T10:00:02Z",
  "tenant_id": "demo",
  "event_type": "click",
  "item_id": "item_2"
}
```

If you need reliable evaluation joins, produce `outcome.v1` for `recsys-eval` (it requires `request_id` and
`user_id`).

## Artifact manifest (pipelines → service): what version is live

In artifact/manifest mode, pipelines publish artifacts and update a manifest pointer. The service reads the current
manifest and fetches referenced blobs.

Canonical schema: [Manifest schema (JSON)](artifacts/manifest.schema.json)

Minimal example:

```json
{
  "schema_version": "manifest.v1",
  "tenant_id": "demo",
  "created_at": "2026-02-05T10:05:00Z",
  "version": "2026-02-05T10:05:00Z",
  "artifacts": {}
}
```

## Versioning rules (practical)

- **Never change the meaning of an existing version.**
  - Add a new version instead (for example: `interaction.v2`), and keep transforms explicit.
- **Treat schemas as strict by default.**
  - `recsys-eval validate` uses JSON Schema with strictness that will reject missing required fields and unexpected keys.
- **Keep IDs stable and privacy-safe.**
  - Use pseudonymous user IDs; do not log raw PII.

## Read next

- Exposure logging and attribution: [Exposure logging and attribution](../../explanation/exposure-logging-and-attribution.md)
- How shipping/rollback ties to contracts: [Suite architecture](../../explanation/suite-architecture.md)
- Data modes (DB-only vs artifact/manifest): [Data modes: DB-only vs artifact/manifest](../../explanation/data-modes.md)
