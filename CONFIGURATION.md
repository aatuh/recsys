# RecSys Configuration & Data Ingestion Guide

This guide shows you how to configure the RecSys API and structure your data to get high‑quality recommendations across common business scenarios. It’s written in plain English, with copy‑paste payloads and a clear mapping between environment variables and per-request overrides.

> API base (demo): `https://recsys-production.up.railway.app`

---

## TL;DR (first things to set)
1) **Wire data**
- Upsert **items** (id, availability, price, tags, optional 384‑dim embedding).  
- Upsert **users** (id + optional traits).  
- Batch **events** (user_id, item_id, type, value, timestamp in RFC3339).

2) **Set global defaults via env**
- Popularity half-life, co-visitation window, diversity (MMR), brand/category caps, exclude-purchased behavior, personalization knobs (profile window/top-N/boost), blend weights, and bandit holdout settings.

3) **Tune per request**
- Use `blend` and `overrides` in `/v1/recommendations` so each surface (Home, PDP, Cart, Email) gets the right mix without redeploys.

4) **(Optional) Rules, manual overrides, segments, bandits**
- Apply pin/boost/block guardrails, register ad-hoc boosts or suppressions, target different knob bundles per segment, or let bandits pick between policies automatically.

---

## 1) Mental model & data flow

**Ingestion → Signals → Blending → Personalization → Diversity/Caps → Rules → Response**

- **Signals**: popularity (time‑decayed events), co‑visitation (co‑occurrence in a window), and optional embedding similarity.  
- **Blending**: final score is a weighted mix of those signals (think α·pop + β·cooc + γ·embedding).  
- **Personalization**: a lightweight user profile built from recent interactions (e.g., tag distributions) boosts matching items.  
- **Diversity & caps**: Maximal Marginal Relevance (MMR) plus brand/category caps prevent monotony.  
- **Rules**: pin/boost/block ops run near the end so business guardrails win when needed.

---

## 2) Environment variables (global defaults)

Set these in your `.env` before starting the service. Invalid values are rejected at boot with a clear error.

### Windows & decay
- `POPULARITY_HALFLIFE_DAYS` (float > 0) — recency vs memory for popularity (default 4).
- `COVIS_WINDOW_DAYS` (float > 0) — event lookback window for co-visitation (default 28).
- `POPULARITY_FANOUT` (int > 0) — how many popularity candidates to prefetch (default 500).

### Diversity & business constraints
- `MMR_LAMBDA` ([0,1]) — 1 = pure relevance, 0 = pure diversity (MMR off if 0).
- `BRAND_CAP`, `CATEGORY_CAP` (int ≥ 0) — limit repeats per brand/category in the final list.
- `BRAND_TAG_PREFIXES`, `CATEGORY_TAG_PREFIXES` (csv) — how the engine detects brand/category from `tags` (e.g., `brand`, `category,cat`).

### Exclude‑purchased & personalization
- `RULE_EXCLUDE_EVENTS` (bool) — remove items the user already interacted with (e.g., purchases).
- `EXCLUDE_EVENT_TYPES` (csv of small ints) — which event types to consider “excludable” (e.g., 3 = purchase).
- `PURCHASED_WINDOW_DAYS` (float > 0) — lookback window for exclude‑purchased.
- `PROFILE_WINDOW_DAYS` (float > 0 or -1) — how far back to build user profiles; -1 = all time.
- `PROFILE_TOP_N` (int > 0) — number of top profile features (e.g., tags) to keep.
- `PROFILE_BOOST` (float ≥ 0) — strength of profile-based boosting; 0 disables personalization (default 0.7 for balanced lift).

### Blending defaults
- `BLEND_ALPHA`, `BLEND_BETA`, `BLEND_GAMMA` — default weights for popularity/co-vis/embedding. Current recommended defaults: 0.25 / 0.35 / 0.40 (front-load embeddings for novel items).

### Bandit experiment controls
- `BANDIT_ALGO` — online policy selector (`thompson` or `ucb1`).
- `BANDIT_EXPERIMENT_ENABLED` — toggles exploration holdout for live testing.
- `BANDIT_EXPERIMENT_HOLDOUT_PERCENT` — fraction (0–1) of traffic routed to control.
- `BANDIT_EXPERIMENT_SURFACES` — CSV of surfaces that participate in the experiment (`home,cart` by default).
- `BANDIT_EXPERIMENT_LABEL` — human-readable tag used in logs and dashboards.

- ### Operational
- `ORG_ID`, `DATABASE_URL`, `API_PORT` — required infra basics. The same `ORG_ID`
  must be sent by clients in the `X-Org-ID` header (the shop sample reads
  `RECSYS_ORG_ID` and injects it automatically).
- Optional tracing/”explain” toggles for debugging and dashboards (implementation dependent).

> Tip: keep a checked‑in `.env.example` describing each variable and safe defaults.

---

## 3) The *same knobs*, per request (no redeploy)

Every call to `/v1/recommendations` can override the env defaults:

- `blend`: `{ "pop": α, "cooc": β, "als": γ }` (als = embedding similarity weight).  
- `overrides`: mirrors the key env vars so you can adjust windows, diversity, caps, fanout, exclude‑purchased, and personalization.

**Schema highlights** (representative; check your Swagger for exact fields)

```jsonc
// POST /v1/recommendations
{
  "user_id": "u_123",
  "namespace": "default",
  "k": 20,
  "include_reasons": true,
  "explain_level": "tags | numeric | full",
  "blend": { "pop": 0.7, "cooc": 0.2, "als": 0.1 },
  "overrides": {
    "mmr_lambda": 0.4,
    "brand_cap": 3,
    "category_cap": 5,
    "rule_exclude_events": true,
    "purchased_window_days": 365,
    "profile_boost": 0.25,
    "profile_window_days": 14,
    "profile_top_n": 32,
    "popularity_halflife_days": 14,
    "covis_window_days": 30,
    "popularity_fanout": 200
  },
  "constraints": {
    "include_tags_any": ["category:toys"],
    "exclude_item_ids": ["sku_456"],
    "price_between": [10, 50],
    "created_after": "2025-06-01T00:00:00Z" // RFC3339
  }
}
```

---

## 4) Ingest the data (copy‑paste payloads)

### 4.1 Items

**Endpoint**: `POST /v1/items:upsert`  
**Shape** (representative):
- `item_id` (string, required)
- `available` (bool, required for serving)
- `price` (number, optional)
- `tags` (string[], optional; use prefixes like `brand:acme`, `category:shoes`)
- `props` (object, optional; any custom JSON like titles, attributes)
- `embedding` (float[] length **384**, optional; required for embedding similarity)

```json
POST /v1/items:upsert
{
  "namespace": "default",
  "items": [
    {
      "item_id": "sku_123",
      "available": true,
      "price": 19.99,
      "tags": ["brand:acme", "category:toys", "color:red"],
      "props": { "title": "Acme Racer", "stock": 42 },
      "embedding": [0.0, 0.1, 0.0, /* ... 381 more ... */ 0.0]
    }
  ]
}
```

**Why tags matter**: brand/category caps and some similarity heuristics depend on your tag prefixes. Set `BRAND_TAG_PREFIXES` and `CATEGORY_TAG_PREFIXES` to match whatever you put before the colon.

---

### 4.2 Users

**Endpoint**: `POST /v1/users:upsert`  
**Shape**:
- `user_id` (string, required)
- `traits` (object, optional; e.g., `{ "loyalty_tier": "gold" }`)

```json
POST /v1/users:upsert
{
  "namespace": "default",
  "users": [
    { "user_id": "u_123", "traits": { "loyalty_tier": "gold" } }
  ]
}
```

---

### 4.3 Events

**Endpoint**: `POST /v1/events:batch` (returns 202 Accepted on enqueue)  
**Shape**:
- `user_id`, `item_id` (strings)
- `type` (small int; e.g., 0=view, 1=click, 2=add, 3=purchase, 4=custom)
- `value` (float; e.g., 1 for a single event, or price/qty for purchases)
- `ts` (RFC3339 string; optional; server fills if omitted)
- `meta` (optional object), `source_event_id` (optional idempotency)

```json
POST /v1/events:batch
{
  "namespace": "default",
  "events": [
    { "user_id": "u_123", "item_id": "sku_123", "type": 0, "value": 1, "ts": "2025-09-07T12:34:56Z" },
    { "user_id": "u_123", "item_id": "sku_456", "type": 3, "value": 1 }
  ]
}
```

**Event-type weights (optional but powerful)**  
Endpoints: `POST /v1/event-types:upsert`, `GET /v1/event-types?namespace=default`  
- Assign relative `weight` per type (e.g., purchase >> view).
- You can also set `half_life_days` per type if supported by your build.

```json
POST /v1/event-types:upsert
{
  "namespace": "default",
  "types": [
    { "type": 0, "name": "view",     "weight": 0.1 },
    { "type": 3, "name": "purchase", "weight": 1.0, "half_life_days": 365 }
  ]
}
```

---

## 5) Asking for recommendations

**Basic**
```json
POST /v1/recommendations
{ "user_id": "u_123", "namespace": "default", "k": 20 }
```

**With filters and explanations**
```json
POST /v1/recommendations
{
  "user_id": "u_123",
  "namespace": "default",
  "k": 20,
  "include_reasons": true,
  "explain_level": "tags",
  "constraints": {
    "include_tags_any": ["category:toys"],
    "price_between": [10, 50],
    "created_after": "2025-06-01T00:00:00Z"
  }
}
```

**Blending examples**
- PDP “related items”: favor co‑vis + embedding: `"blend": {"pop":0.1,"cooc":0.6,"als":0.3}`  
- New user: rely on popularity + category filters: `"blend":{"pop":1,"cooc":0,"als":0}` + `include_tags_any`

**Similar items sanity‑check**  
If exposed by your build: `GET /v1/items/{id}/similar?namespace=default&k=20` uses embedding neighbors and/or co‑visitation to show nearest neighbors for a given anchor item.

---

## 6) Common business setups

### A) E‑commerce Home (diverse, fresh, safe)
- **Env defaults**: `POPULARITY_HALFLIFE_DAYS≈14`, `COVIS_WINDOW_DAYS≈30`, `MMR_LAMBDA≈0.3–0.5`, `BRAND_CAP=3`, `CATEGORY_CAP=5`.
- **Per request**: `overrides.rule_exclude_events=true`, `purchased_window_days=365`, `include_reasons=true` for transparency.

### B) PDP “Similar items”
- **Blend**: shift toward co‑vis and embedding: `{"pop":0.1,"cooc":0.6,"als":0.3}`.
- **Filters**: keep category/price near the PDP item if needed via `constraints`.

### C) Logged‑in personalization
- **Env**: `PROFILE_WINDOW_DAYS=14`, `PROFILE_TOP_N=32`, `PROFILE_BOOST=0.25`.
- **Per request**: experiment with `overrides.profile_*` to tune influence without redeploys.

### D) Editorial/Compliance guardrails
- **Rules**: enable pin/boost/block endpoints; dry‑run before activating.
- **Segments**: attach different profiles (knob sets) to cohorts or surfaces.

### E) Policy selection & A/B via bandits
- Define several policy bundles (different blends, MMR, caps, personalization).  
- Call `/v1/bandit/recommendations` using Thompson sampling or UCB1 to auto-select a policy online.

### F) Merchandising overrides & promo pushes
- Use `POST /v1/admin/manual_overrides` to register a boost or suppression for a specific item, namespace, and surface. Include optional expiry and notes so the override auto-expires and remains auditable.
- List current overrides with `GET /v1/admin/manual_overrides?namespace=...` and cancel with `POST /v1/admin/manual_overrides/{override_id}/cancel` once the campaign ends.
- Overrides write through the rule engine, so they cooperate with caps, segments, and decision tracing automatically.

---

## 7) Quality tuning cheat‑sheet (env ↔ request mapping)

| Goal                        | Env var                                                 | Per‑request override                                                                  |
|-----------------------------|---------------------------------------------------------|---------------------------------------------------------------------------------------|
| Favor recency in popularity | `POPULARITY_HALFLIFE_DAYS`                              | `overrides.popularity_halflife_days`                                                  |
| Use longer co‑vis history   | `COVIS_WINDOW_DAYS`                                     | `overrides.covis_window_days`                                                         |
| More candidate breadth      | `POPULARITY_FANOUT`                                     | `overrides.popularity_fanout`                                                         |
| Stronger diversity          | lower `MMR_LAMBDA`                                      | `overrides.mmr_lambda`                                                                |
| Limit brand repetition      | `BRAND_CAP`                                             | `overrides.brand_cap`                                                                 |
| Limit category repetition   | `CATEGORY_CAP`                                          | `overrides.category_cap`                                                              |
| Exclude purchased items     | `RULE_EXCLUDE_EVENTS`, `PURCHASED_WINDOW_DAYS`          | `overrides.rule_exclude_events`, `overrides.purchased_window_days`                    |
| Adjust personalization      | `PROFILE_BOOST`, `PROFILE_WINDOW_DAYS`, `PROFILE_TOP_N` | `overrides.profile_boost`, `overrides.profile_window_days`, `overrides.profile_top_n` |
| Change signal mix           | `BLEND_ALPHA/BETA/GAMMA`                                | `blend.pop/cooc/als`                                                                  |

Overrides are applied before ranking; they temporarily replace the corresponding env defaults for that single call.

---

## 8) Segments & profiles (targeted knob sets)

- **Profiles**: pre‑defined bundles of knobs (blend, MMR, caps, personalization, windows, fanout, tag prefixes).
- **Segments**: user or surface cohorts (e.g., “new users”, “PDP”, “email”) mapped to a profile.
- **Workflow**: upsert profiles → upsert segments → associate → dry‑run → activate.

---

## 9) Bandits (automatic policy selection)

- **Policies**: named knob bundles; think “A=diverse”, “B=more co‑vis”, “C=embedding‑heavy”.
- **Serve**: `/v1/bandit/recommendations` returns both the chosen policy id and the items.
- **Algorithms**: Thompson sampling or UCB1 (choose in config).
- **Shop safety checks**: the demo shop pings `/v1/bandit/policies` on startup. If the configured policy IDs are missing (or the lookup fails), it disables exploration automatically and falls back to `/v1/recommendations`, emitting a single warning instead of repeated 500s.
- **Profiles**: Algorithm weights live in the `RecommendationProfile` table. The shop admin UI (or `/api/admin/recommendation-settings`) lets you create profiles, mark defaults per surface, and clients can request one via the `profileId` query parameter; inline overrides still win if both are supplied.

---

## 10) Ops & observability

- **Explanations / decision traces**: enable to capture input config, anchors, scoring reasons, MMR decisions, and rule applications per request.
- **Data management**: list/delete endpoints for users/items/events by namespace/time range for backfills and cleanup.
- **Manual overrides**: the admin endpoints emit audit metadata (`created_by`, `cancelled_by`, timestamps). Monitor override counts and stale entries; expired overrides are auto-marked and can be cleaned regularly.
- **Catalog freshness**: run `make catalog-backfill` for the initial metadata/embedding backfill, and `make catalog-refresh SINCE=24h` (or your preferred window) on a schedule to keep new products enriched.
- **SLOs**: track p99 latency, 5xx rate, and recommendation “fill rate”. Add counters for rule hits, override usage, and cap-induced re-ranks.

---

## 11) Pitfalls & how to avoid them

- **Embedding dimension mismatch**: embeddings must be exactly **384** dims; otherwise item upsert fails.
- **Bad timestamps**: `created_after` and event `ts` must be RFC3339 (e.g., `2025-06-01T00:00:00Z`).
- **Missing tag prefixes**: brand/category caps won’t apply unless you set `BRAND_TAG_PREFIXES` / `CATEGORY_TAG_PREFIXES` to match your `tags`.
- **Zero/negative event weights**: keep weights positive; otherwise the upsert may fail validation.
- **Unavailable anchors**: items you interacted with can guide co‑vis/embedding even if they’re no longer available; that’s OK—they just won’t be served.

---

## 12) Minimal “from zero to recs” sequence

1) **Items**
```json
POST /v1/items:upsert
{ "namespace":"default", "items":[
  {"item_id":"A","available":true,"tags":["brand:acme","category:slots"]},
  {"item_id":"B","available":true}
]}
```

2) **Users**
```json
POST /v1/users:upsert
{ "namespace":"default", "users":[ {"user_id":"u1"} ] }
```

3) **Events**
```json
POST /v1/events:batch
{ "namespace":"default",
  "events":[ {"user_id":"u1","item_id":"A","type":0,"value":1},
             {"user_id":"u1","item_id":"B","type":3,"value":1} ] }
```

4) **Recommendations (with reasons)**
```json
POST /v1/recommendations
{ "user_id":"u1","namespace":"default","k":5,"include_reasons":true }
```

---

## 13) When to use which signal (intuition)

- **Popularity**: robust and cold‑start friendly; shorten the half‑life to catch trends faster.
- **Co‑visitation**: strongest for PDP “similar”, bundles, and “often bought with”. Tune the co‑vis window.
- **Embeddings**: best when you have rich text/image features; keep γ modest at first so it complements (not replaces) popularity.

---

## 14) Next steps (actionable checklist)

1) Fill `.env` with sane defaults, e.g.:  
   `POPULARITY_HALFLIFE_DAYS=14`  
   `COVIS_WINDOW_DAYS=30`  
   `MMR_LAMBDA=0.4`  
   `BRAND_CAP=3`  
   `CATEGORY_CAP=5`  
   `PROFILE_BOOST=0.25`  
   `PROFILE_WINDOW_DAYS=14`  
   `PROFILE_TOP_N=32`  
   `BRAND_TAG_PREFIXES=brand`  
   `CATEGORY_TAG_PREFIXES=category,cat`

2) Upsert items, users, and events using the payloads above (batch is fine).  
3) Call `/v1/recommendations` with surface‑specific `blend` and `overrides`.  
4) Add rules (pin/boost/block) where merchandising needs control; test via dry‑run.  
5) If you want automatic policy exploration, define policies and switch that surface to `/v1/bandit/recommendations`.

---

**Questions or a specific surface/KPI?** Tell me your surfaces (Home, PDP, Search, Email, Push) and KPIs (CTR, AOV, retention), and I’ll propose concrete values for `blend`, MMR, caps, and personalization tuned to your goals.
