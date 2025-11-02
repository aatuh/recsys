# Candidate Generation Audit (RT-1A)

## Current pipeline (October 2025)

1. **Popularity retriever**
   - Files: `internal/store/popularity.go`, `internal/algorithm/engine.go`.
   - Inputs: exponentially-decayed interaction counts (`POPULARITY_HALFLIFE_DAYS`, `POPULARITY_FANOUT`).
   - Output: global trending items; primary driver when user history is sparse.

2. **Co-visitation retriever**
   - Files: `internal/store/covisitation.go`, consumed via `CandidateData.Cooc`.
   - Inputs: event co-occurrence within `COVIS_WINDOW_DAYS`.
   - Output: items frequently interacted with alongside anchors; currently folded into blend (beta weight).

3. **Embedding retriever**
   - Files: `internal/store/embedding.go`, ALS factor store.
   - Inputs: matrix factorization vectors stored in Postgres.
  - Output: similar items via latent factors; feeds blend gamma component.

4. **Profile personalization**
   - Files: `internal/algorithm/personalization.go`.
   - Inputs: user profile aggregates (`PROFILE_WINDOW_DAYS`, `PROFILE_BOOST`).
   - Output: boosts scoring but does not introduce new candidates; relies on candidate pool from steps above.

5. **Ranking & diversity layer**
   - File: `internal/algorithm/engine.go`.
   - Responsibilities: merge weighted scores (alpha/beta/gamma), apply MMR/caps, produce reasons.

### Observed gaps vs. RT-1 goals
- No standalone collaborative top-N retriever per user (ALS factors only influence scoring).
- No content-based retriever leveraging rich item metadata/embeddings controllable by shop pipeline.
- No session-sequence retriever beyond simple co-vis heuristics.
- Observability now instrumented: `candidate_source_metrics` logs and audit extras expose per-source coverage/latency (see `analytics/retriever_dashboard.md`).

## Roadmap alignment
1. **RT-1B:** Introduce `CollaborativeRetriever` service that queries ALS user vectors -> top-N unseen items, available via API.
2. **RT-1C:** Stand up content-based similarity index (tags/embeddings) for cold-start coverage.
3. **RT-1D:** Prototype session-aware recency retriever (last-k events -> next best predictions).
4. **RT-1E:** Implement orchestrator that unions/weights source candidates, ensures dedupe, attaches source metadata.
5. **RT-1F:** Add coverage/latency metrics per source, surface via dashboards and alerts.
