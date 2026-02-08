---
diataxis: explanation
tags:
  - explanation
  - ml
  - developer
  - recsys-algo
---
# Candidate generation vs ranking
This page explains Candidate generation vs ranking and how it fits into the RecSys suite.


## Who this is for

- Engineers integrating `recsys-service` who want predictable behavior
- Recommendation engineers explaining “why did we show this?”
- Operators debugging empty or surprising results

## What you will get

- A mental model for how a request becomes a ranked top-K list
- The difference between **candidate generation** (recall) and **ranking** (precision)
- How overrides, constraints, and rules interact (and in what order)
- A small worked example you can reason about

## Mental model (end-to-end)

Serving is a pipeline. Each stage has a different job:

1. **Candidate generation**: build a high-recall pool (many plausible items).
2. **Ranking**: score candidates and pick the best top-K.
3. **Policy and controls**: apply constraints and business rules.
4. **Response**: return items + metadata + warnings for debugging.

## Candidate generation (high recall)

Candidate generation answers: “What items should we even consider?”

This suite supports multiple candidate sources. Some sources may be unavailable depending on your data mode and which
signals you have built; the service returns warnings like `SIGNAL_UNAVAILABLE` or `SIGNAL_PARTIAL` when a signal cannot
contribute.

### Candidate sources used in this suite (conceptually)

- **Popularity** (always available in DB-only mode; great baseline)
  - “What’s trending / frequently interacted with in this surface/namespace?”
- **Co-visitation** (requires interaction history)
  - “Users who engaged with X also engaged with Y.”
- **Similarity signals** (optional; depends on what stores/signals you provide)
  - collaborative (“people like you”), content/tag similarity, and session-based similarity

### Anchors: how you “seed” candidates

Some candidate sources need one or more “anchor items” (for example: co-visitation).

In this API you can provide anchors explicitly:

- `anchors.item_ids` (explicit seed items)

In addition, `candidates.include_ids` is treated as an anchor source internally (useful when you want to force the
engine to consider a specific set of items).

## Ranking (precision)

Ranking answers: “Which of these candidates are best for this request?”

`recsys-algo` is deterministic and can explain its output. At a high level it:

- merges candidates from available sources
- computes per-item scores (a blend of signals)
- optionally adds explainability metadata (`options.explain`, `options.include_reasons`)

### Signal weights (`weights`)

RecSys supports request-time tuning via a **`weights`** object. Use it for safe, reversible changes (ship/rollback) without retraining or redeploying.

If you want the field-by-field contract, see:

- Reference: [Recommend request fields](../reference/api/recommend-request.md)

## Controls: constraints and rules (order matters)

After the algorithm produces ranked items, the service applies **controls** (policy and filtering) in a predictable order.

Typical controls:

1. **Pins** (move specific items to the top)
2. **Constraints** (tag-based filtering/limits)
3. **Candidate allow/deny lists**

Why it matters: controls can change the final list even if the underlying ranking stays the same.

Field reference:

- Reference: [Recommend request fields](../reference/api/recommend-request.md)
- Rules (admin/control plane): [Admin API + local bootstrap (recsys-service)](../reference/api/admin.md)

## Worked example: request → candidates → output

Request (illustrative):

```json
{
  "surface": "home",
  "k": 5,
  "user": { "user_id": "u_1" },
  "anchors": { "item_ids": ["item_10"] },
  "candidates": {
    "include_ids": ["item_1", "item_2", "item_99"],
    "exclude_ids": ["item_3"]
  },
  "constraints": {
    "forbidden_tags": ["adult"],
    "max_per_tag": { "category:shoes": 2 }
  },
  "options": { "include_reasons": true, "explain": "summary" }
}
```

How to think about it:

1. Candidate generation produces a pool, e.g.:
   - popularity candidates: `item_1, item_2, item_3, item_4, item_5`
   - co-visitation from anchor `item_10`: `item_6, item_7`
2. Ranking scores candidates and returns a ranked list (top-K):
   - `item_2, item_4, item_3, item_6, item_1`
3. Exclusions remove `item_3` (because it is explicitly excluded).
4. Post-ranking constraints remove items with `adult` tags and enforce `max_per_tag`.
5. Allow-list keeps only `item_1, item_2, item_99` (and drops everything else).
6. If a pin rule injected `item_99`, it can appear even if it was not in the original pool.

If the allow-list contains items that never appear, you will likely end up with fewer than `k` results (and a warning).

## Debugging checklist

- Call `POST /v1/recommend/validate` first to see the normalized request and early warnings.
- Check `warnings[]` in responses for:
  - `SIGNAL_UNAVAILABLE` / `SIGNAL_PARTIAL`
  - `CONSTRAINTS_FILTERED`
  - `CANDIDATES_INCLUDE_*`
  - `RULE_PIN_INJECTED`
- Turn on explainability for development:
  - `options.include_reasons=true` for per-item `reasons[]`
  - `options.explain=summary` or `options.explain=full` for structured `explain.signals`
- If you get empty results in production, follow the runbook:
  [Runbook: Empty recs](../operations/runbooks/empty-recs.md)

## Read next

- Exposure logging (to measure impact): [Exposure logging and attribution](exposure-logging-and-attribution.md)
- Surface namespaces (avoid mismatches): [Surface namespaces](surface-namespaces.md)
- Data modes (DB-only vs artifact/manifest): [Data modes: DB-only vs artifact/manifest](data-modes.md)
