---
tags:
  - reference
  - evaluation
  - developer
---

# Minimum instrumentation spec (for credible evaluation)

## Who this is for

- Developers implementing exposure/outcome logging for RecSys
- Data engineers preparing datasets for `recsys-eval`
- Teams that want “ship / hold / rollback” decisions to be auditable

## What you will get

- The minimum events and fields required for trustworthy KPI + guardrail metrics
- The join key (`request_id`) and the invariants you must enforce
- Common pitfalls that cause low join-rate or misleading results

## Core invariants (do not compromise)

1. **`request_id` is unique per rendered list.**
2. The same **`request_id` propagates** exposure → outcome → assignment (if experiments).
3. `user_id` is **stable and pseudonymous** (do not log raw PII).
4. `surface` and `tenant_id` are present (recommended in `context`) and names are stable.
5. Timestamps are **RFC3339** (`ts`) and use a consistent clock source.

## Events you must produce

The strict schemas and examples live here:

- [`reference/data-contracts/eval-events.md`](data-contracts/eval-events.md)

### `exposure.v1` (required)

Required fields:

- `request_id`
- `user_id`
- `ts`
- `items[]` with `{ item_id, rank }` (`rank` is 1-based)

Strongly recommended `context` keys (string values):

- `tenant_id`
- `surface`
- `segment` (if you segment recommendations)

Optional (but useful guardrails):

- `latency_ms` (p95/p99 guardrail and rollout safety)
- `error` (detect “served empty because of error”)

### `outcome.v1` (required)

Required fields:

- `request_id`
- `user_id`
- `item_id`
- `event_type` (`click` or `conversion`)
- `ts`

Optional:

- `value` (required if you want revenue-based metrics)

Mapping guidance:

- Map your product events into `click` and `conversion` consistently.
  - Example: add-to-cart could be treated as `conversion` for a top-of-funnel pilot.

### `assignment.v1` (required for experiments)

Required fields:

- `experiment_id`
- `variant`
- `request_id`
- `user_id`
- `ts`

## KPI specs (minimum)

### CTR (click-through rate)

- **Definition:** `click` outcomes / exposures
- **Join key:** `request_id`
- **Required fields:**
  - exposure: `request_id`, `user_id`, `items[].item_id`, `items[].rank`
  - outcome: `request_id`, `user_id`, `item_id`, `event_type=click`
- **Common pitfalls:**
  - outcomes missing `request_id` (join-rate collapses)
  - clicks logged without `item_id` (can’t attribute to rank)

### Conversion rate

- **Definition:** `conversion` outcomes / exposures
- **Join key:** `request_id`
- **Required fields:**
  - exposure: `request_id`, `user_id`, `items[].item_id`
  - outcome: `request_id`, `user_id`, `item_id`, `event_type=conversion`
- **Common pitfalls:**
  - logging conversions without the originating recommendation `request_id`
  - reusing a single `request_id` for multiple renders (attribution smears)

### Revenue per exposure

- **Definition:** `sum(value)` for `conversion` outcomes / exposures
- **Join key:** `request_id`
- **Required fields:**
  - outcome: `event_type=conversion`, `value`
- **Common pitfalls:**
  - missing/zero `value` (metric becomes meaningless)
  - currency/unit mismatches (document the unit)

### Offline ranking proxies (HitRate@K, NDCG@K, MAP@K, …)

- **Definition:** compare the ranked list (`exposure.v1.items[]`) to outcomes as relevance signals
- **Join key:** `request_id`
- **Required fields:**
  - exposure: `items[].rank` (1-based), `items[].item_id`
  - outcome: `item_id`, `event_type` (what counts as “relevant”)
- **Common pitfalls:**
  - treating all outcome events as equally relevant (be explicit)
  - running offline metrics on a dataset with low join-rate

## Guardrails (minimum)

### Join integrity (must pass before trusting any KPI)

- **What you need:** `request_id` present in exposures + outcomes (+ assignments in experiments)
- **Join key:** `request_id`
- **Common pitfalls:**
  - generating `request_id` twice (one for the API call, another for logging)
  - not propagating `request_id` to client-side outcome events

See: [`reference/data-contracts/join-logic.md`](data-contracts/join-logic.md)

### Empty-recs rate

- **What you need:** `exposure.v1.items[]` (may be empty)
- **Join key:** none (computed from exposures)
- **Common pitfalls:**
  - logging exposures for requests you never rendered
  - treating “empty because error” as a normal exposure (use `error` when possible)

### Latency and error rate

- **What you need:** `exposure.v1.latency_ms` and `exposure.v1.error` (optional fields supported by schema)
- **Join key:** none (computed from exposures)
- **Common pitfalls:**
  - measuring latency in the wrong place (client vs server) and mixing the two
  - missing the long tail (track p95/p99, not only averages)

## Verify (Definition of Done)

- `recsys-eval validate` passes for `exposure.v1` and `outcome.v1` (and `assignment.v1` if experiments).
- Join-rate is measured by surface and is not near-zero.
- You can compute at least one KPI and one guardrail end-to-end.

## Read next

- Data contracts hub: [`reference/data-contracts/index.md`](data-contracts/index.md)
- Decision playbook (ship/hold/rollback): [`recsys-eval/docs/decision-playbook.md`](../recsys-eval/docs/decision-playbook.md)
- Integration checklist: [`how-to/integration-checklist.md`](../how-to/integration-checklist.md)
