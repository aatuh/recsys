# TKT-01A – New User Personalization Regression

## Segment Impact
- `new_users` miss the ≥10% lift target: NDCG@10 drops −27.6% and MRR@10 drops −26.0% compared to the popularity baseline while only Recall@20 improves (+26.5%) (`analysis_v2/quality_metrics.json:27-45`).
- Other cohorts exceed +16% lift, so the regression is isolated to onboarding, confirming this is not a global ranking failure (`analysis_v2/quality_metrics.json:48-88`).

## Trace & Candidate Observations
- Recommendation traces for `new_users` lean almost entirely on popularity: the two sampled users show `popularity.count = 300` vs `content.count ≈ 17` with zero collaborative or session candidates (`analysis_v2/evidence/recommendation_samples_after_seed.json`, users `user_0047`, `user_0092`).
- No starter profile is attached to the trace extras, so `starter_profile` is absent from the debug payload despite the segment being `new_users` (`analysis_v2/evidence/recommendation_samples_after_seed.json`).
- Scenario S7 confirms the starter profile map is empty (`Starter profile tags=[]`), so brand/category presets never seed the cold-start list (`analysis_v2/scenarios.csv:8`).
- Anchors collected for `user_0047` include five distinct items, which means our personalization layer treats the user as fully “warmed up” and applies the full 0.7 boost (anchors ≥ `PROFILE_MIN_EVENTS_FOR_BOOST`) even though the user only has a handful of recent interactions.

## Likely Root Causes & Hypotheses
1. **Starter profile gating is too strict.**  
   We only build a starter profile when `userHasHistory` returns false (`api/internal/services/recommendation/onboarding.go:33-84`). Because it checks for *any* recent event within the 30‑day profile window (`PROFILE_WINDOW_DAYS=30`, `api/.env:19-23`), new users with one or two post-split events never receive the curated tag seed and therefore rely on sparse organic history.
2. **Cold-start boosts run at full strength on noisy anchors.**  
   With anchors ≥5, the personalization code applies the full multiplier (0.7) because the attenuation threshold is `PROFILE_MIN_EVENTS_FOR_BOOST=3` (`api/internal/algorithm/engine.go:1222-1272`, `api/.env:20-23`). Those boosted candidates shuffle the diversified list (MMR λ=0.3, `api/.env:7-13`) away from the limited ground-truth items that define the evaluation labels.
3. **Candidate breadth for `new_users` is shallow.**  
   Content-based retrieval averages only 17.5 items per request while we never backfill with collaborative or session candidates, so the 300 popularity candidates dominate the merge. Any mistakes in the popularity pool translate directly into recall loss for sparse-history users.

## Instrumentation Gaps
- We do not publish whether a starter profile fired, why it was skipped, or how many anchors/feed items were available. Adding per-request trace counters (`starter_profile_applied`, `anchor_count`, `profile_event_count`) plus Prometheus tallies would let us alert on missing seeds.
- Logging when the `userHasHistory` gate short-circuits (with the observed event count) would confirm whether the rule is misclassifying new users.
- Scenario automation should assert that S7 returns a non-empty `starter_profile` payload and at least one item tagged by the preset so we catch regressions automatically.

## Recommended Next Steps (feed TKT-01B/TKT-01C)
1. Rework `userHasHistory` to allow the starter preset whenever recent events `< PROFILE_MIN_EVENTS_FOR_BOOST`, ensuring users with <3 interactions still get the curated seed.
2. Retune the cold-start blend for `new_users`: consider lowering MMR λ for this segment and/or defaulting to the starter profile when collaborative/session sources are empty.
3. Add the instrumentation above and extend the evaluation harness to assert on starter profile usage so we can graph adoption and catch future regressions.
