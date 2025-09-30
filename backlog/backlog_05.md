# Segment Profiles — Implementation Plan (Deterministic, Auditable, Fast)

> Status: Phase 1 and Phase 2 implemented (Phase 3 optional / not required).

**Short answer:** Implement Segment Profiles as deterministic, rule-driven
"profile" selection at request time. A profile is just a named bundle of your
existing algorithm knobs (blend weights, diversity, caps, personalization,
windows). No external LLM at runtime. If you ever use an LLM, keep it optional
and offline to help author rules, not to decide segments.

---

## What a "Segment Profile" is

A named bundle of ranking parameters you already expose, for example:

* Blend weights: `alpha` (popularity), `beta` (co-visitation), `gamma`
  (embeddings)
* Diversity & caps: `mmr_lambda`, `brand_cap`, `category_cap`
* Light personalization: `profile_boost`, `profile_top_n`, `profile_window_days`
* Recency windows & rules: `half_life_days`, `co_vis_window_days`,
  `purchased_window_days`, `rule_exclude_purchased`

A profile is selected per request by matching a segment; then the engine ranks
with the profile’s knobs.

---

## How a segment is chosen

Deterministic rules on request context and user traits, for example:

* User tier: VIP, new, returning
* Jurisdiction / region
* Device, time of day, surface (home, lobby, detail)
* Simple behavior facts (e.g., "played slots in last 7 days")

Rules output a `segment_id`, which maps to a `profile_id`.

---

## Request-time precedence (who wins)

Make the effective config from these layers (left < right = higher priority):

`env defaults` → `namespace defaults` → `segment profile` →
`campaign overrides (optional)` → `per-request overrides`

This keeps behavior predictable and easy to reason about.

---

## Minimal data model (Postgres)

**segments**

* `namespace`, `segment_id`, `name`, `priority`, `active`, `profile_id`,
  `description`

**segment\_rules**

* `segment_id`, `rule_expr_json` (rule DSL), `enabled`

**profiles**

* `profile_id`, all the knobs listed above (alpha/beta/gamma, mmr\_lambda,
  caps, profile\_boost, windows, exclude\_purchased, etc.)

Optional:

* **segment\_audit\_log** for "who got what and why" (debug & compliance).

---

## Rule DSL (simple and auditable)

JSON of AND/OR over predicates. Example:

```json
{
  "any": [
    { "all": [
      { "eq": ["user.tier", "VIP"] },
      { "gte": ["user.ltv_eur", 500] }
    ]},
    { "all": [
      { "eq": ["ctx.region", "FI"] },
      { "in": ["ctx.surface", ["home", "casino"]] },
      { "gte_days_since": ["user.last_play_ts", 30] }
    ]}
  ]
}
```

Predicates read only the request context (user traits, geo) and cached facts
(e.g., `last_play_ts`). Keep it deterministic and cheap.

---

## Request pipeline (where it plugs in)

1. **Assemble context**: `(namespace, user_id, traits, geo, surface, time_of_day,
   session facts)`.
2. **Segment match**: Evaluate `segment_rules` by descending `priority`. First
   match wins; fallback to `segment_id = "default"`.
3. **Load profile**: Fetch `profiles.profile_id` for the selected segment.
4. **Build effective config**: Apply precedence chain
   (defaults → segment → overrides).
5. **Rank**: Run the existing pipeline (blend, personalization, MMR, caps,
   windows, rules) using the effective config.
6. **(Optional) Bandit**: Add `segment_id` as a bandit context feature so
   policy selection adapts per-segment.
7. **Explain**: Include `segment_id` and `profile_id` in the debug/"why"
   payload for traceability (e.g., "segment=VIP; profile=high-novelty").

---

## Admin UX (business-friendly)

* **Segments list**: name, rule summary, priority, profile, on/off.
* **Rule editor**: form + JSON view with validator and a dry-run tester
  ("Given context X, what segment?").
* **Profiles editor**: sliders/inputs for all knobs (reuse demo tuning UI).
* **Simulator**: pick a user/traits → see assigned segment → show the
  "effective config" and a preview recommendation list.

---

## Example profile (YAML)

```yaml
profile_id: vip-high-novelty
blend:
  alpha: 0.8
  beta: 0.2
  gamma: 0.4
diversity:
  mmr_lambda: 0.6
caps:
  brand_cap: 2
  category_cap: 3
personalization:
  profile_boost: 0.25
  profile_window_days: 30
  profile_top_n: 12
windows:
  half_life_days: 3
  co_vis_window_days: 14
rules:
  rule_exclude_purchased: true
  purchased_window_days: 7
```

A "new-user" profile might reduce co-visitation (no anchors yet) and emphasize
embeddings (`gamma`) with modest popularity (`alpha`).

---

## Why not an LLM for segments?

* **Determinism & auditability**: Segments affect what users see. Keep decisions
  consistent and reviewable. LLM output is not deterministic.
* **Latency & cost**: Rule evaluation is microseconds; LLM calls add 100–800 ms
  and cost per request.
* **Privacy/regulation**: iGaming buyers want transparent, explainable, on-prem
  decisions.

**OK uses of an LLM (optional, offline):**

* Help product teams author readable rule descriptions that compile to the JSON
  DSL.
* Generate onboarding help text and documentation, not runtime decisions.

---

## Phased implementation

**Phase 1 (MVP)**

* Tables: `segments`, `segment_rules`, `profiles`.
* Request hook to evaluate rules and load a profile.
* Precedence merge into existing config.
* Add `segment_id` to response debug block.

**Phase 2 (Ops & UX)**

* Admin UI for segments/profiles + dry-run simulator.
* Metrics: counts by `segment_id`, CTR/CVR per segment (Prometheus labels).
* Audit: log `(user_hash, segment_id, profile_id, rule_id)` to aid investigations.

**Phase 3 (Optional ML)**

* Train a small classifier to propose segments (features: recency/frequency/
  monetary, top categories, device, geo). Use it only to propose; rules still
  decide or act as fallback.
* Feed `segment_id` into the bandit context so policy selection adapts per
  segment rather than globally.

---

## Testing & guardrails

* **Golden tests**: tiny fixtures asserting contexts map to `segment_id` and
  that effective knobs change ranking as expected.
* **Perf budget**: rule eval + profile fetch <= 1 ms p50; cache profiles by
  `profile_id` and segments by `(namespace, segment_id)`.
* **Safety**: if rule eval errors, fall back to the default segment; record a
  metric.
* **Explainability**: include `segment_id/profile_id` in admin-facing "why"
  views next to reason tags (diversity, personalization, popularity, etc.).

---

## TL;DR

Build Segment Profiles as a rules → profile-selection layer that chooses among
pre-defined bundles of knobs you already have. Keep it deterministic, auditable,
fast, and privacy-preserving. Use bandit/ML only as controlled, optional
enhancements. No external LLM in the runtime path.
