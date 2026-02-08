---
diataxis: explanation
tags:
  - explanation
  - evaluation
  - ops
  - developer
  - recsys-service
  - recsys-eval
---
# Exposure logging and attribution
This page explains Exposure logging and attribution and how it fits into the RecSys suite.


## Who this is for

- Integrators wiring RecSys into a product (webshop, content feed, etc.)
- Recommendation engineers and analysts running `recsys-eval`
- Operators who need to debug “why are metrics wrong?” incidents

## What you will get

- The minimum you must log to measure recommendations
- How to attribute outcomes to exposures safely (joins that work)
- How to configure `recsys-service` to emit eval-compatible exposure logs
- Common logging bugs and the symptoms they cause

## The one rule: log exposures

If you only log clicks/purchases (outcomes) but not “what you showed” (exposures), you cannot evaluate recommendation
quality. Clicks without exposures are not attributable.

Exposure logging is also the foundation for:

- offline regression (quality gates before shipping)
- online experiments (measuring lift)
- incident debugging (“did we serve the wrong config/rules/algo?”)

## End-to-end flow (request → exposure → outcome)

```mermaid
sequenceDiagram
  participant Client
  participant Service as recsys-service
  participant DB as Postgres
  participant Store as Object store
  participant Log as Exposure log
  participant Events as Outcome events

  Client->>Service: POST /v1/recommend
  Service->>DB: Load config + rules
  alt Artifact mode enabled
    Service->>Store: Fetch manifest + artifacts
  end
  Service-->>Client: Response (items[], request_id)
  Service->>Log: Write exposure.v1 (request_id, items, context)
  Client->>Events: Emit outcome.v1 (request_id, item_id, ts)
```

The join key is `request_id`. Your product must carry it from the recommend call to the outcome event.

## What to log (minimum viable)

At minimum you need:

1. **Exposure** (ranked list you showed)
2. **Outcome** (click/conversion you observed)
3. **Assignment** (optional, for A/B testing)

All joins are driven by **`request_id`**.

Field-level contract (schemas + examples):

- Reference: [Exposure, outcome, and assignment schemas](../reference/data-contracts/exposure-outcome-assignment.md)
- Reference: [Event join logic (exposures ↔ outcomes ↔ assignments)](../reference/data-contracts/join-logic.md)

## Getting `request_id` right (attribution correctness)

To attribute outcomes to the right exposure, you need a single `request_id` that flows:

`client request → recsys-service response → outcome event`

You have two common options:

- **Client-supplied request IDs**: set `X-Request-Id` when calling `/v1/recommend`, then reuse that ID in outcome events.
- **Server-generated request IDs**: read `meta.request_id` from the response and attach it to outcome events.

Pick one, implement it consistently, and test joins early with `recsys-eval validate`.

## `recsys-service` exposure logging (built-in)

The service can write exposure logs as JSONL to a file or a directory:

- `EXPOSURE_LOG_ENABLED=true`
- `EXPOSURE_LOG_PATH=/app/tmp/exposures.jsonl` (file) or `/app/tmp/` (directory)
- `EXPOSURE_LOG_RETENTION_DAYS=30` (directory mode rotates and prunes old files)

There are two output formats:

- `service_v1`: service-native exposure event (good for audit/debugging)
- `eval_v1`: `recsys-eval` compatible `exposure.v1` records (recommended for evaluation workflows)

Recommended for evaluation:

```bash
EXPOSURE_LOG_ENABLED=true
EXPOSURE_LOG_FORMAT=eval_v1
EXPOSURE_LOG_PATH=/app/tmp/exposures.eval.jsonl
```

Config reference:

- [recsys-service configuration](../reference/config/recsys-service.md)

### Privacy: stable pseudonymous IDs

!!! warning
    Do not log raw PII (emails, names, phone numbers). Use stable pseudonymous identifiers and keep the hash salt
    secret.

The service logs **hashed** identifiers (HMAC-SHA256) rather than raw user IDs. Set a secret salt so hashes are stable
and non-guessable:

```bash
EXPOSURE_HASH_SALT=change-me-to-a-secret
```

If no user/session identifier is available for eval output, the service falls back to using `request_id` as `user_id` to
keep schemas valid (but this weakens user-level evaluation).

## Common logging bugs (and symptoms)

- **Only outcomes, no exposures**
  - Symptom: you can’t compute offline metrics; experiments are ambiguous.
- **Outcome events missing `request_id`**
  - Symptom: join rate collapses; reports look “too good” or “too bad” randomly.
- **Different `user_id` values in exposures vs outcomes**
  - Symptom: low join rate or joins that only work for some platforms (web vs app).
- **Multiple request IDs per single rendered list**
  - Symptom: duplicates, inconsistent attribution, confusing on-call investigations.
- **Logging raw PII**
  - Symptom: security review blocks adoption; you may breach internal policy.

## Verification checklist (do this early)

- Validate schemas:
  - `recsys-eval validate --schema exposure.v1 --input exposures.jsonl`
  - `recsys-eval validate --schema outcome.v1 --input outcomes.jsonl`
- Compute a basic join rate in your warehouse:
  - `% of exposures with at least one matching outcome by request_id`
- Ensure your top slicing keys exist (at minimum: `tenant_id`, `surface`).

## From logs to evaluation (what you do next)

Once your exposure/outcome logs validate and join, run an offline evaluation and produce a report you can share.

Start with the default pack:

- [`recsys-eval` default evaluation pack (recommended)](../recsys-eval/docs/default-evaluation-pack.md)
- Run eval and make ship/hold/rollback decisions: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)

## Read next

- Data contracts hub: [Data contracts](../reference/data-contracts/index.md)
- Event join logic: [Event join logic (exposures ↔ outcomes ↔ assignments)](../reference/data-contracts/join-logic.md)
- Experimentation model (A/B, interleaving, OPE): [Experimentation model (A/B, interleaving, OPE)](experimentation-model.md)
- Candidate vs ranking: [Candidate generation vs ranking](candidate-vs-ranking.md)
- Run eval and ship: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)
