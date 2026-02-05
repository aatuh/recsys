---
tags:
  - explanation
  - ml
  - developer
  - recsys-algo
---

# Cold start strategies

## Who this is for

- Product and data stakeholders planning a pilot or rollout
- Engineers integrating `recsys-service` who need predictable fallbacks
- Recommendation engineers defining “good enough” behavior for new users/items

## What you will get

- A taxonomy of cold-start scenarios (new user, new item, new surface)
- Practical strategies that work with today’s RecSys suite capabilities
- A recommended fallback ladder (most personalized → least personalized)
- “Gotchas” that commonly cause empty or low-quality results

## Cold-start scenarios

Cold start usually means “we don’t have a strong signal yet”. In this suite, it often shows up as:

- **New user / guest**: no interaction history and no explicit anchors.
- **Sparse-history user**: a few events; personalization signals are noisy.
- **New item**: exists in the catalog but has little/no interaction signal.
- **New surface (namespace)**: you have not seeded signals for a `surface` yet.

## Baseline behavior in this suite

- In **DB-only mode**, the service reads signals from Postgres (at minimum `item_popularity_daily` and `item_tags`). If
  `item_popularity_daily` has no rows for the surface namespace, you should expect **empty recs** (see the runbook).
- Candidate sources are **opportunistic**: when a signal/store is missing, the service still returns results when it
  can, but emits warnings like `SIGNAL_UNAVAILABLE` / `SIGNAL_PARTIAL`.
- **Rule pins can inject items** that were not in the candidate pool (useful for curated cold-start defaults).
- `segment` defaults to `default` when omitted. Segments are used to scope rules and to slice evaluation.

## Strategy 1: Catalog-only (curated defaults via rules)

If you have a catalog but not enough interaction data yet, start with a curated “starter set”:

- Add segment-scoped pin rules for `segment=guest` / `segment=new_user`.
- Roll pins forward/back by updating rules (versioned + cacheable).

Minimal example (pin two items for guest users on `home`):

```json
[
  {
    "action": "pin",
    "target_type": "item",
    "item_ids": ["item_101", "item_202"],
    "surface": "home",
    "segment": "guest",
    "priority": 100
  }
]
```

Notes:

- Pin rules can inject items that are not in the candidate pool, but constraints/caps may still filter them if you use
  tag-based constraints.
- If you need tighter control over how many pins a rule can place, set `max_pins` on the rule.

## Strategy 2: Popularity priors (bootstrap new items and new surfaces)

Popularity is the simplest reliable fallback, but new items won’t show up until they have signal. You can bootstrap
them using a **prior**:

- When an item is created, write a small initial score into `item_popularity_daily` for “today”.
- Keep priors small so they don’t dominate real popularity, and let the configured half-life decay them naturally.

This is also how you avoid “new surface cold start”:

- Seed a minimal popularity table for the new surface namespace.
- If you intentionally want a cross-surface fallback, use the `default` namespace fallback described in
  [`explanation/surface-namespaces.md`](surface-namespaces.md).

## Strategy 3: Segment defaults (different policies per cohort)

Segments are a lightweight way to make cold-start behavior explicit without changing surfaces:

- Set `segment` in requests (examples: `guest`, `new_user`, `returning`).
- Use segment-scoped rules to pin/boost/block items differently per cohort.
- Use `segment` in `recsys-eval` to slice metrics (“does cold-start improve without harming returning users?”).

## Strategy 4: A fallback ladder (what to try, in order)

Treat cold start as a **fallback ladder**: try the most specific signal you have, then degrade gracefully.

Recommended ladder:

1. **Contextual anchors**: if you can provide `anchors.item_ids` (for example, the PDP item), do so.
2. **Co-visitation** (when available): similar-by-context, even for new users (anchors don’t require user history).
3. **Popularity**: surface-level trending / frequently interacted.
4. **Curated pins**: rules-based starter set per surface/segment.
5. **Application-level fallback**: if the API returns empty, render a safe default and log that it happened.

## Common cold-start failure modes (and fixes)

- **Empty results because the surface has no popularity rows**
  - Fix: seed `item_popularity_daily` for the surface namespace (or intentionally rely on `default` fallback).
  - Runbook: [`operations/runbooks/empty-recs.md`](../operations/runbooks/empty-recs.md)
- **Overly strict allow-lists**
  - Symptom: `CANDIDATES_INCLUDE_EMPTY`.
  - Fix: prefer `anchors.item_ids` for seeding; use `candidates.include_ids` only when you mean “only these items”.
- **Constraints filtering everything**
  - Symptom: `CONSTRAINTS_FILTERED`.
  - Fix: ensure `item_tags` exists for the same surface namespace and relax constraints during cold start.

## Read next

- Candidate vs ranking (controls order and warnings): [`explanation/candidate-vs-ranking.md`](candidate-vs-ranking.md)
- Admin API (rules scoping by segment): [`reference/api/admin.md`](../reference/api/admin.md)
- Minimal pilot (DB-only): [`tutorials/minimal-pilot-db-only.md`](../tutorials/minimal-pilot-db-only.md)
