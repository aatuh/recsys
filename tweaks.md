# Recommendation Tweaks Backlog

- [x] **RT-1A:** Audit current candidate generation and document existing popularity path and ranking blend.
- [x] **RT-1B:** Build collaborative-filtering retriever that surfaces top-N items per user from recent interactions (e.g., matrix factorization or co-visitation index).
- [x] **RT-1C:** Deliver content-based retriever using item tags/embeddings with similarity search to cover cold-start products.
- [x] **RT-1D:** Add short-term sequence/recency retriever for session-aware candidates (last-k viewed -> next best).
- [x] **RT-1E:** Implement merge/deduping layer that unions all retrievers, scores candidates, and exposes sampling knobs per source.
- [x] **RT-1F:** Ship integration tests and dashboards confirming coverage & latency across all retrievers.

## 2. Implement user-level seen-item exclusions
- [x] **RT-2A:** Define recency windows and event-type policy for “seen” cache (view, click, add, purchase).
- [x] **RT-2B:** Create fast lookup store (e.g., Redis/in-memory) populated from user telemetry to retrieve seen IDs at request time.
- [x] **RT-2C:** Update recommendation request pipeline to filter seen IDs before ranking, with overrides for merchandising.
- [x] **RT-2D:** Add monitoring to track exclusion hit rate and fallback behaviour.

## 3. Retune and monitor blend weights
- [x] **RT-3A:** Set up offline evaluation harness comparing blend configurations on historical data.
- [x] **RT-3B:** Design online A/B test plan for updated blend weights (alpha/beta/gamma) and establish success metrics.
- [x] **RT-3C:** Implement automated deployment of chosen blend parameters via config service or feature flags.
- [x] **RT-3D:** Instrument overlap/personalization metrics dashboards with alerts when personalization lift drops.

## 4. Add diversity constraints to the ranker
- [x] **RT-4A:** Formalize diversity rules (brand/category caps, price bands, recency windows) with product requirements.
- [x] **RT-4B:** Implement MMR or constraint-based reranker that applies those rules post-scoring.
- [x] **RT-4C:** Provide per-surface configuration so product can tune caps and diversity knobs.
- [x] **RT-4D:** Validate diversity impact via simulation + live metrics (intra-list diversity, CTR).

## 5. Enrich catalog metadata with embeddings
- [x] **RT-5A:** Inventory current product attributes and identify gaps needed for content-based models.
- [x] **RT-5B:** Generate textual/visual embeddings for products using chosen ML pipeline and store in feature service.
- [x] **RT-5C:** Extend item ingestion API to accept embeddings and additional attributes.
- [x] **RT-5D:** Backfill existing catalog with new metadata and create freshness jobs to keep embeddings current.

## 6. Ship a cold-start candidate generator
- [x] **RT-6A:** Define “fresh arrival” eligibility (age threshold, availability) and exposure budget.
- [x] **RT-6B:** Build freshness index that tracks new SKUs and serves top candidates per segment.
- [x] **RT-6C:** Integrate cold-start pool into recommendation pipeline with adjustable weight.
- [x] **RT-6D:** Measure cold-start CTR/conversion and tune exposure accordingly.

## 7. Explore new items with bandit policies
- [x] **RT-7A:** Specify exploration framework (slot count, eligible candidates, reward signal).
- [x] **RT-7B:** Implement Thompson/UCB policy service consuming cold-start candidates and logging decisions.
- [x] **RT-7C:** Wire reward feedback loop from telemetry to update bandit posteriors.
- [x] **RT-7D:** Run controlled experiment to verify exploration doesn’t harm core metrics.

## 8. Instrument cold-start exposure and performance KPIs
- [x] **RT-8A:** Define KPI taxonomy (impressions, CTR, CVR, revenue) specifically for cold-start inventory.
- [x] **RT-8B:** Implement event enrichment tagging recommended items as “cold_start” for analytics pipelines.
- [x] **RT-8C:** Build dashboards and alerting on cold-start performance, including per-surface breakdowns.
- [x] **RT-8D:** Expose tooling for manual boosts/suppression with audit logs for merchandising teams.
