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
  - Cold-start users now receive a curated starter profile based on their segment traits. The first few interactions still matter: the profile boost is attenuated until a configurable number of events are present, preventing the system from over-fitting to a single click.

- **Diversity & caps** (optional) to avoid showing too many items from
  one brand/category.

- **Manual overrides & catalog tooling** so merchandising teams can boost,
  suppress, or pin items in real time (with audit history) and keep item
  metadata fresh through backfills.

- **Blended scoring** decides how high each candidate should rank, and the
  re‑ranker (MMR + caps) decides which of the high scorers make the final top‑K.

- **Contextual multi-armed bandit** (optional) picks the best ranking
  policy per surface and context, learning online from later rewards
  (e.g., click or purchase).

- **Audit trail** allows listing recent decisions with filters for namespace,
  time range, user hash, or request id. Fetch the full stored trace for a single
  decision.

- **Rule engine** allows you to create business rules that override the normal
  recommendation algorithm. You can block certain items from appearing, pin
  specific items to the top of results, or boost items with extra score. Rules
  can be scoped to specific surfaces (like "homepage" or "product page") and
  user segments, with optional time limits. The system includes a dry-run mode
  to test which rules would apply to a set of items before making changes.

### Bringing Signals Online

A fresh environment only ships with popularity enabled. Run the following flow
to activate richer signals during development or evaluation:

1. **Seed catalog + users + events** via the API or demo shop so the popularity
   candidate pool is non-empty.
2. **Generate embeddings (gamma signal)** — run `make catalog-backfill` once to
   populate deterministic embeddings, then `make catalog-refresh SINCE=24h` on a
   schedule to keep them fresh.
3. **Synthesize collaborative factors (ALS signal)** — `make collab-factors
   SINCE=24h` writes `recsys_item_factors` and `recsys_user_factors` using the
   embedding-backed heuristics so the collaborative retriever returns
   candidates.
4. **Allow rule cache refresh** (defaults to `RULES_CACHE_REFRESH=2s`) after
   creating manual overrides or merchandising rules. Use
   `/v1/admin/manual_overrides` or `/v1/admin/rules` to confirm activation.
5. **Verify coverage** by calling `/v1/recommendations` with
   `"include_reasons": true` and inspecting `trace.extras.candidate_sources`;
   non-zero counts for `collaborative`, `content`, and `popularity` confirm all
   signals are live.

Repeat steps 2–3 after large catalog uploads so candidate sources stay fresh.

### Quick regression check

After changing ranking knobs or data, run the scenario harness to validate policy and cold-start behaviour:

```bash
python analysis/scripts/run_scenarios.py --base-url https://api.pepe.local --org-id "$RECSYS_ORG_ID"
# or locally (with the dockerized API running):
make scenario-suite SCENARIO_BASE_URL=http://localhost:8000 SCENARIO_ORG_ID="$RECSYS_ORG_ID"
```

The script saves evidence under `analysis/evidence/` by default (override with `SCENARIO_EVIDENCE_DIR`, which CI points to `analysis_v2/evidence/ci`) and rewrites `analysis/scenarios.csv`; scenario **S7** now records the starter profile applied to cold-start users, enforces average `mrr@10` ≥ 0.2 and ≥ 4 categories (override with `--s7-min-avg-mrr` / `--s7-min-avg-categories`), while **S8/S9** confirm boost and trade-off telemetry. The workflow at `.github/workflows/scenario-suite.yml` runs the same harness on every push/PR using the `ci` profile and leaves the collected artifacts in GitHub Actions.

Need the recommended diversity settings for a surface? Call `GET /v1/admin/recommendation/presets` to retrieve the current `mmr_lambda` presets (parsed from `MMR_PRESETS` or the built-in defaults) so tooling can present validated dropdowns.

For end-to-end quality checks, run:

```bash
python analysis/scripts/run_quality_eval.py --base-url https://api.pepe.local --org-id "$RECSYS_ORG_ID"
# CI example (see .github/workflows/quality-eval.yml):
python analysis/scripts/run_quality_eval.py --base-url http://localhost:8000 --org-id "$RECSYS_ORG_ID" --namespace default
```

This script rewrites `analysis/quality_metrics.json`, refreshes `analysis/evidence/recommendation_samples_after_seed.json`, and enforces the evaluation rubric (≥10% lift per segment by default, coverage ≥0.60, long-tail share ≥0.20). Use `--min-segment-lift-ndcg` / `--min-segment-lift-mrr` if you need to temporarily override the +10% guardrail; the CI workflow stores artifacts under `analysis_v2/evidence/quality` so reviewers can diff results.

Both the scenario suite and quality eval accept `--env-file` and automatically embed `env_file` / `env_hash` metadata within `analysis/scenarios.csv`, `analysis/evidence/scenario_summary.json`, and `analysis/quality_metrics.json`, ensuring every evidence snapshot is traceable to the exact `api/.env` used.

Need to automate the whole workflow for a customer profile? Use the simulation orchestrator:

```bash
python analysis/scripts/run_simulation.py \
  --customer beta-retail \
  --base-url http://localhost:8000 \
  --org-id "$RECSYS_ORG_ID" \
  --env-profile dev \
  --env-override PROFILE_BOOST=0.55 \
  --fixture-path analysis/fixtures/sample_customer.json
```

The CLI chains env patching → restart → reset → seed → quality → scenarios and drops a provenance bundle under `analysis/reports/<customer>/<timestamp>/simulation_metadata.json` so every experiment is auditable.

Each run also copies the key artifacts (quality metrics, scenario summary, seeding manifests, scenario CSV) into the same report folder and writes a Markdown digest that highlights overall lifts and scenario pass rates—perfect for sharing with customers or reviewers.

Prefer a turnkey example? Run `analysis/scripts/examples/run_marketplace_simulation.sh` (set `BASE_URL`, `NAMESPACE`, etc.) to seed the marketplace template, enforce guardrails, and collect reports in one step.

To blast through multiple customers in one go, supply a manifest (YAML or JSON). Example (`analysis/fixtures/batch_simulations.yaml`):

```yaml
runs:
  - customer: beta-retail
    fixture_path: analysis/fixtures/sample_customer.json
    env_overrides:
      - PROFILE_BOOST=0.55
  - customer: media-plus
    namespace: media
    env_profile: dev
    env_overrides:
      - PROFILE_STARTER_BLEND_WEIGHT=0.45
      - POPULARITY_FANOUT=800
```

Execute with:

```bash
python analysis/scripts/run_simulation.py --batch-file analysis/fixtures/batch_simulations.yaml --batch-name pilot-rollout
```

The command iterates through each run, generates the per-customer report folders described above, and drops an aggregate summary under `analysis/reports/batches/pilot-rollout_<timestamp>.json`.

### Guardrail configuration

Guardrails are centralized in `guardrails.yml`. The file contains a `defaults` block and optional per-customer overrides (namespaces, segment lift thresholds, catalog coverage minimums, S7 expectations). `run_simulation.py`, the CI workflows, and the Make targets load this file automatically so every evidence artifact—including `analysis/quality_metrics.json`, `analysis/scenarios.csv`, and CI runs—enforces the correct targets.

- To tweak a customer, add or update its entry under `customers:` (or `customers.<name>.namespaces.<ns>` if you need per-namespace granularity).
- Run dry runs with `python analysis/scripts/run_simulation.py --customer <name> --guardrails-file guardrails.yml --dry-run` to confirm the resolved thresholds.
- CI reads the same file, so guardrail changes automatically propagate to `.github/workflows/quality-eval.yml` and `scenario-suite.yml`. Keep PRs that adjust guardrails small and include rationale in the description.

If you need to test with a bespoke config, pass `--guardrails-file path/to/custom.yml` to the simulation runner (or set `GUARDRAILS_FILE` in workflows). Set the flag to an empty string to disable guardrails during local debugging.

For an end-to-end walkthrough (profiles, fixtures, simulations, evidence), see `docs/bespoke_simulations.md`.

### Onboarding & Coverage Checklist

1. **Start with a clean namespace** – execute `analysis/scripts/reset_namespace.py --base-url <API> --namespace <NS> --org-id <ORG> --force` (or `make reset-namespace SCENARIO_BASE_URL=<API> SCENARIO_ORG_ID=<ORG>`). The helper calls `/v1/events:delete`, `/v1/users:delete`, and `/v1/items:delete` so bespoke fixtures never mingle with previous data and writes `analysis/evidence/reset_<timestamp>.json` for auditing.
2. **Seed data consistently** – run `analysis/scripts/seed_dataset.py` against the target API so scenarios/quality/determinism reuse the catalog snapshot stored in `analysis_v2/evidence/seed_manifest.json`. Pass `--fixture-path analysis/fixtures/sample_customer.json` (or your own JSON) to ingest bespoke items/users/events when testing customer-specific catalogs; the script now emits `seed_segments.json` with per-segment counts and sample traits for quick validation.
   - Need inspiration? The templates under `analysis/fixtures/templates/` (marketplace, media, retail) plus the accompanying `analysis/fixtures/README.md` show the required fields and props you can copy into your own fixture.
3. **Validate cold-start** – ensure scenario S7 passes and inspect `analysis_v2/evidence/scenario_s7_cold_start.json` for ≥4 categories plus personalization reasons on the first item.
4. **Track segment lifts** – compare fresh `analysis/quality_metrics.json` with the evaluation baseline (`analysis_v2/quality_metrics.json`) to confirm each cohort stays ≥+10% on NDCG/MRR.
5. **Watch coverage telemetry** – export `policy_item_served_total`, `policy_coverage_bucket_total`, and `policy_catalog_items_total`; alert if catalog coverage <0.60 or long-tail share <0.20 per the queries documented in `docs/rules-runbook.md`.
6. **Share remediation summary** – include `analysis_v2/remediation_summary.md` (plus `analysis_v2/report.md`) in rollout notes so stakeholders see which epics are complete and which CI guardrails enforce the targets.

### Cold-start guardrails

- Scenario suite S7 enforces `avg_mrr@10 ≥ 0.2` and `avg_categories ≥ 4`; override via `--s7-min-avg-mrr` / `--s7-min-avg-categories` when running `analysis/scripts/run_scenarios.py`.
- Quality eval fails the job if any segment’s NDCG/MRR lift falls below the configured thresholds (defaults `--min-segment-lift-ndcg=0.1`, `--min-segment-lift-mrr=0.1`). CI relies on these exits, so only override locally when experimenting.
- Evidence for both guardrails is stored under `analysis_v3/evidence/` (`scenario_s7_cold_start.json`, `quality_metrics.json`) for easy diffing in reviews.

## Configuration profiles and feature flags

- The API recognises three configuration profiles: `development`, `test`, and
  `production`. Set the `ENV` environment variable (defaults to `development`)
  to pick the profile. Each profile applies sensible defaults for debug mode,
  feature toggles (rules, decision trace, LLM explain), and observability.
- Generate ready-to-use `.env` files with `make env PROFILE=dev|test|prod|ci`. The
  command copies the matching template from `api/env/` into `api/.env`
  (and `api/.env.test` for the `test` profile). For scripted tweaks, run
  `python analysis/scripts/configure_env.py --namespace default --set PROFILE_BOOST=0.45`
  (repeat `--set` for multiple overrides). Every run rewrites `api/.env` and records
  the change—including the SHA and before/after values—in
  `analysis/env_history/<timestamp>.json`, keeping experiments auditable. Follow it
  with `python analysis/scripts/restart_api.py --base-url http://localhost:8000`
  (or `make restart-api`) to force `docker compose up -d --force-recreate api` and wait
  for `/health` before seeding or running scenarios.
- Run `python analysis/scripts/check_env_profiles.py --strict` to ensure
  every profile under `api/env/` contains the same algorithm env vars as
  `api/.env`. Update `config/profiles.yml` to document which namespaces use each
  profile—`analysis/scripts/configure_env.py` will auto-select the matching profile
  when you pass `--namespace`, so you can hot-swap namespaces without restarting
  the API (simulation runner integration coming next).
- The `ci` profile mirrors production knobs while disabling external
  integrations (e.g., LLMs) so GitHub Actions can spin up the stack, run
  the scenario suite deterministically, and publish artifacts without secrets.
- The config loader exposes consolidated feature flags via
  `config.Config.Features`, keeping the effective state of optional subsystems
  in one place.
- `make test` automatically runs the suite with the `test` profile so the API
  starts in a deterministic, offline-friendly mode.
- `make catalog-backfill` and `make catalog-refresh SINCE=24h` help populate and
  keep catalog metadata/embeddings fresh (run after adjusting the `.env`).
- `make collab-factors SINCE=24h` materialises collaborative item/user factors
  so `/v1/bandit/recommendations` and `/v1/recommendations` can draw from the
  ALS candidate source during development.

## Recommendation Algorithms Used

These algorithms are used together in the recommendation pipeline.

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

## How Recommendation Works

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

#### Candidate sources

- **Popularity retriever** pulls the global trending pool using the configured half-life and fanout; it's the baseline when we lack other signals.
- **Collaborative retriever** queries ALS factors via `CollaborativeTopK` to surface user-specific recommendations that the popularity pool missed.
- **Content retriever** reuses the user tag profile and `ContentSimilarityTopK` to add catalog-rich items that align with declared interests.
- **Session retriever** (`SessionSequenceTopK`) looks at the most recent interactions and adds frequently-followed items for short-term intent.
- **Source telemetry**: every request records per-source counts/durations in the decision trace (`sourceMetrics`) so dashboards and scenario evidence show which retrievers contributed and how long they took.

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
[Request-time] Re-rank    : MMR + caps -> Top-K
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
  - Typical: `0.1..0.3` to keep coverage high while preserving relevance. Push toward `1.0` if you prefer pure relevance.
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

### Recommendation FAQs

- **What happens for anonymous users?**
  - Steps with anchors, co‑visitation, embeddings vs anchors are skipped or
    produce zeros. The system still works using popularity and, optionally, MMR.
  - Personalization is skipped (no user profile).

- **What happens when embeddings are missing?**
  - `embed_raw` is absent -> `embed_norm = 0` -> no contribution to the blend.
  - Co‑visitation can still add context if the user has recent anchors.
  - Otherwise, popularity carries the result (still stable and explainable).

## Explainability & Governance Features

### "Why this recommendation?" explanations

- Every recommendation API response can include an item-level `reasons` array
  when `include_reasons=true`.
- Structured explanations are available via the `explain_level` parameter
  (`tags`, `numeric`, `full`). When requested, the API returns an `explain`
  block that exposes blend contributions, personalization overlap, MMR details,
  diversity caps, and the recent anchors used during scoring.
- These signals are the same ones the ranker used, which makes it easy to
  surface badges in the UI or to debug ranking behaviour during integrations.

### Segment profiles & rules

- You can use the API to upload segment profiles to define reusable weight
  bundles (blend weights, MMR, caps, personalization settings, time windows).
- Attach those profiles to audience segments with the API. Segments can match
  users on traits, request context, or namespace-level rules. The active profile
  is applied automatically during recommendation.
- Dry running lets you test which profile a hypothetical user would receive.
  This makes it straightforward to tailor ranking knobs for cohorts such as
  "new", "returning", or "VIP" players without code changes.

### Manual overrides for boosts and suppressions

- Merchants can register ad-hoc overrides (boost or suppress) through the
  `/v1/admin/manual_overrides` API. Each override creates a dedicated rule
  behind the scenes, honours expiry, and records who created or cancelled it.
- Overrides are listed, queried, and cancelled through the same endpoint, and
  every action is stored so you can audit merchandising changes later.
- Overrides share the same namespace/surface scoping as regular rules, so
  guardrails stay isolated per placement and tenant.

### Decision audit trail

- Each recommendation can be persisted as a `rec_decisions` record containing
  request metadata, effective configuration, pre/post ranking snapshots,
  reasons, and optional bandit context.
- Enable the writer via the `AUDIT_DECISIONS_*` environment variables; the async
  worker batches inserts to avoid impacting request latency.
- These traces power compliance reviews, post-mortems, and "show me exactly what
  the user saw" workflows, using the same hashed identifiers described in the
  audit configuration.

## Contextual Multi‑Armed Bandit

A lightweight, online‑learning AI component that chooses which ranking
policy (a bundle of scorer knobs) to use per surface and request
context. It learns directly from the live rewards you send later (e.g.,
click/purchase), balancing exploration vs exploitation without
offline training.

- **Policy** = one complete set of ranker knobs:
  `blend_alpha`, `blend_beta`, `blend_gamma`, `mmr_lambda`, `brand_cap`,
  `category_cap`, plus metadata like `policy_id`, `name`, `active`.

- **Contextual** = learning is tracked per surface (placement like
  `"home_top"`, `"pdp_carousel"`) and a compact context bucket derived
  from a small map you send (e.g., `{"device":"ios","locale":"fi"}` ->
  `ctx:device=ios|locale=fi`).

The bandit does not tune knobs one‑by‑one. It selects among the predefined
policies and shifts traffic toward the winners.

The recommendation ranking pipeline stays the same. The bandit only chooses the
policy for this request. You can call the one‑shot endpoint to "decide + rank" in
one go, then send a reward later when you know the outcome.

> ℹ️  When you run the demo shop, it first calls `/v1/bandit/policies` with the
> configured policy IDs. If none exist (or the lookup fails), the shop disables
> exploration automatically, logs a single warning, and falls back to the
> standard `/v1/recommendations` endpoint so local runs aren’t flooded with 500s.

Profiles that drive the core blend/diversity knobs are stored in the shop
database (`RecommendationProfile`). Use `/api/admin/recommendation-settings`
or the Admin UI to create/edit profiles, then pass `profileId=<id>` on any shop
request. Inline overrides in the request body still win if you need a quick
one-off experiment.

### Usage: end‑to‑end with examples

#### Define policies (the arms)

```json
POST /v1/bandit/policies:upsert
{
  "namespace": "default",
  "policies": [
    {
      "policy_id": "p_baseline",
      "name": "Baseline blend",
      "active": true,
      "blend_alpha": 1.0,
      "blend_beta": 0.1,
      "blend_gamma": 0.1,
      "mmr_lambda": 0.8,
      "brand_cap": 0,
      "category_cap": 0
    },
    {
      "policy_id": "p_diverse",
      "name": "Diverse caps",
      "active": true,
      "blend_alpha": 1.0,
      "blend_beta": 0.2,
      "blend_gamma": 0.2,
      "mmr_lambda": 0.12,
      "brand_cap": 1,
      "category_cap": 2
    }
  ]
}
```

Inspect later:

```plaintext
GET /v1/bandit/policies?namespace=default
```

#### Decide a policy for this request

Or use the one‑shot below.

```json
POST /v1/bandit/decide
{
  "namespace": "default",
  "surface": "home_top",
  "context": { "device": "ios", "locale": "fi" },
  "candidate_policy_ids": ["p_baseline", "p_diverse"],
  "algorithm": "thompson",
  "request_id": "req-12345"
}
```

Example response:

```json
{
  "policy_id": "p_diverse",
  "algorithm": "thompson",
  "surface": "home_top",
  "bucket_key": "ctx:device=ios|locale=fi",
  "explore": true,
  "explain": { "emp_best": "p_baseline" }
}
```

##### One‑shot: decide + recommend

Returns items + bandit metadata.

```json
POST /v1/bandit/recommendations
{
  "user_id": "u_123",
  "namespace": "default",
  "k": 20,
  "surface": "home_top",
  "context": { "device": "ios", "locale": "fi" },
  "candidate_policy_ids": ["p_baseline", "p_diverse"],
  "algorithm": "thompson",
  "include_reasons": true
}
```

Truncated response:

```json
{
  "items": [{ "item_id": "i_101", "score": 0.87 }],
  "chosen_policy_id": "p_diverse",
  "algorithm": "thompson",
  "bandit_bucket": "ctx:device=ios|locale=fi",
  "explore": true,
  "bandit_explain": { "emp_best": "p_baseline" }
}
```

##### Or do it manually

Call `/v1/bandit/decide`, then call your normal `/v1/recommendations` with that
policy's knobs as overrides.

#### Reward

Later, when you know the outcome.

```json
POST /v1/bandit/reward
{
  "namespace": "default",
  "surface": "home_top",
  "bucket_key": "ctx:device=ios|locale=fi",
  "policy_id": "p_diverse",
  "reward": true,
  "algorithm": "thompson",
  "request_id": "req-12345"
}
```

**What counts as a reward?** Whatever you decide (click, add‑to‑cart,
purchase, dwell‑time threshold, etc.). The bandit updates online stats
per `(surface, bucket, policy, algorithm)` and adapts future choices.

### Algorithms Used In the Bandit

We support two classic, minimal‑config bandit algorithms. Both are
"greedy after computing a score" and require only success/failure counts.

#### Thompson Sampling (Beta–Bernoulli)

- Maintain a Beta posterior per arm: `Beta(α, β)` where  
  `α = prior_success + successes`, `β = prior_failure + failures`.
- Decision time: sample one plausible CTR for each arm and pick the
  best.
  - Implementation trick: sample `X ~ Gamma(α,1)`, `Y ~ Gamma(β,1)` and
    set `p = X / (X + Y)`; then choose the arm with max `p`.

Beta posterior means is current belief about an arm's true success
rate (e.g., CTR) after seeing data. It is a `Beta(α, β)` distribution
whose shape narrows as you collect more evidence (more α+β), meaning
you are more confident.

Thompson sampling uses Gamma distribution sampling `gamma(shape, scale)`. It
returns a positive random number. In Thompson we use two draws with
`shape = α and β` (scale=1 for both): `X ~ Gamma(α,1)`, `Y ~ Gamma(β,1)`,
then `Beta(α, β)`, which is `p = X/(X+Y)`.

- Bigger **α (successes)** shifts X larger on average -> **higher** p.
- Bigger **β (failures)** shifts Y larger -> **lower** p.
- Bigger **α+β** (keeping the ratio α/(α+β) fixed) makes p **less noisy**
  (narrower distribution around the mean α/(α+β)).

**Why it works (intuition)**  
Treat each arm's success rate as uncertain. Arms with little data have a
wide posterior so they sometimes sample high (exploration). As evidence
grows the posterior narrows and the best arm wins more often
(exploitation).

- **Pros**
  - Naturally balances explore/exploit via uncertainty.
  - Handles cold start via priors; easy to add decay/sliding windows.
  - Strong empirical performance with binary rewards.

- **Cons**
  - Randomized decisions (harder to replay deterministically).
  - Requires an RNG; audits must record the sampled values to reproduce
    exact choices.

#### UCB1 (Upper Confidence Bound)

For each arm at time N (total trials), compute and pick the arm with the largest
`score_i`:

```plaintext
score_i = mean_i + sqrt(2 * ln(N) / n_i)
```

- **mean_i**: Equal to `successes_i / n_i` (0 if `n_i=0`), `n_i` = trials for
  arm i.
- **N**: total trials across all arms so far. Larger N increases the
  bonus slowly (via ln(N)), ensuring occasional exploration never dies.
- **n_i**: trials for arm i. Larger n_i **reduces** the bonus
  `sqrt(2 ln N / n_i)`, so heavily-tested arms rely mostly on their mean.
- **mean_i**: empirical success rate `successes_i / n_i` (0 if `n_i=0`).
  Higher mean directly raises the score.
- **Practical effect**: if two arms have the same mean, the one with
  fewer trials will be chosen (bigger bonus). As n_i grows, its bonus
  shrinks and selection depends more on mean.

**Why it works (intuition)**
Be optimistic about what you don't know: under‑tried arms get a
bigger bonus, so they're explored. As `n_i` grows the bonus shrinks and
you exploit the better‐measured mean.

- **Pros**
  - Simple, deterministic; strong classical regret bounds.
  - Clear audit story: "mean + uncertainty bonus."

- **Cons**
  - Can over‑explore early; no priors (cold start is uniform).
  - Sensitive to non‑stationarity unless you add decay or windowing. This means
    it considers time‑decayed counts or windowed statistics for arms if your
    environment shifts.

### Bandit FAQs

- **Is this "AI"?** Yes, this is online machine learning. It learns from your
  live rewards. No offline training is needed.

- **Does it tune knobs one‑by‑one?** No. It picks among full policies you
  define. You can add/retire policies anytime.

- **Where does contextual come in?** Learning is segmented by surface and a
  deterministic context bucket built from your `context` map. The same policy
  can be great on mobile but not on web, and the bandit learns that difference.

- **Cold start?** Both algorithms ensure every arm gets tried. Thompson can also
  use informative priors to speed this up. Informative prior are starting
  pseudo-counts (`α0, β0`) that encode prior knowledge or a sensible default.
  Example: `α0=1, β0=1` is "uninformative" (uniform). `α0=10, β0=10` centers
  belief near 50% but with more initial confidence. Priors mainly affect early
  decisions.

## Audit Trail

The audit trail captures each recommendation decision so you can see
what was shown, why items were ordered that way, and which inputs
influenced the outcome. Each trace includes request metadata,
effective configuration, candidate and score snapshots, any bandit
choices, and per‑item reasons. Writing is asynchronous and can be
sampled to limit overhead. Browse summaries with filters, then fetch a
full trace for a single request to support debugging, compliance, and
"show me exactly what the user saw" workflows.

## Rule Engine

The rule engine adds clear business controls on top of the ranking
algorithm. You can block items from appearing, pin specific items to
fixed positions, or boost items by increasing their score. Rules can
be scoped by namespace and surface, limited to user segments, and
scheduled with start/end times. During ranking, blocks remove
candidates, boosts adjust scores, and pins can reserve positions. The
ranker then finalizes the list with diversity and caps. A dry‑run mode
shows which rules would fire for a given set of items so you can
verify intent safely. Common uses include hiding out‑of‑stock or
restricted items, promoting campaigns, curating hero slots, and
enforcing merchandising policies.

You can list existing rules and filter by namespace, surface, segment,
status, action, or active window to find exactly what applies to a
placement. You can update rules at any time. The dry‑run endpoint returns which
rules would match for a given set of item IDs, including matched rule
metadata and per‑item effects. It works without changing state.

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
| `CORS_ALLOWED_ORIGINS`   | string       | Comma-separated allowed CORS origins.         |       |
| `CORS_ALLOW_CREDENTIALS` | bool         | Allow credentials in CORS requests.           |       |

### Proxy vars

| Variable          | Type / Range | What it does                                | Notes                                     |
|-------------------|--------------|---------------------------------------------|-------------------------------------------|
| `WEB_DOMAIN`      | string       | External domain for the web UI.             | Used for self-signed mkcert certificates. |
| `API_DOMAIN`      | string       | External domain for the API.                | Same mkcert flow as the web domain.       |
| `SWAGGER_DOMAIN`  | string       | External domain for the Swagger UI service. | Optional; set to expose docs via proxy.   |
| `WEB_BACKEND`     | host:port    | Upstream address for the web UI container.  | Defaults to `recsys-web:3000`.            |
| `API_BACKEND`     | host:port    | Upstream address for the API container.     | Defaults to `recsys-api:8000`.            |
| `SWAGGER_BACKEND` | host:port    | Upstream address for Swagger container.     | Defaults to `recsys-swagger:8080`.        |

### Windows, decay, and candidate fan-out vars

| Variable                   | Type / Range | What it does                                 | Effect of higher / lower                  |
|----------------------------|--------------|----------------------------------------------|-------------------------------------------|
| `POPULARITY_HALFLIFE_DAYS` | float > 0    | How fast old events fade.                    | Smaller = favors recency; larger = memory |
| `COVIS_WINDOW_DAYS`        | float > 0    | Lookback for co-vis and user anchors.        | Larger = more seasonal signal             |
| `POPULARITY_FANOUT`        | int > 0      | How many popularity candidates to pre-fetch. | Larger = more choice, more DB work        |

### Diversity & business rule vars

| Variable                | Type / Range   | What it does                                    | Notes                           |
|-------------------------|----------------|-------------------------------------------------|---------------------------------|
| `MMR_LAMBDA`            | float in [0,1] | MMR trade-off: 1.0 = relevance, 0.0 = diversity | Set `0` to disable              |
| `MMR_PRESETS`           | string (csv)   | Surface presets such as `home=0.25`.            | Served via `/v1/admin/recommendation/presets`. |
| `BRAND_CAP`             | int ≥ 0        | Max items per brand in the final top-K.         | `0` disables                    |
| `CATEGORY_CAP`          | int ≥ 0        | Max items per category in the final top-K.      | `0` disables                    |
| `RULE_EXCLUDE_EVENTS`   | bool           | Exclude items the user purchased recently.      | Requires `user_id`              |
| `PURCHASED_WINDOW_DAYS` | float > 0      | Lookback for the exclude-purchased rule.        | Required if the rule is enabled |
| `EXCLUDE_EVENT_TYPES`   | string (csv)   | Event type IDs to exclude when the rule is on.  | Comma-separated int16 values    |

### Rule telemetry

Every recommendation response emits structured policy summary data:

- `policy_rule_actions` log lines capture boost/pin/block counts and the rule IDs that fired.
- `policy_rule_exposure` logs report how many boosted or pinned items reached the final list.
- Prometheus counters prefixed with `policy_` (e.g., `policy_rule_actions_total`, `policy_rule_exposure_total`) mirror the same data for dashboards and alerts.
- `policy_rule_zero_effect` logs/counters warn when boosts or pins fire but the item never surfaces (see the runbook below).
- Coverage guardrail counters (`policy_item_served_total`, `policy_coverage_bucket_total`, `policy_catalog_items_total`) power catalog coverage and long-tail alerts—wire them into Prometheus/Grafana to watch the ≥60 % / ≥20 % targets.

Expose the `/metrics` endpoint or ship the JSON logs into your observability stack to watch override health in production.

> Need a deeper playbook? See [docs/rules-runbook.md](docs/rules-runbook.md) for
detailed troubleshooting steps and operational checklists.

### Per-request algorithm overrides

The `/v1/recommendations` request body accepts an optional `overrides` object so tooling can experiment without rewriting `api/.env`. Supported fields map directly to the algorithm config:

| Field | Type | Description |
|-------|------|-------------|
| `popularity_halflife_days` | int | Override the time decay applied to popularity. |
| `covis_window_days` | int | Override the co-visitation lookback window. |
| `popularity_fanout` | int | Increase/decrease the number of popularity candidates seeded into the ranking loop. |
| `mmr_lambda` | float in [0,1] | Adjust the diversification trade-off for the request. |
| `brand_cap` / `category_cap` | int | Tighten or relax per-brand/category caps. |
| `rule_exclude_events` | bool | Toggle the exclude-purchased rule. |
| `purchased_window_days` | int | Set the lookback window for the exclude-purchased rule. |
| `profile_window_days`, `profile_boost`, `profile_top_n` | ints/floats | Tune personalization strength mid-request. |
| `profile_starter_blend_weight` | float in [0,1] | Nudge how much the curated starter profile influences cold-start results. |
| `blend_alpha`, `blend_beta`, `blend_gamma` | float ≥ 0 | Adjust the weighted blend for pop / co-vis / emb retrievers. |
| `bandit_algo` | string (`thompson`, `ucb1`) | Switch the bandit strategy for the request (use sparingly). |

Example:

```json
{
  "user_id": "customer_user_001",
  "namespace": "customer_a",
  "k": 20,
  "overrides": {
    "popularity_fanout": 500,
    "mmr_lambda": 0.3,
    "profile_boost": 0.4,
    "profile_starter_blend_weight": 0.45
  }
}
```

The service applies these overrides on top of the active env profile; if a field is omitted, the baseline env value is used. Stick to the fields listed above—everything else must be changed via env profiles to keep runs reproducible.

### Tag prefix vars

| Variable                | Type / Range | What it does                            | Notes                                      |
|-------------------------|--------------|-----------------------------------------|--------------------------------------------|
| `BRAND_TAG_PREFIXES`    | string (csv) | Tag prefixes that denote brand tags.    | Example: `brand`. Lowercased; ':' ignored. |
| `CATEGORY_TAG_PREFIXES` | string (csv) | Tag prefixes that denote category tags. | Example: `category,cat`.                   |

### Rules engine vars

| Variable              | Type / Range       | What it does                                      | Notes                             |
|-----------------------|--------------------|---------------------------------------------------|-----------------------------------|
| `RULES_ENABLE`        | bool               | Global kill‑switch for the rules engine.          | `false` disables rule evaluation. |
| `RULES_CACHE_REFRESH` | Go duration string | Poll interval for reloading rules (e.g., `2s`).   |                                   |
| `RULES_MAX_PIN_SLOTS` | int > 0            | Maximum number of pin slots allowed per response. |                                   |
| `RULES_AUDIT_SAMPLE`  | float in [0,1]     | Sample rate for emitting rule evaluation audits.  |                                   |

### Light personalization vars

| Variable              | Type / Range      | What it does                          | Notes                        |
|-----------------------|-------------------|---------------------------------------|------------------------------|
| `PROFILE_WINDOW_DAYS` | float > 0 or `-1` | Lookback for building user profile.   |                              |
| `PROFILE_TOP_N`       | int > 0           | Keep only the strongest N tags.       | Higher N = broader, noisier  |
| `PROFILE_BOOST`       | float ≥ 0         | Strength of the multiplicative boost. | `0` disables personalization |
| `PROFILE_MIN_EVENTS_FOR_BOOST` | int ≥ 0 | Minimum recent events required before the full boost applies. | Keeps cold-start users from overpowering the ranking. |
| `PROFILE_COLD_START_MULTIPLIER` | float in [0,1] | Attenuation factor when event count < min. | `0.5` halves the boost while still nudging results. |

### Starter profile & new-user overrides

| Variable | Type / Range | What it does | Notes |
|----------|--------------|--------------|-------|
| `PROFILE_STARTER_BLEND_WEIGHT` | float in [0,1] | Baseline weight assigned to the curated starter profile before any decay. | Higher values lean harder on curated tags even when we observe history. |
| `PROFILE_STARTER_DECAY_EVENTS` | int > 0 | Number of interactions required before the starter profile fully decays back to the base blend weight. | Smaller numbers keep the curated mix dominant for longer. |
| `PROFILE_STARTER_PRESETS` | JSON string | Segment → category weight map that seeds the starter profile. | Example: `'{"new_users":{"electronics":0.25,"books":0.2}}'`. Keys are lower-cased automatically. |
| `NEW_USER_BLEND_ALPHA` / `NEW_USER_BLEND_BETA` / `NEW_USER_BLEND_GAMMA` | float ≥ 0 (optional) | Override the default blend weights when the user is tagged as `new_users` or has sparse history. | Leave unset to inherit global blend weights. |
| `NEW_USER_MMR_LAMBDA` | float in [0,1] (optional) | Override the diversity vs. relevance trade-off for new users. | Pair with blend overrides to keep lists diverse while boosting relevance. |
| `NEW_USER_POP_FANOUT` | int > 0 (optional) | Larger popularity fan-out for cold-start requests so we have more inventory to diversify from. | Defaults to at least 1000 when unset; raise further to hit coverage targets. |

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

### Coverage guardrail vars

| Variable                               | Type / Range     | What it does                                                | Notes                                               |
|----------------------------------------|------------------|-------------------------------------------------------------|-----------------------------------------------------|
| `COVERAGE_CACHE_TTL`                   | Go duration      | Refresh interval for the metadata cache used by guardrails. | Default `10m`; raise if the catalog rarely changes. |
| `COVERAGE_LONG_TAIL_HINT_THRESHOLD`    | float in [0,1]   | `props.popularity_hint` cutoff that marks an item long-tail.| Default `0.01`; tune so alerts match business goals.|

### Contextual Bandit

| Variable      | Type / Range | What it does                                       | Notes |
|---------------|--------------|----------------------------------------------------|-------|
| `BANDIT_ALGO` | string       | Multi-armed bandit algorithm (`thompson`, `ucb1`). |       |

### Decision audit vars

| Variable                           | Type / Range       | What it does                                                            | Notes                                              |
|------------------------------------|--------------------|-------------------------------------------------------------------------|----------------------------------------------------|
| `AUDIT_DECISIONS_ENABLED`          | bool               | Turns the decision-trace pipeline on/off.                               | When `false`, recommendations skip queuing traces. |
| `AUDIT_DECISIONS_SAMPLE_DEFAULT`   | float in [0,1]     | Default sampling rate for namespaces when recording decisions.          | `1.0` = capture all requests.                      |
| `AUDIT_DECISIONS_SAMPLE_OVERRIDES` | string (`ns=rate`) | Comma-separated per-namespace sampling overrides.                       | Example: `casino=1.0,vip=0.5`.                     |
| `AUDIT_DECISIONS_QUEUE`            | int > 0            | Size of the in-memory queue feeding the async writer.                   | Increase for bursty traffic; consumes RAM.         |
| `AUDIT_DECISIONS_BATCH`            | int > 0            | Maximum number of traces persisted per database batch insert.           | Larger batches reduce round-trips.                 |
| `AUDIT_DECISIONS_FLUSH_INTERVAL`   | Go duration string | Max wait before flushing even if the batch is not full (e.g., `250ms`). | Tune for latency vs. throughput.                   |
| `AUDIT_DECISIONS_SALT`             | string             | Secret salt mixed into the user hash stored in audits.                  | Rotate to invalidate old hashes; keep private.     |

## Environment Flags for ExplainLLM

To enable the LLM-powered RCA endpoint following environment variables:

| Variable              | Type / Example   | Description                                       | Notes                             |
|-----------------------|------------------|---------------------------------------------------|-----------------------------------|
| `LLM_EXPLAIN_ENABLED` | `true` / `false` | Toggle the ExplainLLM feature.                    |                                   |
| `LLM_PROVIDER`        | `openai`         | Provider identifier for LLM API.                  | Use `openai` for built-in client. |
| `LLM_MODEL_PRIMARY`   | `o4-mini`        | Default model for ExplainLLM.                     |                                   |
| `LLM_MODEL_ESCALATE`  | `o3`             | Fallback model for large fact packs.              |                                   |
| `LLM_TIMEOUT`         | `6s`             | Request timeout (Go duration string).             | Example: `6s`                     |
| `LLM_MAX_TOKENS`      | integer          | Maximum tokens to request from the model.         |                                   |
| `LLM_API_KEY`         | string           | Provider API key (required if feature enabled).   | Keep secret.                      |
| `LLM_BASE_URL`        | string (URL)     | Optional override for the Responses API endpoint. |                                   |

## Tuning Cheat-Sheet

- Start with `alpha=0.1`, `beta=0.35`, `gamma=0.55` for balanced coverage (popularity, co-vis, ALS respectively).
- Raise **alpha** if top-line relevance drops; raise **beta** for "also viewed" behaviour; raise **gamma** for cold-start breadth.
- `MMR_LAMBDA=0.12` is the catalog-coverage tuned starting point; raise it toward `0.3` if you need tighter relevance.
- Keep light personalization gentle: `PROFILE_BOOST` around `0.1–0.3`.
- Caps (`BRAND_CAP`, `CATEGORY_CAP`) enforce catalog variety.

## Tenancy

- Multi-tenant by `(org_id, namespace)`.
- IDs are opaque; keep user privacy on the client side.

## Explore

- **Swagger UI**: served by the Swagger service (default `http://localhost:8081` or `https://docs.<your-domain>`).
- **Makefile**: see targets for dev, tests, and migrations.
- **Web UI**: Access the interactive web UI with user traits editor at the web UI URL.

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
- `excluded_events` (item was filtered earlier due to rule)
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
