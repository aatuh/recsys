# Concepts & Metrics Primer

Read this when you need plain-language definitions of the jargon that appears throughout the RecSys documentation. It consolidates the glossary, explains how we measure success, and clarifies how guardrails fit in.

---

## 1. Core concepts

- **Namespace** – A logical tenant/surface bucket. Every request specifies it explicitly so data, guardrails, and overrides stay isolated.
- **Org** – A company/customer identifier passed via `X-Org-ID`. One org can own many namespaces.
- **Candidate** – An item ID that is eligible to be ranked. Candidate pools come from popularity, collaborative filtering, content similarity, and session heuristics before we pick the final top-K.
- **Signals** – Independent evidence that scores a candidate (popularity, co-visitation, embeddings, personalization boosts, etc.). Signals are normalized to `[0,1]` within a request before blending.
- **Blended scoring** – A weighted sum of normalized signals: `alpha*pop + beta*co_vis + gamma*embed + …`. Overrides let you change weights per call.
- **Retrieval vs. ranking** – Retrieval builds the candidate pool (fanout) from the data lake. Ranking (MMR + caps + overrides) orders those candidates and enforces constraints.
- **Personalization** – A lightweight user tag profile built from recent events. Candidates that share tags with the profile receive a boost controlled by `PROFILE_BOOST` and related knobs.
- **MMR & caps** – Maximal Marginal Relevance plus brand/category caps inject diversity so the top of the list is not dominated by one brand or topic.
- **Guardrails** – Automatic checks (NDCG/MRR floors, coverage targets, diversity expectations) that block configs or deployments when the corpus looks risky. Implemented via simulations and CI jobs (`docs/simulations_and_guardrails.md`).
- **Rules & overrides** – Manual pin/boost/block operations that merchandising or ops teams can apply per namespace, surface, or segment. Guardrails verify their effect before rollout.
- **Bandits** – Optional contextual multi-armed bandit that chooses among policies (e.g., “dense personalization” vs “high diversity”) based on downstream rewards.

---

## 2. Metrics primer

These show up in tuning dashboards, guardrails, and CI reports.

- **NDCG (Normalized Discounted Cumulative Gain):** Measures how well the order of recommendations matches ground-truth relevance. The closer items your users actually clicked/purchased appear to the top, the closer NDCG is to 1.0. Dropping a relevant item from rank 1 to rank 5 hurts more than shifting items lower in the list.
- **MRR (Mean Reciprocal Rank):** Looks at the position of the first relevant item in each list: `1/rank`. Averaged across users/segments. Useful for “did we show *anything* they'll love near the top?” guardrails (e.g., the starter-profile scenario).
- **Segment lift:** Percentage improvement of a metric (NDCG, conversion, etc.) for a given cohort vs. baseline. Example: `((new_ndcg - baseline_ndcg) / baseline_ndcg) * 100`. Ensures no cohort regresses badly when tuning.
- **Catalog coverage:** Fraction of the catalog that appears in the eligible candidate pool over a window. We use it to verify the system doesn’t overfit to a handful of products. Guardrails often check coverage ≥ 0.60.
- **Long-tail share:** Portion of impressions allocated to items outside the top popularity decile. High share = better breadth; low share may require raising `POPULARITY_FANOUT` or increasing MMR.
- **Diversity index:** Derived from MMR/cap telemetry (`policy_mmr_selected_count`, cap hit counters). Ensures surfaces show ≥N categories/brands within the top K.
- **Determinism:** Re-running the same request 10× should yield identical item order when the namespace is frozen. Guardrails call `make determinism` (local repo command) to ensure reproducibility for audits.

---

## 3. Guardrails in context

Guardrails are small YAML policies (`guardrails.yml`) interpreted by the simulation harness before a config ships. They answer questions like:

- “Did the starter profile scenario still produce ≥4 unique categories and MRR ≥ 0.2?”
- “Did long-tail items receive at least 30% exposure during rerank?”
- “Do any rules break tie-breaking or introduce dead ends?”

Failures block merges/rollouts until someone tunes the offending knobs or rules. Learn how to define and run them in `docs/simulations_and_guardrails.md`. Business stakeholders can read `docs/business_overview.md#safety-and-guardrails` for the narrative version.

---

## 4. Glossary

**ALS** – The embedding similarity signal (the “gamma” term). In our system it measures cosine similarity versus anchors; it is not a matrix-factorization job despite the familiar acronym.

**Anchors** – A user’s most recent interacted items (within `COVIS_WINDOW_DAYS`). Used to compute co-visitation edges and embedding similarity for the current request.

**Blended scoring** – Converts multiple normalized signals into one score: `alpha*pop_norm + beta*co_vis_norm + gamma*embed_norm`. Falls back to popularity if other weights are zero.

**Caps** – Hard limits such as `BRAND_CAP`/`CATEGORY_CAP` that prevent too many items from the same brand/category appearing in the final list.

**Cold start** – Lack of historical data for a user or item. Starter profiles, embeddings, and popularity help fill the gap until enough events accrue.

**Constraints** – Filters applied before/during ranking (`exclude_item_ids`, exclude purchased, price/availability filters, etc.).

**Co-visitation** – “Users who touched X also touched Y shortly after.” Aggregated globally per namespace within a rolling window; used as the beta term.

**Embeddings** – 384‑dimensional vectors that represent item meaning; cosine distance measures similarity.

**Event** – Evidence about interactions (view, add-to-cart, purchase). Weighted via `EVENT_TYPE_WEIGHTS`.

**Fanout** – Number of popularity candidates fetched before reranking (`POPULARITY_FANOUT`). Must exceed `k` to give downstream filters room to work.

**Half-life** – How fast old events decay in popularity scoring (e.g., 14 days halves influence).

**Light personalization** – Multiplicative boost derived from overlap between a user’s tag profile and a candidate’s tags. Controlled by `PROFILE_BOOST`, `PROFILE_WINDOW_DAYS`, and friends.

**MMR (Maximal Marginal Relevance)** – Balances relevance with novelty. `lambda=1` behaves like pure relevance; `lambda=0` maximizes diversity.

**Normalization** – Rescales each signal within the candidate set to `[0,1]` per request so weights are comparable.

**Reasons** – Audit trace labels per item (popularity, co_visitation, embedding, personalization, diversity, caps, excluded events). Only the applicable ones appear.

**Signals** – Popularity, co-visitation, embedding similarity, personalization, contextual bandit priors, etc.

**Top-K** – The K items returned after ranking, caps, and overrides.

**UserTagProfile** – Short-lived map of tag → weight summarizing a user’s recent interests. Drives personalization reasons in responses.

**Windows** – Lookback durations such as `COVIS_WINDOW_DAYS` or `PROFILE_WINDOW_DAYS`. Shorter windows favor recency; longer capture more history.
