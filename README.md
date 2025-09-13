# RecSys: A Recommendation Service

A domain-agnostic recommendation API. You send opaque IDs (no PII) for
users, items, and events. The service returns top-K recommendations and
"similar items." Tenants can customize event types and their weights
(e.g., view/click/purchase).

- Safe defaults
- Multi-tenant by design
- Works for products, content, listings, etc.

---

## What This Service Does

- **Trending / Popular now** with time decay: recent, important events
  push items up.
- **"People who engaged with X also like Y"** using co-visitation.
- **"Show me items like this"** using semantic similarity (embeddings).
- **Light personalization** (optional) from a user’s recent tags.
- **Diversity & caps** (optional) to avoid showing too many items from
  one brand/category.
- **Blended scoring** decides how high each candidate should rank, and the
  re‑ranker (MMR + caps) decides which of the high scorers make the final top‑K.

---

## How It Works (Big Picture)

### The Ranking Pipeline

1) candidates (popularity) -> 2) signals -> 3) normalize -> 4) blend
-> 5) personalize -> 6) re‑rank (MMR + caps) -> 7) reasons.

1. **Build candidates** from time-decayed popularity.
2. **Compute extra signals** per candidate:
   - Co-visitation vs. the user’s recent "anchor" items
   - Embedding similarity vs. those same anchors
3. **Normalize** each signal to `[0, 1]`, then **blend** them with
   weights `alpha`, `beta`, and `gamma`.
4. **Personalize** (optional): small multiplicative boost if the
   user’s tag profile overlaps with the item’s tags.
5. **Enforce diversity** (optional): MMR re-ranking plus brand/category
   caps.
6. **Return reasons** so you can see why items ranked (e.g., popularity,
   co-visitation, embeddings, personalization, diversity).

### Data + Algorithm Relationships

#### Entities (DB vs derived vs request-time)

```plaintext
[DB] Items        : id, tags, metadata
[DB] Events       : t, user_id, item_id, type
[DB] Embeddings   : item_id, vector
[DB] Tenancy      : org_id, namespace

[Derived] Popularity      : item_id -> score(t-decay + weights)
[Derived] CoVis Graph     : (item_id, item_id) -> weight
[Derived] UserTagProfile  : user_id -> {tag: weight}

[Request-time] Anchors    : recent item_ids for this user
[Request-time] Candidates : item_ids (≥ max(K, FANOUT))
[Request-time] Signals    : {pop_norm, co_vis_norm, embed_norm, …}
[Request-time] Score      : alpha*pop + beta*co + gamma*embed
[Request-time] Re-rank    : MMR + caps → Top-K
[Response]     Reasons    : ["popularity","co_visitation","embedding",...]
```

#### Flow

```plaintext
Events ──> Popularity ─┐
                       ├──> Candidates ────────────────┐
User  ──> Anchors ─────┘                               │
Anchors + CoVisGraph ───────────> CoVis signal ────────┤
Anchors + Embeddings ───────────> Embed signal ────────┤
                                     Normalize + Blend ├──> MMR + Caps ──> Top-K
UserTagProfile ────────────────> Light personalization ┘
```

---

## Pseudocode (Simplified)

```text
function recommend(req):
  k = req.k or 20

  # 1) Popularity candidates
  fetchK = max(k, POPULARITY_FANOUT)
  pop = popularityTopK(
    org, ns,
    halfLife = POPULARITY_HALFLIFE_DAYS,
    k = fetchK,
    constraints = req.constraints
  )

  # Optional: exclude items the user just bought
  exclude = req.constraints.exclude_ids
  if RULE_EXCLUDE_PURCHASED and req.user_id:
    since = now - days(PURCHASED_WINDOW_DAYS)
    exclude += listUserPurchasedSince(org, ns, req.user_id, since)

  cand = pop minus exclude
  meta = listItemsMeta(org, ns, ids(cand))

  # 2) Signals from user anchors (need user + recent activity)
  cooc, emb = zeros(), zeros()
  if req.user_id and (BLEND_BETA > 0 or BLEND_GAMMA > 0):
    since   = now - days(COVIS_WINDOW_DAYS)
    anchors = listUserRecentItemIDs(org, ns, req.user_id, since, limit=10)
    if BLEND_BETA > 0:
      cooc = coVisStrengthForCandidates(anchors, cand, since)
    if BLEND_GAMMA > 0:
      emb = maxEmbeddingSimForCandidates(anchors, cand)

  # 3) Normalize and blend
  for each item in cand:
    popN = normalize(pop[item])
    coN  = normalize(cooc[item])
    emN  = normalize(emb[item])
    score[item] = alpha*popN + beta*coN + gamma*emN
  if alpha == 0 and beta == 0 and gamma == 0:
    score = popN; alpha = 1

  ranked = sortBy(score)

  # 4) Light personalization
  if PROFILE_BOOST > 0 and req.user_id:
    window = PROFILE_WINDOW_DAYS or POPULARITY_WINDOW_DAYS
    profile = buildUserTagProfile(org, ns, req.user_id, window, PROFILE_TOP_N)
    for item in cand:
      overlap = sum(profile[tag] for tag in meta[item].tags)
      score[item] *= (1 + PROFILE_BOOST * overlap)
    ranked = sortBy(score)

  # 5) Diversity & caps
  if MMR_LAMBDA > 0 or BRAND_CAP > 0 or CATEGORY_CAP > 0:
    ranked = mmrReRank(
      ranked, meta, k,
      lambda = MMR_LAMBDA,
      brandCap = BRAND_CAP,
      categoryCap = CATEGORY_CAP
    )

  # 6) Done
  return topKWithReasons(ranked, k)
```

## Detailed Pseudocode

### Inputs and data shapes

#### Request (example)

```json
{
  "org_id": "org-123",
  "namespace": "shop",
  "user_id": "u-42",
  "k": 3,
  "constraints": { "exclude_ids": ["C"] }
}
```

#### Events (stored)

```json
[
  { "t": "2025-09-07T10:00Z", "user_id": "u-11", "item_id": "A", "type": "view" },
  { "t": "2025-09-08T13:00Z", "user_id": "u-11", "item_id": "B", "type": "purchase" },
  { "t": "2025-09-08T18:00Z", "user_id": "u-42", "item_id": "X", "type": "view" },
  { "t": "2025-09-09T08:00Z", "user_id": "u-42", "item_id": "Y", "type": "purchase" }
]
```

#### Items (stored)

```json
[
  { "id": "A", "tags": ["brand:NOVA", "cat:sneaker"] },
  { "id": "B", "tags": ["brand:NOVA", "cat:sneaker"] },
  { "id": "C", "tags": ["brand:ELMO", "cat:boot"] }
]
```

#### Embeddings (stored)

```plaintext
vector("A"), vector("B"), vector("C"), vector("X"), vector("Y")
```

### Step‑by‑step

#### 1) Build the candidate list from popularity

- Query recent events within POPULARITY_WINDOW_DAYS.
- Apply time decay with POPULARITY_HALFLIFE_DAYS.
- Sum per item to get "raw popularity."
- Keep the top POPULARITY_FANOUT items (at least K).

#### Example outcome from building candidate list

```plaintext
raw_pop = { A: 8.4, B: 5.1, C: 2.3, ... }
candidates = [A, B, C, ...]
```

#### 2) Apply business rules and fetch metadata

- Remove explicit excludes (`constraints.exclude_ids`, or "exclude purchased"
  if enabled).
- Fetch item tags/metadata for the survivors.

#### Example outcome from business rules

```plaintext
candidates = [A, B] # C was excluded by constraints
meta[A].tags = ["brand:NOVA","cat:sneaker"]
meta[B].tags = ["brand:NOVA","cat:sneaker"]
```

#### 3) Gather user anchors (if user_id present)

- Look up the user’s most recent items within COVIS_WINDOW_DAYS.
- These anchors give context for co‑visitation and embeddings.

#### Example anchors

```plaintext
anchors = ["X","Y"] # from u-42's recent activity
```

#### 4) Compute per‑candidate signals

- **Popularity**: already have `raw_pop` from step 1.
- **Co‑visitation**: how often anchors co‑occurred with each candidate.
- **Embeddings**: cosine similarity between each candidate and the anchors
  (use the max or a small aggregate like mean or 95th percentile).

#### Example raw signals

```plaintext
pop_raw: { A: 8.4, B: 5.1 }
co_vis_raw (vs anchors X,Y): { A: 3, B: 1 }
embed_raw (max cosine vs anchors): { A: 0.62, B: 0.40 }
```

#### 5) Normalize signals to [0,1]

- Do a min‑max per signal over the current candidate set.
- If a signal is missing for an item, treat it as 0.
- If all values are equal, the normalized values become 1.0 for those items.

#### Example normalization

```plaintext
pop_norm:   A:1.00, B:0.00     (min=5.1, max=8.4)
co_vis_norm A:1.00, B:0.00     (min=1,   max=3)
embed_norm: A:1.00, B:0.00     (min=0.40,max=0.62)
```

#### 6) Blend the signals (the scoring rule)

```plaintext
score = alpha*pop_norm + beta*co_vis_norm + gamma*embed_norm
```

- If alpha=beta=gamma=0, fall back to popularity‑only with alpha=1.
- Missing signals contribute 0 and do not hurt an item.

#### Example with alpha=1.0, beta=0.1, gamma=0.1

```plaintext
A: 1.0*1.00 + 0.1*1.00 + 0.1*1.00 = 1.20
B: 1.0*0.00 + 0.1*0.00 + 0.1*0.00 = 0.00
```

#### 7) Light personalization (optional)

- Build the user’s decayed tag profile over PROFILE_WINDOW_DAYS (or fallback).
- Compute overlap between profile and item tags.
- Multiply the item’s score by (1 + PROFILE_BOOST * overlap).

#### Example

```plaintext
profile: { "brand:NOVA": 0.8, "cat:sneaker": 0.6 }
overlap(A) = 0.8 + 0.6 = 1.4
if PROFILE_BOOST = 0.2:
  A: 1.20 * (1 + 0.2*1.4) = 1.20 * 1.28 = 1.536
```

#### 8) Diversity re‑rank and caps (optional)

- Use MMR with parameter MMR_LAMBDA to balance "score" vs "be different from
  those already chosen."
- Enforce BRAND_CAP and CATEGORY_CAP during selection.
- Result is a final order and a truncated top‑K.

#### Example (narrative)

```plaintext
Pick 1st: A (highest score)
For 2nd: candidates are penalized if too similar to A; pick next best that
balances score and novelty. Caps can skip items that would break limits.
```

#### 9) Build the response with reasons

- For each returned item, include a compact reason vector such as:
  ["popularity","co_visitation","embedding","personalization","diversity"].

#### Example response

```json
[
  { "item_id": "A",
    "score": 1.536,
    "reasons": ["popularity","co_visitation","embedding","personalization"] },
  { "item_id": "B",
    "score": 0.000,
    "reasons": ["popularity"] }
]
```

### What happens for anonymous users?

- Steps 3–4 (anchors, co‑visitation, embeddings vs anchors) are skipped or
  produce zeros. The system still works using popularity and, optionally, MMR.
- Personalization is skipped (no user profile).

### What happens when embeddings are missing?

- `embed_raw` is absent -> `embed_norm = 0` -> no contribution to the blend.
- Co‑visitation can still add context if the user has recent anchors.
- Otherwise, popularity carries the result (still stable and explainable).

---

## Algorithms (Plain English)

### Time-decayed popularity

Each event adds to its item’s score. Recent, high-weight events count
more. A **half-life** controls how fast old events fade. Example:
"14-day half-life" means an event loses half its influence every 14
days. This is robust, fast, and easy to reason about.

### Co-visitation

"Users who engaged with X also engaged with Y" within a recent window.
This captures "viewed together" or "bought together" patterns and
seasonality. It powers `/items/{id}/similar` and also helps ranking
when a user has recent anchors.

### Semantic similarity (embeddings)

Each item has a vector fingerprint derived from its text (and/or
images). Similar meaning = similar vectors. We use cosine similarity to
find "neighbors." This is great for cold start: it works before you
have event data. If embeddings are missing, we fall back to co-vis.

### Light personalization

We build a simple, decayed profile of the user’s top tags (e.g., brands
or categories). If a candidate shares those tags, we apply a small,
controlled boost. This is designed to be gentle, not overwhelming.

### Diversity and caps

Maximal Marginal Relevance (MMR) trades off "more relevant" vs "more
diverse." You choose the trade-off with `MMR_LAMBDA`. Caps limit how
many items per brand/category make it to the final top-K.

---

## Configuration (Environment Variables)

Put these in your service environment (see your `.env.example` files).

### API service vars

| Variable                 | Type / Range | What it does                                  | Notes |
|--------------------------|--------------|-----------------------------------------------|-------|
| `API_PORT`               | int 1..65535 | API port inside container.                    |       |
| `DATABASE_URL`           | Postgres URL | Connection string for API and migrations.     |       |
| `ENV`                    | string       | Environment name (`dev`, `staging`, `prod`).  |       |
| `APP_DEBUG`              | bool         | Force debug logging if `true`.                |       |
| `ORG_ID`                 | UUID         | Fallback org when a request header is absent. |       |
| `MIGRATE_ON_START`       | bool         | Run database migrations on startup.           |       |
| `MIGRATIONS_DIR`         | string       | Directory containing migration files.         |       |
| `SWAGGER_HOST`           | string       | Host for Swagger documentation.               |       |
| `SWAGGER_SCHEMES`        | string       | URL schemes for Swagger (`https`).            |       |
| `CORS_ALLOWED_ORIGINS`   | string       | Comma-separated allowed CORS origins.         |       |
| `CORS_ALLOW_CREDENTIALS` | bool         | Allow credentials in CORS requests.           |       |

### Windows, decay, and candidate fan-out vars

| Variable                   | Type / Range | What it does                                 | Effect of higher / lower                  |
|----------------------------|--------------|----------------------------------------------|-------------------------------------------|
| `POPULARITY_HALFLIFE_DAYS` | float > 0    | How fast old events fade.                    | Smaller = favors recency; larger = memory |
| `POPULARITY_WINDOW_DAYS`   | float > 0    | Hard lookback for popularity events.         | Larger = more history, more work          |
| `COVIS_WINDOW_DAYS`        | float > 0    | Lookback for co-vis and user anchors.        | Larger = more seasonal signal             |
| `POPULARITY_FANOUT`        | int > 0      | How many popularity candidates to pre-fetch. | Larger = more choice, more DB work        |

### Diversity & business rule vars

| Variable                 | Type / Range   | What it does                                    | Notes                           |
|--------------------------|----------------|-------------------------------------------------|---------------------------------|
| `MMR_LAMBDA`             | float in [0,1] | MMR trade-off: 1.0 = relevance, 0.0 = diversity | Set `0` to disable              |
| `BRAND_CAP`              | int ≥ 0        | Max items per brand in the final top-K.         | `0` disables                    |
| `CATEGORY_CAP`           | int ≥ 0        | Max items per category in the final top-K.      | `0` disables                    |
| `RULE_EXCLUDE_PURCHASED` | bool           | Exclude items the user purchased recently.      | Requires `user_id`              |
| `PURCHASED_WINDOW_DAYS`  | float > 0      | Lookback for the exclude-purchased rule.        | Required if the rule is enabled |

### Light personalization vars

| Variable              | Type / Range      | What it does                          | Notes                              |
|-----------------------|-------------------|---------------------------------------|------------------------------------|
| `PROFILE_WINDOW_DAYS` | float > 0 or `-1` | Lookback for building user profile.   | `-1` uses `POPULARITY_WINDOW_DAYS` |
| `PROFILE_TOP_N`       | int > 0           | Keep only the strongest N tags.       | Higher N = broader, noisier        |
| `PROFILE_BOOST`       | float ≥ 0         | Strength of the multiplicative boost. | `0` disables personalization       |

### Blended scoring weight vars

We rescore each candidate using normalized signals:

```plaintext
final = alpha*pop_norm + beta*co_vis_norm + gamma*embed_norm
```

| Variable      | Type / Range | What it does                                | Notes                                     |
|---------------|--------------|---------------------------------------------|-------------------------------------------|
| `BLEND_ALPHA` | float ≥ 0    | Weight for normalized popularity.           | If all three are zero, we fallback to pop |
| `BLEND_BETA`  | float ≥ 0    | Weight for normalized co-vis strength.      | Needs user anchors                        |
| `BLEND_GAMMA` | float ≥ 0    | Weight for normalized embedding similarity. | Needs embeddings + user anchors           |

**Why normalize?** Raw signals live on different scales (counts,
decayed sums, cosine similarity). Normalizing to `[0, 1]` makes the
weights intuitive and the blend stable. Channels with no signal produce
0 and have no effect.

### Contextual Bandit

| Variable      | Type / Range | What it does                                       | Notes |
|---------------|--------------|----------------------------------------------------|-------|
| `BANDIT_ALGO` | string       | Multi-armed bandit algorithm (`thompson`, `ucb1`). |       |

---

## Tuning Cheat-Sheet

- Start with `alpha=1.0`, `beta=0.1`, `gamma=0.1`.
- Raise **beta** if you want more "also viewed/bought together."
- Raise **gamma** for cold start and meaning-based tilt.
- `MMR_LAMBDA=0.6` is a reasonable diversity starting point.
- Keep light personalization gentle: `PROFILE_BOOST` around `0.1–0.3`.
- Caps (`BRAND_CAP`, `CATEGORY_CAP`) enforce catalog variety.

---

## Tenancy

- Multi-tenant by `(org_id, namespace)`.
- IDs are opaque; keep user privacy on the client side.
- Tenant config via `/v1/event-types:upsert` and `/v1/event-types`.

---

## Demo UI Features

The demo UI includes several powerful features for testing and exploring the recommendation system:

### User Traits Editor

The demo now includes a comprehensive **User Traits Editor** integrated as an accordion within the "Seed Data" section that allows you to:

1. **Quick Preview**: See a summary of configured traits and their probabilities without opening the editor
2. **Configure Dynamic Traits**: Define custom trait keys (e.g., `plan`, `age_group`, `interests`) with:
   - **Include Probability**: Control how often each trait appears in generated users (0-1)
   - **Value Options**: Define multiple possible values for each trait
   - **Value Probabilities**: Set weighted probabilities for each value (e.g., 60% "free", 30% "plus", 10% "pro")

3. **Edit User Traits in Browser**
   - Select any generated user from a dropdown
   - View and edit their current traits
   - Add new traits or modify existing ones
   - Update user traits directly in the browser

4. **Accordion Interface**
   - Collapsible section to save space
   - Always-visible preview of current configuration
   - Easy toggle between collapsed and expanded states

5. **Example Trait Configurations**:

  ```json
  {
    "plan": {
      "probability": 1.0,
      "values": [
        {"value": "free", "probability": 0.6},
        {"value": "plus", "probability": 0.3},
        {"value": "pro", "probability": 0.1}
      ]
    },
    "age_group": {
      "probability": 0.8,
      "values": [
        {"value": "18-24", "probability": 0.2},
        {"value": "25-34", "probability": 0.3},
        {"value": "35-44", "probability": 0.25},
        {"value": "45-54", "probability": 0.15},
        {"value": "55+", "probability": 0.1}
      ]
    }
  }
  ```

### Dynamic User Generation

When seeding data, users are now generated with traits based on your configuration:

- Each trait has a chance to be included based on its probability
- Values are selected using weighted random selection
- Fallback to default `plan` trait if no configurations are provided

### Events per User Configuration

The seeding system now supports realistic event generation:

- **Min/Max Events per User**: Set a range of events each user will generate
- **Randomized Distribution**: Each user gets a random number of events between min and max
- **Consistent Events**: Set min=max for the same number of events per user
- **Realistic Patterns**: More closely mimics real user behavior with varied activity levels

### Usage

1. **Configure Traits**: Use the "User Traits Configuration" accordion in the "Seed Data" section to set up your desired trait configurations
2. **Preview Configuration**: See a quick summary of your trait setup without opening the accordion
3. **Seed Data**: Click "Seed Data" to generate users with your configured traits
4. **Edit Users**: Select generated users from the dropdown to view and edit their traits
5. **Test Recommendations**: Use the updated user data to test how traits affect recommendations

## Explore

- **Swagger UI**: open `/docs` when the service is running.
- **Makefile**: see targets for dev, tests, and migrations.
- **Demo UI**: Access the interactive demo with user traits editor at the demo UI URL.

---

## Glossary

**ALS**
The embedding-similarity signal used in the blended score (the "gamma" term).
It measures semantic similarity via cosine similarity of item embeddings to
anchors (recent user items or a specific item). It is not matrix factorization
here. Increasing ALS gives more weight to semantic similarity relative to
popularity and co‑visitation.

**Anchors**  
A user’s most recent items (within `COVIS_WINDOW_DAYS`) i.e. items the current
user touched recently (their context). We compare candidates against anchors
to compute co-visitation and embedding similarity.

**Blended scoring (Linear Blend of Normalized Signals)**  
The rule that turns multiple normalized signals into one score:
`alpha*pop_norm + beta*co_vis_norm + gamma*embed_norm`. If all weights
are zero, we fall back to popularity (`alpha=1`).

**Candidate**  
An item (an item_id) that is eligible to be shown to the user and under
consideration for ranking It is not an event. that enters the ranking pipeline.

Candidates usually come from time-decayed popularity (the "candidate pool")
before we add other signals. The pool is typically larger than K so later steps
have room to reorder and filter.

**Caps**  
Hard limits like `BRAND_CAP` and `CATEGORY_CAP` that prevent too many
items from the same brand/category in the final list.

**Cold start**  
When items or users have little or no event history. Embeddings and
popularity help here; co-visitation needs anchors/history.

**Constraints**  
Filters applied before or during ranking (e.g., explicit `exclude_ids`,
"exclude purchased," availability, tenant rules).

**Co-visitation**
"How often did people who touched X also touch Y (soon/nearby)?" We aggregate
this globally (per tenant) inside a recent window and then, at request time,
look up the edges for the user’s anchors.

**Embeddings**
Numeric vectors that represent item meaning; cosine similarity measures
closeness.

**Event**
Evidence about items (view, add-to-cart, purchase).

**Fan-out**  
How many popularity candidates we pull *before* re-ranking
(`POPULARITY_FANOUT`). Typically `>= K` so downstream steps have choice.

**Half-life**  
How quickly old events fade in popularity scoring. A 14-day half-life
means an event’s influence halves every 14 days.

**Item**
something you could recommend (e.g., product "A").

**Light personalization**
A small score boost for items that share tags with what this user has
recently engaged with. The engine builds a short-lived tag profile from the
user’s own events (using your event-type weights and half-lives), then, for
each candidate item, sums the overlap of its tags with that profile and
multiplies the item’s score by `1 + PROFILE_BOOST * overlap`. Set
`PROFILE_BOOST=0` to turn it off; `PROFILE_WINDOW_DAYS` and `PROFILE_TOP_N`
control how the profile is built. Items boosted this way get a
"personalization" reason in the response.

**MMR (Maximal Marginal Relevance)**  
A standard relevance-vs-diversity trade-off method. Used as a
re-ranking step that balances "high score" vs "be different from what
we already picked," controlled by `MMR_LAMBDA`.

**Normalization**  
Rescales each signal across the current candidate set to `[0, 1]` so
weights are comparable. Missing signals become `0` (neutral, not
harmful).

**Personalization (Light)**  
A small multiplicative boost based on overlap between the user’s decayed
tag profile and the candidate’s tags: `score *= (1 + PROFILE_BOOST *
overlap)`.

**Reasons**
A compact audit trail per returned item that explains *why* it ranked.
Typical entries include:

- `popularity` (time-decayed demand)
- `co_visitation` (co-occurred with user anchors)
- `embedding` (semantic similarity to anchors)
- `personalization` (boost from tag profile overlap)
- `diversity` (selected by MMR for novelty)
- `cap_brand` / `cap_category` (caps enforced during selection)
- `excluded_purchased` (item was filtered earlier due to rule)
Not every reason appears on every item; only the applicable ones.

**Signals**  
Independent evidence used to score a candidate (e.g., popularity,
co-visitation, embedding similarity). Each signal is computed per
candidate before blending.

**Top-K**
Return the K highest-scoring items after filters.

**UserTagProfile**
A lightweight map of "tag" -> weight that summarizes a user’s recent interests,
normalized to sum to 1
(e.g., `{"t:android": 0.6, "category:phone": 0.3, "brand:acme": 0.1}`). It’s
computed from the user’s past events by applying your per-type weights and time
decay (half-life), optionally limited to a recent window and capped to the top-N
tags for performance. No model training or embeddings are required; it’s just
decayed counts on item tags.

**Windows**  
Lookback durations for data (e.g., `POPULARITY_WINDOW_DAYS`,
`COVIS_WINDOW_DAYS`, `PROFILE_WINDOW_DAYS`). Larger windows mean more
history; smaller windows favor recency.
