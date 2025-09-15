# RecSys: A Recommendation Service

A domain-agnostic recommendation API. You send opaque IDs (no PII) for
users, items, and events. The service returns top-K recommendations and
"similar items." Tenants can customize event types and their weights
(e.g., view/click/purchase).

- Safe defaults
- Multi-tenant by design
- Works for products, content, listings, etc.

## What This Service Does

- **Trending / Popular now** with time decay: recent, important events
  push items up.
- **"People who engaged with X also like Y"** using co-visitation.
- **"Show me items like this"** using semantic similarity (embeddings).
- **Light personalization** (optional) from a user's recent tags.
- **Diversity & caps** (optional) to avoid showing too many items from
  one brand/category.
- **Blended scoring** decides how high each candidate should rank, and the
  re‑ranker (MMR + caps) decides which of the high scorers make the final top‑K.

## Algorithms Used

### Time-decayed popularity

Each event adds to its item's score. Recent, high-weight events count
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

We build a simple, decayed profile of the user's top tags (e.g., brands
or categories). If a candidate shares those tags, we apply a small,
controlled boost. This is designed to be gentle, not overwhelming.

### Diversity and caps

Maximal Marginal Relevance (MMR) trades off "more relevant" vs "more
diverse." You choose the trade-off with `MMR_LAMBDA`. Caps limit how
many items per brand/category make it to the final top-K.

## How It Works (Big Picture)

### The Ranking Pipeline

1) Candidates (popularity): from time-decayed popularity.
2) Signals per candidate.
   - Co-visitation vs. the user's recent "anchor" items.
   - Embedding similarity vs. those same anchors.
3) Normalize and blend.
   - Each signal to `[0, 1]`, then blend them with weights
  `alpha`, `beta`, and `gamma`.
4) Light personalization.
   - Small multiplicative boost if the user's tag profile overlaps with the
   item's tags.
5) Re‑rank (MMR + caps)
   - MMR re-ranking plus brand/category.
6) Output and reasons
   - See why items ranked (e.g., popularity, co-visitation, embeddings,
   personalization, diversity).

### Data + Algorithm Relationships

```plaintext
[DB] Items        : id, tags
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

## Detailed Recommendation Flow

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

#### Build the candidate list from popularity

- Doesn't take into account user information i.e. user ID is not used but other
  constraints and filters like namespace and organization ID are used.
- Apply time decay with POPULARITY_HALFLIFE_DAYS.
- Sum per item to get "raw popularity."
- Keep the top POPULARITY_FANOUT items (at least K).

Scoring formula used for single item to find candidates. Each score is summed
among all events with the same event ID ("raw popularity").

```plaintext
0.5 ^ ( age_seconds / (hl_days * 86400) ) * event_type_weight * event_value
```

- 86400 is one day in seconds.
- Half-life means: every hl_days, the contribution halves.
- Larger event weight or event value increases the contribution. Older
  timestamps decrease it exponentially.

#### Example outcome from building candidate list

```plaintext
raw_popularity = {
  A: 8.4,
  B: 5.1,
  C: 2.3,
}
```

#### Apply business rules and fetch tags

- Remove excluded results (e.g. recently purchased items, specific item IDs).
- Fetch item tags for the survivors.

#### Example outcome from business rules

```plaintext
candidates = [A, B] # C was excluded by constraints
tags = {
  A: {"brand": "NOVA", "cat": "sneaker"},
  B: {"brand": "NOVA", "cat": "sneaker"},
}
```

#### Gather user anchors (if user_id present)

- Look up the user's most recent items within COVIS_WINDOW_DAYS.
- These anchors give context for co‑visitation and embeddings.

#### Example anchors

```plaintext
anchors = ["X","Y"] # from u-42's recent activity
```

#### Compute per‑candidate signals

- **Popularity**: already have `raw_popularity` from step 1.
- **Co‑visitation**: how often anchors co‑occurred with each candidate.
- **Embeddings**: compute cosine similarity between each anchor and all
  items, then retain scores only for items already in the candidate pool.
  The candidate's embedding score is the maximum similarity across anchors.

#### Example raw signals

```plaintext
pop_raw: { A: 8.4, B: 5.1 }
co_vis_raw (vs anchors X,Y): { A: 3, B: 1 }
embed_raw (max cosine vs anchors): { A: 0.62, B: 0.40 }
```

#### Normalize signals to [0,1]

- Do a min‑max per signal over the current candidate set.
- If a signal is missing for an item, treat it as 0.
- If all values are equal, the normalized values become 1.0 for those items.

Example normalization:

```plaintext
pop_norm:   A:1.00, B:0.00     (min=5.1,  max=8.4)
co_vis_norm A:1.00, B:0.00     (min=1,    max=3)
embed_norm: A:1.00, B:0.00     (min=0.40, max=0.62)
```

#### Blend the signals (the scoring rule)

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

#### Light personalization (optional)

- Build a user tag profile over PROFILE_WINDOW_DAYS by summing each tag's
  time-decayed, type-weighted event contributions. Keep the top N tags.
- Normalize the profile so the weights sum to 1 (probability distribution).
- For each candidate item, compute overlap = sum of profile weights for the
  candidate's tags (in [0,1]).
- Multiply the candidate's score by (1 + PROFILE_BOOST * overlap).

Example personalization:

```plaintext
profile = {"brand:NOVA": 0.57, "cat:sneaker": 0.43}
overlap(item) = 0.57 (only brand matches)
PROFILE_BOOST = 0.2  => multiplier = 1 + 0.2*0.57 = 1.114
```

#### Diversity re‑rank and caps (optional)

- Use MMR with parameter `MMR_LAMBDA` to balance "score" vs "be different from
  those already chosen."
- The algorithm penalizes on max found similarity of already selected items vs
- evaluated candidate using the value of `λ` as weight.
- Enforce BRAND_CAP and CATEGORY_CAP during selection.
- Result is a final order and a truncated top‑K.

`plaintext
MMR(c) = λ * normScore(c.Score, maxScore) - (1-λ) * maxSim(c, Selected)
`

- `c` is the evaluated candidate.
- `λ` is the trade-off/weight parameter between relevance and diversity `[0,1]`.
  - `λ=1.0` = pure relevance, no diversity.
  - `λ=0.0` = pure diversity (far from selected), ignoring base scores.
  - Typical: `0.6..0.9` to prefer relevance but still spread items.
- `normScore(x, maxScore)` scales the candidate's base score into `[0,1]`.
- `maxSim(c, Selected)` is the maximum similarity between the candidate and any
  already-selected item.

Similarity between tag sets is done using Jaccard similarity, which returns
the similarity between tag sets within range `[0,1]`.

Tiny example (3 items, tags)

- Assume
  - `λ = 0.75`
  - Base scores: `[A: 100, B: 90, C: 85]`
  - Tags: `A: {red, leather}, B: {red, suede}, C: {blue, canvas}`
  - `maxScore = 100`, so normalized scores are `[1.00, 0.90, 0.85]`.

- Round 1 (`selected = {}`):
  - `MMR(A) = 0.75 * 1.00 - (1-0.75) * 0 = 0.75`
  - `MMR(B) = 0.75 * 0.90 - (1-0.75) * 0 = 0.675`
  - `MMR(C) = 0.75 * 0.85 - (1-0.75) * 0 = 0.6375`
  - Pick A.
  - Note: Nothing to input to Jaccard yet.

- Round 2 (`selected = {A}`):
  - `Jaccard(B, A) = |{red}|/|{red, leather, suede}| = 1/3 ~ 0.333`
  - `Jaccard(C, A) = 0` (disjoint tag sets)
  - `MMR(B) = 0.75 * 0.90 - (1-0.75) * 0.333 ~ 0.675 - 0.083 = 0.592`
  - `MMR(C) = 0.75 * 0.85 - (1-0.75) * 0 = 0.6375`
  - Pick C
  - Note: Diversity wins.

- Round 3 (`selected = {A,C}`):
  - `Jaccard(B, A) = 1/3 ~ 0.333`
  - `Jaccard(B, C) = 0`
  - `MMR(B) = 0.75 * 0.90 - (1-0.75) * 0.333 ~ 0.675 - 0.083 = 0.592`
  - Pick B.
  - Note: B was the only remaining candidate.

#### Build the response with reasons

- For each returned item, include a compact reason array such as:
  `["popularity", "co_visitation", "embedding", "personalization", "diversity"]`.

### What happens for anonymous users?

- Steps 3–4 (anchors, co‑visitation, embeddings vs anchors) are skipped or
  produce zeros. The system still works using popularity and, optionally, MMR.
- Personalization is skipped (no user profile).

### What happens when embeddings are missing?

- `embed_raw` is absent -> `embed_norm = 0` -> no contribution to the blend.
- Co‑visitation can still add context if the user has recent anchors.
- Otherwise, popularity carries the result (still stable and explainable).

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

| Variable              | Type / Range      | What it does                          | Notes                        |
|-----------------------|-------------------|---------------------------------------|------------------------------|
| `PROFILE_WINDOW_DAYS` | float > 0 or `-1` | Lookback for building user profile.   |                              |
| `PROFILE_TOP_N`       | int > 0           | Keep only the strongest N tags.       | Higher N = broader, noisier  |
| `PROFILE_BOOST`       | float ≥ 0         | Strength of the multiplicative boost. | `0` disables personalization |

### Blended scoring weight vars

We rescore each candidate using normalized signals:

```plaintext
final = alpha*pop_norm + beta*co_vis_norm + gamma*embed_norm
```

| Variable      | Type / Range | What it does                                | Notes                             |
|---------------|--------------|---------------------------------------------|-----------------------------------|
| `BLEND_ALPHA` | float ≥ 0    | Weight for normalized popularity.           | If all three are zero, set to `1` |
| `BLEND_BETA`  | float ≥ 0    | Weight for normalized co-vis strength.      | Needs user anchors                |
| `BLEND_GAMMA` | float ≥ 0    | Weight for normalized embedding similarity. | Needs embeddings + user anchors   |

**Why normalize?** Raw signals live on different scales (counts,
decayed sums, cosine similarity). Normalizing to `[0, 1]` makes the
weights intuitive and the blend stable. Channels with no signal produce
0 and have no effect.

### Contextual Bandit

| Variable      | Type / Range | What it does                                       | Notes |
|---------------|--------------|----------------------------------------------------|-------|
| `BANDIT_ALGO` | string       | Multi-armed bandit algorithm (`thompson`, `ucb1`). |       |

## Tuning Cheat-Sheet

- Start with `alpha=1.0`, `beta=0.1`, `gamma=0.1`.
- Raise **beta** if you want more "also viewed/bought together."
- Raise **gamma** for cold start and meaning-based tilt.
- `MMR_LAMBDA=0.6` is a reasonable diversity starting point.
- Keep light personalization gentle: `PROFILE_BOOST` around `0.1–0.3`.
- Caps (`BRAND_CAP`, `CATEGORY_CAP`) enforce catalog variety.

## Tenancy

- Multi-tenant by `(org_id, namespace)`.
- IDs are opaque; keep user privacy on the client side.

## Explore

- **Swagger UI**: open `/docs` when the service is running.
- **Makefile**: see targets for dev, tests, and migrations.
- **Demo UI**: Access the interactive demo with user traits editor at the demo UI URL.

## Glossary

**ALS**
The embedding-similarity signal used in the blended score (the "gamma" term).
It measures semantic similarity via cosine similarity of item embeddings to
anchors (recent user items or a specific item). It is not matrix factorization
here. Increasing ALS gives more weight to semantic similarity relative to
popularity and co‑visitation.

**Anchors**  
A user's most recent items (within `COVIS_WINDOW_DAYS`) i.e. items the current
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
look up the edges for the user's anchors.

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
means an event's influence halves every 14 days.

**Item**
something you could recommend (e.g., product "A").

**Light personalization**
A small score boost for items that share tags with what this user has
recently engaged with. The engine builds a short-lived tag profile from the
user's own events (using your event-type weights and half-lives), then, for
each candidate item, sums the overlap of its tags with that profile and
multiplies the item's score by `1 + PROFILE_BOOST * overlap`. Set
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
A small multiplicative boost based on overlap between the user's decayed
tag profile and the candidate's tags: `score *= (1 + PROFILE_BOOST *
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
A lightweight map of "tag" -> weight that summarizes a user's recent interests,
normalized to sum to 1
(e.g., `{"t:android": 0.6, "category:phone": 0.3, "brand:acme": 0.1}`). It's
computed from the user's past events by applying your per-type weights and time
decay (half-life), optionally limited to a recent window and capped to the top-N
tags for performance. No model training or embeddings are required; it's just
decayed counts on item tags.

**Windows**  
Lookback durations for data (e.g., `COVIS_WINDOW_DAYS`, `PROFILE_WINDOW_DAYS`).
Larger windows mean more history; smaller windows favor recency.
