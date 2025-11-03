# Business Analysis – Remediation Tickets

## Key Findings
- Collaborative, content-based, and session candidate retrievers exist in code but currently return zero results because prerequisite data (ALS factors, tag similarity view) is missing or misconfigured.
- `content_similarity_topk.sql` references a non-existent `items_tags` relation, silently disabling tag-based candidate generation.
- The ranking pipeline limits the popularity pool to `k`, so exclusions and rule boosts have no headroom to introduce additional items—manual boosts (e.g., `watch_gps_pro`) never surface.
- Embedding-powered scoring requires running the catalog backfill tool; without it, gamma contributions remain zero and documentation leaves this implicit.
- Documentation promises world-class signals without instructing operators how to activate them, creating expectation gaps during evaluations.

## Detailed Description

We dug into the codebase to see why the demo looked “popularity only” and why the manual boost didn’t fire. Here’s what we found and how we can fix it.

1. Candidate sources stuck at zero → Functional gaps (plus missing guidance).

- internal/algorithm/engine.go already tries to pull collaborative, content, and session candidates, but the data they depend on never materialises.
- Collaborative: Store.CollaborativeTopK queries recsys_user_factors / recsys_item_factors (see internal/store/collaborative.go and queries/user_factor_topk.sql). Those tables aren’t created or populated by any existing job, so we’ll always log collaborative.count = 0. We either need to ship the batch job that
  trains and writes those factors, or down-scope the feature/documentation until it exists.
- Content: ContentSimilarityTopK runs queries/content_similarity_topk.sql, which references a non-existent items_tags relation. Because of the 42P01 guard, we silently return [], hence content.count = 0. Fix is straightforward: change the SQL to unnest items.tags directly (or create the view in migrations) and
  add a regression test so we don’t regress.
- Embeddings: without running make catalog-backfill (cmd/catalog_backfill), items.embedding stays NULL, so even once the SQL bug is fixed we’ll have no gamma signal. We should (a) document that backfill/refresh must run before evaluating, and (b) consider wiring the backfill job into dev bootstrap so embeddings
  exist by default.

2. Manual overrides ignored → Functional bug.
The boost rule is written correctly and stored in rules (confirmed via SQL), but the ranker never sees the boosted item because mergeCandidates caps the popularity pool at exactly k. When RULE_EXCLUDE_EVENTS later removes a couple of those items, we end up with fewer than k candidates and no room for the
boosted SKU to be reconsidered.

- Fix: respect the configured fanout (POPULARITY_FANOUT) all the way through candidate merging. A minimal patch is to let mergeCandidates take fetchK (or at least k + exclusions) instead of slicing to k, so we always have slack for rules/MMR. We should add an integration test that creates a manual boost and
  asserts the boosted item rises to the top.
- Optional: after we touch mergeCandidates, record the actual candidate count in trace/audit so we can watch for similar shortages.

3. Documentation follow-up.
Docs promise ALS/content signals, but the code path requires extra setup. Once we fix the SQL bug and widen the candidate pool, we still need instructions that (a) catalog backfill must run to generate embeddings, and (b) collaborative factors need an external training job (or we should remove that promise
until the job ships). We should also note that rules are surface-scoped and cached for 2s, so merchants know to wait a heartbeat after creating overrides.

Next actions I’ll take as developer

1. Patch content_similarity_topk.sql to unnest items.tags, add unit test in internal/store ensuring content candidates emit results.
2. Implement the candidate-reserve fix in mergeCandidates, plus a regression test in internal/algorithm covering manual boosts.
3. Decide on collaborative-plan: either wire up a simple factor trainer (even a nightly job) that writes to recsys_*_factors, or tone down README claims until we have it.
4. Update CONFIGURATION.md / README with a “bring the signals online” checklist (run catalog backfill, run factors job, confirm candidate_source metrics).

Once those are in place, re-run the evaluation script—candidate_sources should show non-zero counts, and the boosted item should appear in ranked output.

## Tickets
- [x] **RS-001 – Repair content-similarity SQL query**  
  Update `queries/content_similarity_topk.sql` (and add regression coverage) to read tags directly from `items.tags` so content-based candidates populate as designed. Validate via unit test and manual API call that `candidate_sources.content.count` becomes non-zero.

- [x] **RS-002 – Preserve candidate fanout for rule overrides**  
  Adjust `Engine.mergeCandidates` (and related call sites) to retain at least the configured `POPULARITY_FANOUT` before exclusions/MMR, ensuring manual boosts and caps have inventory to work with. Add integration test proving a boost rule elevates the target item.

- [ ] **RS-003 – Provision collaborative factors pipeline**  
  Deliver a batch job (or documented process) that writes user/item factors into `recsys_user_factors` & `recsys_item_factors`, or gate the collaborative retriever behind a feature flag with degraded docs. Acceptance: `candidate_sources.collaborative.count > 0` in a seeded environment.

- [ ] **RS-004 – Document signal activation workflow**  
  Update README/CONFIGURATION to include a “Bring signals online” checklist (catalog backfill, factor job status, rule cache behaviour). Call out prerequisites for embeddings, collaborative data, and rule propagation so evaluators can reproduce expected behaviour.
