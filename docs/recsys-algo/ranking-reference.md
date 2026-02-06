# Ranking & constraints reference

## Who this is for

- Recommendation engineers reviewing what RecSys implements (and what it expects from data)
- Developers/operators who need deterministic behavior and debuggable failure modes

## What you will get

- The **implemented** signals and their required stores/artifacts
- The main configuration knobs (service env vars) that change ranking behavior
- Determinism guarantees and the common ways determinism can be broken

!!! info "Evaluation scope"
    Capability boundaries: [`explanation/capability-matrix.md`](../explanation/capability-matrix.md). Non-goals:
    [`start-here/known-limitations.md`](../start-here/known-limitations.md).

## Pipeline order (what runs when)

At a high level, `recsys-algo` runs this sequence:

1. **Candidate pool**: fetch candidates (always at least popularity; optionally other sources).
2. **Exclusions**: remove explicitly excluded items and (optionally) recently-purchased items.
3. **Constraints (metadata-dependent)**: apply include-tags / price / freshness constraints when enabled.
4. **Signals**: gather optional per-request signals (co-visitation, embeddings).
5. **Scoring**: compute a deterministic blended score.
6. **Personalization** (optional): multiplicative boost based on user tag-profile overlap.
7. **Rules** (optional): pin / boost / block (may inject items).
8. **Post-ranking diversity** (optional): MMR re-ranking + brand/category caps.
9. **Response**: sort by score, tie-break by `item_id`, attach reasons/explain blocks when requested.

Formal scoring spec: [`recsys-algo/scoring-model.md`](scoring-model.md)

If you need the serving-layer view (what happens in `recsys-service` before/after the algorithm), read:
[`explanation/candidate-vs-ranking.md`](../explanation/candidate-vs-ranking.md).

## Algorithm modes (baseline strategy)

The service-level default is controlled by `RECSYS_ALGO_MODE`:

- `blend`: use configured blend weights (default behavior when evaluating multiple signals).
- `popularity`: popularity-only baseline.
- `cooc`: co-visitation baseline (requires co-occurrence store/history).
- `implicit`: collaborative baseline (requires collaborative store, e.g. ALS).

Per request, you can also set `algorithm` (see the API reference) to override the mode.

## Signals (implemented)

Signals can contribute in two ways:

- **candidate retrieval**: adds/changes which items are in the pool
- **scoring**: changes how the pool is ranked

### Popularity (required baseline)

- **Signal:** `popularity`
- **Used for:** candidate retrieval + scoring baseline
- **Required store:** `PopularityStore.PopularityTopK`
- **Main knobs:**
  - `RECSYS_ALGO_HALF_LIFE_DAYS` (time-decay)
  - `RECSYS_ALGO_POPULARITY_FANOUT` (how many candidates to fetch vs `k`)
  - `RECSYS_ALGO_MAX_K`, `RECSYS_ALGO_MAX_FANOUT` (safety caps)
- **Common failure modes:**
  - Empty/underfilled popularity table → empty or low-quality results
  - Namespace/surface mismatch → “looks empty” even though data exists elsewhere

### Co-visitation (item co-occurrence)

- **Signal:** `cooc`
- **Used for:** (a) candidate retrieval in `cooc` mode, (b) scoring contribution in `blend` mode when enabled
- **Required stores:**
  - `HistoryStore.ListUserRecentItemIDs` (to find recent anchors), or request-provided anchors
  - `CooccurrenceStore.CooccurrenceTopKWithin` (neighbors for each anchor)
- **Main knobs:**
  - `RECSYS_ALGO_COVIS_WINDOW_DAYS` (window for co-occurrence neighbors)
- **Common failure modes:**
  - No recent user history → no anchors → co-vis contributes nothing
  - Missing store/artifacts → `SIGNAL_UNAVAILABLE` warnings
  - Partial failures per-anchor → `SIGNAL_PARTIAL` warnings

### Similarity (max of sub-signals)

Similarity is treated as a **bucket**. The scoring term uses the maximum normalized value across these sub-signals:

- `embedding`
- `collaborative`
- `content`
- `session`

It is controlled by:

- default weight: `RECSYS_ALGO_BLEND_GAMMA`
- request weight: `weights.emb` (API field name)

#### Embedding similarity

- **Signal:** `embedding`
- **Required stores:** `HistoryStore` (anchors) + `EmbeddingStore.SimilarByEmbeddingTopK`
- **Common failure modes:** missing embeddings / missing store → `SIGNAL_UNAVAILABLE`

#### Collaborative similarity (e.g. ALS)

- **Signal:** `collaborative`
- **Required store:** `CollaborativeStore.CollaborativeTopK`
- **Common failure modes:** missing factors/model → `SIGNAL_UNAVAILABLE`

#### Content/tag similarity

- **Signal:** `content`
- **Required stores:** `ProfileStore.BuildUserTagProfile` + `ContentStore.ContentSimilarityTopK`
- **Main knobs:**
  - `RECSYS_ALGO_PROFILE_WINDOW_DAYS`, `RECSYS_ALGO_PROFILE_TOP_N`
- **Common failure modes:**
  - No usable profile (sparse/no events) → content similarity contributes nothing
  - Missing store/artifacts → `SIGNAL_UNAVAILABLE`

#### Session sequence

- **Signal:** `session`
- **Required store:** `SessionStore.SessionSequenceTopK`
- **Main knobs:**
  - `RECSYS_ALGO_SESSION_LOOKBACK_EVENTS`
  - `RECSYS_ALGO_SESSION_LOOKAHEAD_MINUTES`
- **Common failure modes:** no session events / missing store → no contribution or `SIGNAL_UNAVAILABLE`

### Personalization boost (tag overlap)

Personalization is a **post-score multiplier** applied when a user profile exists.

- **Required store:** `ProfileStore.BuildUserTagProfile`
- **Main knobs:**
  - `RECSYS_ALGO_PROFILE_BOOST` (strength; set `0` to disable)
  - `RECSYS_ALGO_PROFILE_MIN_EVENTS` + `RECSYS_ALGO_PROFILE_COLD_START_MULT`
  - `RECSYS_ALGO_PROFILE_STARTER_BLEND_WEIGHT` (blend starter presets with sparse history)
- **Common failure modes:** sparse history → boost attenuated; store unavailable → boost skipped

## Controls (implemented)

### Exclusions

- Explicit exclude IDs are removed from consideration.
- Optional “exclude by events” can filter recently purchased/engaged items:
  - `RECSYS_ALGO_RULE_EXCLUDE_EVENTS`
  - `RECSYS_ALGO_PURCHASED_WINDOW_DAYS`
  - `RECSYS_ALGO_EXCLUDE_EVENT_TYPES`

### Rules (pin / boost / block)

Rules are a serving-layer control plane feature. When enabled, they can:

- **pin** items to the top (and inject items not in the pool)
- **boost** or **block** items by surface/segment

Key knob:

- `RECSYS_ALGO_RULES_ENABLED`

### Diversity (MMR) and caps

Post-ranking re-ranking supports:

- MMR-style diversification: `RECSYS_ALGO_MMR_LAMBDA` (0 disables)
- Brand/category caps:
  - `RECSYS_ALGO_BRAND_CAP`, `RECSYS_ALGO_CATEGORY_CAP`
  - plus tag prefix settings for how brand/category are extracted

## Determinism guarantees

`recsys-algo` is deterministic **given deterministic inputs**:

- Candidates are sorted by score and **tie-broken by `item_id`**.
- Optional explainability does not change ranking.

Determinism can be broken by:

- **Non-deterministic store backends** (e.g., DB queries without stable ordering on ties).
- **Time-dependent windows** (history/co-vis windows depend on “now”; fix the clock for reproducible tests).
- **Custom algorithm plugins** (if enabled) that use randomness or non-stable ordering.

## Debugging signals and fallbacks

When optional signals are missing or partial, `recsys-service` emits warnings like:

- `SIGNAL_UNAVAILABLE`
- `SIGNAL_PARTIAL`

Typical first checks:

- Namespace/surface mismatch (`surface` → namespace mapping)
- Missing artifacts in artifact mode / missing seed data in DB-only mode
- Join-rate and instrumentation integrity (see the integration checklist)

## Read next

- Store ports (what each signal needs): [`recsys-algo/store-ports.md`](store-ports.md)
- Candidate vs ranking (serving-layer mental model): [`explanation/candidate-vs-ranking.md`](../explanation/candidate-vs-ranking.md)
- Integration checklist: [`how-to/integration-checklist.md`](../how-to/integration-checklist.md)
