---
tags:
  - reference
  - data-contracts
  - developer
  - evaluation
---

# Field catalog (evaluation events)

## Who this is for

- Developers mapping product events into the `recsys-eval` schemas
- Data/analytics owners reviewing "what fields do we actually need?"

## What you will get

- A quick catalog of the most important fields (what they mean and why they matter)
- Pointers to the strict schemas and JSONL examples

## Reference

The canonical strict schemas and JSONL examples live here:

- [`reference/data-contracts/eval-events.md`](eval-events.md)

### `exposure.v1` (what was shown)

Required:

- `request_id`: join key to outcomes/assignments (unique per rendered list)
- `user_id`: stable pseudonymous user identifier
- `ts`: RFC3339 timestamp
- `items[]`: ranked list of `{ item_id, rank }` (rank is 1-based)

Strongly recommended (in `context`, as strings):

- `tenant_id`: tenant scope
- `surface`: placement/surface name (`home`, `pdp`, ...)
- `segment`: segment/bucket when applicable
- `locale`, `device`, `user_type`: common slice keys

Optional (guardrails/debugging):

- `latency_ms`: request latency (for p95/p99 guardrails)
- `error`: whether the call failed or returned degraded results
- `items[].propensity`, `items[].logging_propensity`, `items[].target_propensity`: for OPE and experiment analysis

### `outcome.v1` (what the user did)

Required:

- `request_id`: join key to the exposure
- `user_id`: stable pseudonymous user identifier
- `item_id`: the item acted on
- `event_type`: `click` or `conversion`
- `ts`: RFC3339 timestamp

Optional:

- `value`: numeric value (for revenue-based metrics)

### `assignment.v1` (which variant was shown)

Required:

- `experiment_id`: identifier for the experiment
- `variant`: variant label (`A`, `B`, ...)
- `request_id`: join key to the exposure/outcome
- `user_id`: stable pseudonymous user identifier
- `ts`: RFC3339 timestamp

Optional:

- `context`: string map of slice keys (surface, tenant_id, locale, ...)

## Examples

Minimal `exposure.v1` line:

```json
{"request_id":"req-1","user_id":"user-1","ts":"2026-01-30T12:00:00Z","items":[{"item_id":"item-1","rank":1},{"item_id":"item-2","rank":2}],"context":{"tenant_id":"demo","surface":"home"}}
```

## Read next

- Strict schemas and examples: [`reference/data-contracts/eval-events.md`](eval-events.md)
- Join rules and join-rate debugging: [`reference/data-contracts/join-logic.md`](join-logic.md)
- Minimum instrumentation (practical checklist): [`reference/minimum-instrumentation.md`](../minimum-instrumentation.md)
