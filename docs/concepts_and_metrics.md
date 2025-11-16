# Concepts & Metrics Primer

Read this when you need plain-language definitions of the jargon that appears throughout the RecSys documentation. It consolidates the glossary, explains how we measure success, and clarifies how guardrails fit in.

---

## How to use this doc

Skim the cards below when a term shows up in another doc. Each card follows the same structure:

- **What it is** – 1–2 sentences in plain language.
- **Where you’ll see it** – Endpoints, dashboards, or config files that reference the concept.
- **Why you should care** – Business or engineering stakes.
- **Advanced details** – Optional deeper notes (math, knobs, acronyms). Feel free to skip on your first pass.

---

## Index

- [Core concepts](#1-core-concepts)
- [Metrics primer](#2-metrics-primer)
- [Guardrails in context](#3-guardrails-in-context)
- [Glossary](#4-glossary)

---

## 1. Core concepts

### Namespace

- **What it is:** A logical tenant or surface bucket. Every request includes `namespace` so RecSys knows which catalog, rules, and guardrails to apply.
- **Where you’ll see it:** All ingest calls (`*:upsert`), `/v1/recommendations`, guardrail bundles, admin APIs, and traces.
- **Why you should care:** Namespaces isolate data and safety controls. Mixing customers or surfaces in one namespace makes troubleshooting and guardrails harder.
- **Advanced details (optional):** Pair namespaces with env profiles in `config/profiles.yml`. Simulation reports use namespaces as folder names under `analysis/reports/`.

### Org

- **What it is:** A company/customer identifier passed via `X-Org-ID`.
- **Where you’ll see it:** Every HTTP request header, database schemas (`org_id` column), audit logs.
- **Why you should care:** Multi-tenant isolation happens at org + namespace. Ensure your clients always send the correct UUID.
- **Advanced details (optional):** Org IDs map to tenant-specific auth/keys if `API_AUTH_ENABLED=true`.

### Candidate

- **What it is:** An item ID that is eligible to be ranked.
- **Where you’ll see it:** Trace extras (`candidate_sources`), `analysis/results/*_rerank.json`, retrieval dashboards.
- **Why you should care:** Candidate quality affects recall and coverage before ranking even runs.
- **Advanced details (optional):** Candidate pools come from popularity, co-visitation, embeddings, rules, and session heuristics.

### Signals

- **What it is:** Independent evidence that scores a candidate (popularity, co-visitation, embeddings, personalization boosts, etc.).
- **Where you’ll see it:** Blend overrides (`overrides.blend`), traces, tuning harness summaries.
- **Why you should care:** Understanding signals helps explain “why” an item appeared and where to tune.
- **Advanced details (optional):** Signals are normalized to `[0,1]` per request before blending; see `docs/env_reference.md` for knobs.

### Blended scoring

- **What it is:** A weighted sum of normalized signals (e.g., `alpha*pop + beta*co_vis + gamma*embed`).
- **Where you’ll see it:** `/v1/recommendations` overrides, env profiles, tuning harness output.
- **Why you should care:** It’s the core ranking formula; changing weights shifts the balance between relevance and exploration.
- **Advanced details (optional):** Overrides let you change weights per call; tuning harness stores best-performing blends under `analysis/results/tuning_runs/`.

### Retrieval vs ranking

- **What it is:** Retrieval builds the candidate pool; ranking orders candidates while enforcing diversity caps and rules.
- **Where you’ll see it:** `candidate_sources` in traces, retriever dashboards, rule runbooks.
- **Why you should care:** Retrieval problems show up as missing catalog coverage; ranking problems show up as poor ordering even with good candidates.
- **Advanced details (optional):** Retrieval fanout is controlled by `POPULARITY_FANOUT` and related knobs; ranking uses MMR, caps, guardrails.

### Personalization

- **What it is:** A lightweight tag profile per user built from recent events; overlapping tags boost candidates.
- **Where you’ll see it:** Trace reasons (`personalization`), env vars (`PROFILE_BOOST`), personalization dashboards.
- **Why you should care:** Personalized share is a key KPI and guardrail; too low means bland lists, too high can hurt exploration.
- **Advanced details (optional):** Profile knobs include `PROFILE_WINDOW_DAYS`, `PROFILE_MIN_EVENTS_FOR_BOOST`, `PROFILE_STARTER_BLEND_WEIGHT`.

### MMR & caps

- **What it is:** Maximal Marginal Relevance plus brand/category caps that inject diversity so one brand/category doesn’t dominate.
- **Where you’ll see it:** Guardrail reports, diversity dashboards, env profiles.
- **Why you should care:** Diversity guardrails rely on these knobs; tightening caps can reduce CTR, loosening can degrade long-tail fairness.
- **Advanced details (optional):** `MMR_LAMBDA` controls the trade-off; caps use `BRAND_CAP`/`CATEGORY_CAP`. Exposed via env overrides and admin APIs.

### Guardrails

- **What it is:** Automatic checks (NDCG/MRR floors, coverage targets, starter-profile requirements) defined in `guardrails.yml`.
- **Where you’ll see it:** `make scenario-suite`, `analysis/reports/*/guardrail_summary.json`, CI logs, `docs/simulations_and_guardrails.md`.
- **Why you should care:** Guardrails block unsafe changes before they hit production.
- **Advanced details (optional):** Each customer/namespace can override thresholds; scenario outputs live under `analysis/evidence/`.

### Rules & overrides

- **What it is:** Manual pin/boost/block operations that merchandising or ops teams apply per namespace/surface.
- **Where you’ll see it:** `/v1/admin/rules`, `/v1/admin/manual_overrides`, `docs/rules_runbook.md`, decision traces (`trace.extras.policy`).
- **Why you should care:** Rules let humans enforce campaigns or safety constraints; guardrails verify their effect before rollout.
- **Advanced details (optional):** Overrides compile to rules internally; telemetry surfaces via Prometheus (`policy_rule_actions_total`).

### Bandits

- **What it is:** Contextual multi-armed bandits that choose among policies (e.g., “dense personalization” vs “high diversity”) based on rewards.
- **Where you’ll see it:** `/v1/bandit/*` endpoints, experiments dashboards, analytics plans (`analytics/bandit_*.md`).
- **Why you should care:** Bandits automate experimentation and can gradually move traffic to winning policies.
- **Advanced details (optional):** Policies live in `/v1/bandit/policies`; rewards come from `/v1/bandit/reward`. Use guardrails alongside bandits to enforce safety.

[Back to index](#index)

---

## 2. Metrics primer

### NDCG (Normalized Discounted Cumulative Gain)

- **What it is:** Measures how well the recommendation order matches ground-truth relevance. Higher NDCG means relevant items appear earlier.
- **Where you’ll see it:** `analysis/quality_metrics.json`, tuning dashboards, guardrail reports.
- **Why you should care:** NDCG is often the headline “quality” metric; regressions block rollouts.
- **Advanced details (optional):** Items at the top are weighted exponentially (rank 1 vs rank 5). Guardrails typically expect NDCG lift ≥ +10%.

### MRR (Mean Reciprocal Rank)

- **What it is:** Looks at the position of the first relevant item: `1 / rank`.
- **Where you’ll see it:** Starter-profile guardrails, tuning harness summaries.
- **Why you should care:** Ensures users see at least one great recommendation near the top, critical for cold-start scenarios.
- **Advanced details (optional):** Averaged across users/segments; used heavily in scenario S7 checks.

### Segment lift

- **What it is:** Percent improvement of a metric for a cohort vs. baseline (e.g., `((new - baseline) / baseline) * 100`).
- **Where you’ll see it:** `analysis/quality_metrics.json` under `segments`, tuning reports, CI logs.
- **Why you should care:** Prevents improvements for one cohort from harming another (especially new vs returning users).
- **Advanced details (optional):** Guardrails often require lift ≥ +10% for specific segments.

### Catalog coverage

- **What it is:** Fraction of the catalog that appears in the candidate pool over a window.
- **Where you’ll see it:** Guardrail dashboards, `/analysis/results/*_warm_quality.json`, Prometheus metrics.
- **Why you should care:** Low coverage means RecSys is fixating on a few items; it hurts discovery and revenue.
- **Advanced details (optional):** Guardrails often check coverage ≥ 0.60; tune `POPULARITY_FANOUT` or retriever weights if coverage dips.

### Long-tail share

- **What it is:** Portion of impressions allocated to items outside the top popularity decile.
- **Where you’ll see it:** Diversity dashboards, guardrail summaries.
- **Why you should care:** Ensures the system explores beyond bestsellers and keeps merchants happy.
- **Advanced details (optional):** Often paired with catalog coverage; raising MMR diversity or fanout can lift long-tail share.

### Diversity index

- **What it is:** Entropy-style metrics derived from MMR/cap telemetry (`policy_mmr_selected_count`, cap hit counters).
- **Where you’ll see it:** Diversity validation playbook (`analytics/diversity_validation.md`), Grafana panels.
- **Why you should care:** Low diversity can trigger guardrails and frustrate shoppers with repetitive lists.
- **Advanced details (optional):** Entropy calculations live in nightly jobs; adjust `MMR_LAMBDA` and caps to influence them.

### Determinism

- **What it is:** Re-running the same request should yield identical rankings when the namespace is frozen.
- **Where you’ll see it:** `make determinism`, `.github/workflows/determinism.yml`, `analysis/scripts/check_determinism.py`.
- **Why you should care:** Determinism is vital for audits, customer trust, and reproducibility.
- **Advanced details (optional):** Store baseline payloads under `analysis/results/determinism_baseline.json`; guardrails compare current runs to these baselines.

---

[Back to index](#index)

## 3. Guardrails in context

### Guardrails in practice

- **What it is:** YAML policies interpreted by the simulation harness before a config ships.
- **Where you’ll see it:** `guardrails.yml`, CI logs, simulation reports under `analysis/reports/`.
- **Why you should care:** Guardrails answer questions like:
  - “Did the starter-profile scenario still produce ≥4 unique categories and MRR ≥ 0.2?”
  - “Did long-tail items receive enough exposure?”
  - “Do any rules introduce dead ends?”
- **Advanced details (optional):** See `docs/simulations_and_guardrails.md` for fixtures + CLI usage. Business stakeholders can read `docs/business_overview.md#safety-and-guardrails` for the narrative.

---

## 4. Glossary

### Quick glossary highlights

- **ALS** – Our embedding similarity signal (cosine similarity against anchors). Despite the name, it isn’t a matrix-factorization job here.
- **Anchors** – A user’s most recent interacted items; drive co-visitation and embeddings.
- **Caps** – Hard limits like `BRAND_CAP`/`CATEGORY_CAP` to avoid repeating the same brand/category.
- **Cold start** – Lack of history for a user/item; handled via starter profiles, embeddings, popularity.
- **Constraints** – Filters (`exclude_item_ids`, price bands, availability) applied before or during ranking.
- **Co-visitation** – “Users who touched X also touched Y shortly after.” Feeds the co-visitation signal.
- **Embeddings** – Vector representations of items used for content similarity.
- **Event** – A view/click/add/purchase record; ingested via `/v1/events:batch`.
- **Fanout** – Number of popularity candidates fetched before reranking (`POPULARITY_FANOUT`).
- **Half-life** – How fast old events decay in popularity scoring.
- **Light personalization** – Boost derived from overlap between a user’s tag profile and candidate tags.
- **Normalization** – Rescales each signal to `[0,1]` per request so blend weights stay meaningful.
- **Reasons** – Human-readable explanations per recommendation in traces/responses.
- **Top-K** – The `k` items returned after ranking, caps, and overrides.
- **UserTagProfile** – Short-lived map of tag → weight summarizing a user’s interests.
- **Windows** – Lookback durations (`COVIS_WINDOW_DAYS`, `PROFILE_WINDOW_DAYS`) that trade recency vs history.

[Back to index](#index)
