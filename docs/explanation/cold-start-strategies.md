---
diataxis: explanation
tags:
  - explanation
  - ml
  - developer
  - recsys-algo
---
# Cold start strategies
This page explains Cold start strategies and how it fits into the RecSys suite.


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

## Mapping: scenario → approach → what you need

!!! note "Out-of-the-box vs extra artifacts"
    In DB-only mode, cold start is mostly handled by **popularity + rules**. Co-visitation and similarity require
    additional stores/artifacts (often produced by `recsys-pipelines` or your own jobs).

### New user / guest (no history)

Recommended:

- Popularity fallback + curated defaults (pin/boost rules) scoped by `segment`.

Requires:

- Seeded `item_popularity_daily` for the surface namespace.
- Rules for `segment=guest` / `segment=new_user` (pins/boosts/blocks).

### Sparse-history user (few events)

Recommended:

- Treat as cold start until you have enough events; use the fallback ladder and keep personalization conservative.

Requires:

- Enough joined exposure/outcome history to build a stable user profile.

### New item (no interactions yet)

Recommended:

- Seed a small popularity prior and/or pin/boost during launch so the item can be discovered.

Requires:

- Item in the catalog and tags (if you use tag constraints).
- Optional: a few `item_popularity_daily` rows to give it an initial score.

### New surface (namespace)

Recommended:

- Seed popularity for the new surface namespace; add segment defaults for guest/new_user cohorts.

Requires:

- Surface configured in admin.
- Seeded `item_popularity_daily` (and `item_tags` if constraints rely on tags).

### Anchor-based surfaces (PDP “similar items”, contextual widgets)

Recommended:

- Always send `anchors.item_ids` and use co-visitation/similarity signals when available.

Requires:

- An anchor item ID in the request (you already have this on PDP).
- Co-occurrence / embedding / collaborative stores (see the ranking reference for what each signal needs).

If you need a precise catalog of what’s implemented and what each signal requires, see:
[Ranking & constraints reference](../recsys-algo/ranking-reference.md).

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
  [Surface namespaces](surface-namespaces.md).

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
  - Runbook: [Runbook: Empty recs](../operations/runbooks/empty-recs.md)
- **Overly strict allow-lists**
  - Symptom: `CANDIDATES_INCLUDE_EMPTY`.
  - Fix: prefer `anchors.item_ids` for seeding; use `candidates.include_ids` only when you mean “only these items”.
- **Constraints filtering everything**
  - Symptom: `CONSTRAINTS_FILTERED`.
  - Fix: ensure `item_tags` exists for the same surface namespace and relax constraints during cold start.

## Read next

- Candidate vs ranking (controls order and warnings): [Candidate generation vs ranking](candidate-vs-ranking.md)
- Admin API (rules scoping by segment): [Admin API + local bootstrap (recsys-service)](../reference/api/admin.md)
- Minimal pilot (DB-only): [minimal pilot mode (DB-only, popularity baseline)](../tutorials/minimal-pilot-db-only.md)
