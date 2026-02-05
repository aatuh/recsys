---
tags:
  - explanation
  - ml
  - developer
  - recsys-algo
---

# Candidate generation vs ranking

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

You can influence blending per request:

- `weights.pop`: popularity contribution
- `weights.cooc`: co-visitation contribution
- `weights.emb`: “similarity” contribution (a bucket for non-pop/non-cooc signals)

If you omit `weights`, the service uses its configured defaults (see `RECSYS_ALGO_*` config in
[`reference/config/recsys-service.md`](../reference/config/recsys-service.md)).

## Controls: constraints and rules (order matters)

After the algorithm produces ranked items, the service applies controls in this order:

1. **Rule pins**: pinned items are moved to the top.
   - Pin rules can **inject** items that were not in the candidate pool.
   - If injection happens, you will see `RULE_PIN_INJECTED`.
2. **Post-ranking constraints** (when tag data is available):
   - `constraints.forbidden_tags`
   - `constraints.max_per_tag`
   - If filtering happens, you will see `CONSTRAINTS_FILTERED`.
3. **Candidate allow-list**:
   - `candidates.include_ids` is applied as a final allow-list filter.
   - If it removes some results, you will see `CANDIDATES_INCLUDE_FILTERED`.
   - If it removes everything, you will see `CANDIDATES_INCLUDE_EMPTY`.

Additionally:

- `candidates.exclude_ids` is treated as a strict exclusion and removes those items from consideration.
- `constraints.required_tags` is used to require at least one tag match (for example: require a category tag).

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
  [`operations/runbooks/empty-recs.md`](../operations/runbooks/empty-recs.md)

## Read next

- Exposure logging (to measure impact): [`explanation/exposure-logging-and-attribution.md`](exposure-logging-and-attribution.md)
- Surface namespaces (avoid mismatches): [`explanation/surface-namespaces.md`](surface-namespaces.md)
- Data modes (DB-only vs artifact/manifest): [`explanation/data-modes.md`](data-modes.md)
