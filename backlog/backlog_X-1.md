# X-1 — Rule Engine v1 (Pin / Block / Boost + TTL)

> Status: **Proposed — small, shippable v1**
> Goal: Let product & ops steer results quickly with deterministic rules that are **auditable**, **TTL-bounded**, and **cheap** to evaluate.

---

## TL;DR

Introduce a minimal Rule Engine evaluated at request time. It supports three actions:

1. **BLOCK** candidates, 2) **PIN** items to the top, 3) **BOOST** item scores.
   Rules can be scoped (namespace/surface/segment), have **precedence** (priority), **TTL** (valid\_from/until), and produce **reason tags** + **audit entries**.

---

## User stories

* *As an operator*, I can pin N items for a campaign (this week only).
* *As compliance*, I can block items/brands/categories instantly.
* *As a merchandiser*, I can +10% boost items with tag “new” on home surface.
* *As an analyst*, I can see which rules fired and why.

---

## In scope (v1)

* Actions: **BLOCK, PIN, BOOST** (additive to blended score pre-MMR).
* Targets: `item_id` (single/list), `tag`, `brand`, `category`.
* Scope fields: `namespace`, `surface` (aka placement), optional `segment_id`.
* TTL: `valid_from`, `valid_until`, `enabled`.
* Precedence: `priority` (higher wins), deterministic conflict rules (below).
* **Explain tags** (`rule.block`, `rule.pin`, `rule.boost:<id>`) + **Audit**.
* CRUD + list + **dry-run** admin endpoints.
* In-memory cache (hot rules), DB as source of truth.

**Out of scope (v1):** weighted pins per position, regex targets, geo/device conditions, per-brand caps here (those remain in caps logic), scripts/LLM helpers.

---

## Data model (Postgres)

```sql
-- Rules live under a merchandising schema; adjust schema name as needed.
CREATE TYPE rule_action AS ENUM ('BLOCK','PIN','BOOST');
CREATE TYPE rule_target AS ENUM ('ITEM','TAG','BRAND','CATEGORY');

CREATE TABLE rules (
  rule_id        UUID PRIMARY KEY,
  namespace      TEXT NOT NULL,
  surface        TEXT NOT NULL,          -- e.g., "home", "gamepage"
  name           TEXT NOT NULL,
  description    TEXT,
  action         rule_action NOT NULL,
  target_type    rule_target NOT NULL,
  -- For ITEM target: item_ids populated; otherwise use target_key
  target_key     TEXT,                   -- e.g., tag=“new”, brand=“X”, category=“slots”
  item_ids       TEXT[],                 -- array of item IDs for PIN/BLOCK/BOOST
  boost_value    DOUBLE PRECISION,       -- additive to blended score (if BOOST)
  max_pins       INTEGER,                -- optional per-rule cap (v1 uses global default)
  segment_id     TEXT,                   -- optional segment gating
  priority       INTEGER NOT NULL DEFAULT 0,
  enabled        BOOLEAN NOT NULL DEFAULT TRUE,
  valid_from     TIMESTAMPTZ,
  valid_until    TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON rules (namespace, surface, enabled);
CREATE INDEX ON rules (valid_from, valid_until);
CREATE INDEX ON rules (segment_id);
```

> Note: Use `(now() BETWEEN valid_from AND valid_until OR TTL NULLs)` in queries.

---

## Evaluator (request-time)

**Context:** `{namespace, surface, segment_id?, now, candidate_ids, item_meta(tag,brand,category)}`

**Selection:**
`active_rules = WHERE namespace/surface match AND enabled AND TTL ok AND (segment_id IS NULL OR matches)`
Sort `active_rules` by `priority DESC, created_at ASC`.

**Effect application order (per item):**

1. **BLOCK** ⇒ remove from candidates (wins over everything).
2. **PIN** ⇒ collect into `pinned_items` (dedup), placed at top; bypass MMR/caps in v1 (configurable later).
3. **BOOST** ⇒ `score += boost_value` **before** MMR/diversity caps.

**Targets → items mapping:**

* `ITEM`: exact id(s)
* `TAG` / `BRAND` / `CATEGORY`: match via item meta
* v1 **does not inject** new candidates (except **PIN** can surface an item not in candidates); BOOST applies only if item already in candidates.

**Global guards:**

* `MAX_PIN_SLOTS_PER_RESPONSE` (config or per-rule `max_pins`, default 3)
* Dedup pins vs ranked list; pinned items are removed from the ranked pool.

---

## Precedence & conflicts

* **BLOCK > PIN > BOOST** (safety first).
* If multiple BOOST rules match the same item, **sum** their `boost_value`.
* If multiple PIN rules match, respect `priority` order, then truncate by slots.
* Record **all** matched rules in audit; in explain reasons keep dedupbed tags with details (rule ids, amounts).

---

## Explain & Audit integration

* **Explain reason tags** per affected item:

  * `rule.block` (with `[rule_id]`)
  * `rule.pin` (with `[rule_id]`)
  * `rule.boost:+0.12` (with `[rule_id]`)

* **DecisionTrace additions** (new blocks):

  * `rules_evaluated: [rule_id...]`
  * `rules_matched: [{rule_id, action, target, affected_item_ids[]}]`
  * `rule_effects_per_item: {item_id: {blocked:bool, pinned:bool, boost_delta:float}}`

---

## Admin HTTP (v1)

```
POST   /admin/rules           # create
PUT    /admin/rules/{id}      # update
GET    /admin/rules           # list (filters: namespace,surface,enabled,active_now,segment_id,action,target_type)
POST   /admin/rules/dry-run   # body: {namespace,surface,segment_id?, items:[id..]} → which rules would fire + item effects
```

**Validation:**

* `BOOST` requires `boost_value ≠ 0`.
* `PIN` requires `item_ids` non-empty.
* TTL check: `valid_until > valid_from` if both present.
* Unique safety: optional guard to prevent overlapping **BLOCK** and **PIN** on identical `(namespace,surface,item_id)`.

---

## Config (env)

* `RULES_CACHE_REFRESH=2s` (poll or NOTIFY), `RULES_MAX_PIN_SLOTS=3`
* `RULES_ENABLE=true` (kill-switch)
* `RULES_AUDIT_SAMPLE=1.0` (share existing audit sampler)

---

## Performance

* Hot rules cached by `(namespace,surface)` and optionally `(segment_id)` key.
* Evaluator target mapping is O(#active\_rules + #candidates).
* v1 budget: **< 0.5 ms p50** on 100 rules / 100 candidates.

---

## Testing (must-pass)

* **Unit**:

  * BLOCK removes candidates (even if also PIN/BOOST).
  * PIN places items at top, respects slot cap, dedups.
  * BOOST adds pre-MMR and changes ordering when ties occur.
  * TTL filters work around boundaries.
  * Priority resolves conflicts deterministically.

* **Golden**: request fixtures → final list & explain tags stable.

* **Audit**: matched rule ids & per-item effects present.

* **Dry-run**: returns same matched sets as live eval.

---

## Migration & rollout

1. Run DDL. 2) Deploy engine + admin endpoints with `RULES_ENABLE=false`.
2. Populate sample rules; verify via **dry-run**.
3. Enable for one surface; monitor **audit** & **latency**.
4. Expand surfaces; add authoring UI later.

---

## Example (YAML authoring → JSON payload)

```yaml
- name: Pin weekly heroes
  namespace: lottery
  surface: home
  action: PIN
  item_ids: ["L123","L456"]
  priority: 100
  valid_until: 2025-10-05T21:00:00Z

- name: Hide brand X on home
  namespace: icasino
  surface: home
  action: BLOCK
  target_type: BRAND
  target_key: "BrandX"
  priority: 90

- name: Boost new-tag +0.15
  namespace: icasino
  surface: home
  action: BOOST
  target_type: TAG
  target_key: "new"
  boost_value: 0.15
  priority: 10
```

---

## Acceptance criteria

* CRUD + dry-run endpoints work; validation enforced.
* Evaluator applies BLOCK > PIN > BOOST with TTL and priority.
* Pinned items appear first (up to cap); boosted items show higher pre-MMR score.
* Explain contains `rule.*` tags; audit records matched rules & effects.
* Latency impact within budget; feature guarded by `RULES_ENABLE`.
