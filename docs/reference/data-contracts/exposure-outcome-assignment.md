---
diataxis: reference
tags:
  - reference
  - data-contracts
  - evaluation
---
# Exposure, outcome, and assignment schemas

This page is the **field-level reference** for the minimum events needed to evaluate recommendation impact.

## Who this is for

- Engineers and analysts implementing evaluation event logging.
- Integrators who need exact required fields for joinable telemetry.

## What you will get

- Required and recommended fields for `exposure.v1`, `outcome.v1`, and `assignment.v1`.
- JSONL examples for each schema.
- Links to join semantics and logging guidance.

## Location

- Event schemas and examples live under `docs/reference/data-contracts/`.

## Overview

You need two streams, plus an optional third:

1. **Exposure** (`exposure.v1`): what was shown, in what order
2. **Outcome** (`outcome.v1`): what the user did
3. **Assignment** (`assignment.v1`, optional): experiment variant per request

All joins are driven by `request_id`.

See also:

- Conceptual flow: [Exposure logging and attribution](../../explanation/exposure-logging-and-attribution.md)
- Join semantics: [Event join logic (exposures ↔ outcomes ↔ assignments)](join-logic.md)

---

## `exposure.v1`

### Required fields

- `request_id` (string): join key (stable per recommendation response)
- `user_id` (string): stable pseudonymous identifier (do not log raw PII)
- `ts` (string): RFC3339 timestamp
- `items` (array): list of `{ item_id, rank }` (rank is 1-based)

### Strongly recommended context

- `tenant_id`, `surface`, `segment`
- `algo_version`, `config_version`, `rules_version`

### Example (JSONL)

```json
{"request_id":"req-1","user_id":"u_hash_1","ts":"2026-02-05T10:00:00Z","items":[{"item_id":"item_1","rank":1},{"item_id":"item_2","rank":2}],"context":{"tenant_id":"demo","surface":"home","segment":"default"}}
```

---

## `outcome.v1`

### Required fields

- `request_id` (string): must match the exposure’s `request_id`
- `user_id` (string): should match the exposure’s `user_id` (recommended)
- `item_id` (string): item that was clicked/converted
- `event_type` (string): e.g. `click` or `conversion`
- `ts` (string): RFC3339 timestamp

### Example

```json
{"request_id":"req-1","user_id":"u_hash_1","item_id":"item_2","event_type":"click","ts":"2026-02-05T10:00:03Z"}
```

---

## `assignment.v1` (optional)

Log this if you run A/B tests so analysis can segment by variant.

### Required fields

- `experiment_id` (string)
- `variant` (string)
- `request_id` (string)
- `user_id` (string)
- `ts` (string): RFC3339 timestamp

### Example

```json
{"experiment_id":"exp-1","variant":"A","request_id":"req-1","user_id":"u_hash_1","ts":"2026-02-05T10:00:00Z","context":{"tenant_id":"demo","surface":"home"}}
```

---

## Privacy and safety rules

!!! warning
    Do not log raw PII (emails, names, phone numbers). Use stable pseudonymous IDs and keep any salt/secret out of logs.

For the built-in service logger and hashing controls, see:

- [recsys-service configuration](../config/recsys-service.md)
