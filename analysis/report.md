# Recsys Evaluation Report

**Verdict: FAIL**

## Executive Summary
- Quality lift vs popularity baseline is modest (NDCG@10 +12%, Recall@20 +49%, MRR@10 +40%) yet absolute levels remain low (NDCG@10 = 0.076) with severe regression for new users (−35% NDCG, −30% MRR) — see `analysis/quality_metrics.json`.
- Core filter compliance is broken: tag include/exclude constraints, brand whitelists, and price-like gating fail, causing business-rule leakage in S1/S2/S6 (`analysis/scenarios.csv`, evidence files).
- Merchandising controls (manual boosts, pins, tag boosts) have no discernible effect on ranking (S3, S5, S8, S9), eliminating configurability needed for campaigns.
- Cold-start defaults and explainability work (diverse top-10, reason + explain blocks), but they cannot compensate for missing policy enforcement.
- Without reliable filters/overrides, the system cannot support real-world governance or merchandising and must be considered unsafe for production launches.

## Quality Findings
- Overall lift above popularity baseline meets minimum thresholds, but relevance remains weak in absolute terms. Coverage is strong (57% catalog coverage, long-tail share 52%), yet personalization underperforms for new users (quality below baseline) which risks churn.
- Segment analysis shows only 3/5 segments meet lift targets; “new_users” regress on every metric, indicating missing onboarding logic or cold-start heuristics.
- Diversity knob (`overrides.mmr_lambda`) reduces intra-list similarity without measurable quality loss (S4), demonstrating the underlying re-ranker responds to algorithmic overrides when they reach the re-ranking layer.

## Configurability & Policy Gaps
- **Strict filters**: `constraints.include_tags_any` ignored requested tags; filtered calls returned mixed categories (e.g., `analysis/evidence/scenario_s1_response.json`). Hard block rules on `high_margin` tags also fail to remove disallowed items (`scenario_s2_block_high_margin.json`).
- **Manual overrides & rules**: Boosts and pins leave rankings unchanged despite successful API responses (`scenario_s3_boost.json`, `scenario_s5_pin.json`). BOOST, PIN, and BLOCK actions appear to be no-ops.
- **Whitelist enforcement**: Brand-tag filters return cross-brand items (`scenario_s6_whitelist.json`), signalling missing attribute filters.
- **Multi-objective controls**: Increasing `boost_value` for `high_margin` items shows zero movement in margin share or NDCG (`scenario_s9_tradeoff.json`), so no trade-off curve exists.
- **New item governance**: Fresh items already flood recommendations (88% exposure) and boosting them provides no additional control (`scenario_s8_new_item.json`), indicating lack of exploration tuning and override effect.

## Explainability & Traceability
- `include_reasons` + `explain_level=full` returns blend breakdown, MMR parameters, and model_version (`scenario_s10_explainability.json`). Trace data references blend contributions, suggesting the audit pipeline works.
- Saved trace snippets (e.g., `analysis/evidence/recommendation_samples_after_seed.json`) confirm reason codes and weights are exposed but lack rule-related tags because rules never fire.

## Required Remediation for PASS
1. **Fix filter pipeline**: Ensure `constraints` (tags, price) are enforced before ranking, with regression tests mirroring S1/S6.
2. **Enable merchandising rules**: Investigate why `POST /v1/admin/rules` and manual overrides do not affect scoring; repair integration between rule engine and final response (pins, blocks, boosts).
3. **Validate override knobs**: Once rules work, generate unit/integration coverage demonstrating monotonic boosts, deterministic pins, and margin vs relevance trade-offs.
4. **Address new user regression**: Add fallback personalization (e.g., trait-based profiles or curated starters) to recover NDCG/MRR for `new_users` cohort.

## Evidence Index
- Dataset seed manifest & samples — `analysis/evidence/seed_manifest.json`, `analysis/evidence/seed_samples.json`.
- Quality metrics (aggregate & per segment) — `analysis/quality_metrics.json`.
- Scenario evidence (S1–S10) — `analysis/evidence/scenario_*.json`.
- Scenario outcomes table — `analysis/scenarios.csv`.
- Config matrix — `analysis/config_matrix.md`.

• Communicating Gaps to Developers

- Create blocker tickets for policy failures – reference analysis/scenarios.csv:2-7 and attach evidence (e.g., analysis/evidence/
  scenario_s1_response.json). Summaries like “Tag include filter ignored; mixed categories returned despite constraint” give developers a reproducible
  repro path (/v1/recommendations payload + org header).
- Open merchandising-rule regression issues – cite analysis/config_matrix.md:3-11 rows for boosts/pins/blocks and link the before/after traces in
  analysis/evidence/scenario_s2_block_high_margin.json etc. Include expected vs observed ranks to make tests easy to automate.
- Log a quality gap ticket for new-user cohort – pull the exact lift numbers from analysis/quality_metrics.json:27-46 so product teams see the regression
  severity; propose hypotheses (no onboarding profile, cold-start strategy missing).
- Document working pieces separately – note that diversity override and explainability passed (analysis/evidence/
  scenario_s4_diversity.json, ...scenario_s10_explainability.json). This helps scope fixes to the broken subsystems rather than re-opening everything.
- Add acceptance tests to definition of done – suggest codifying S1–S9 payloads as API contract tests so once fixed they stay fixed; point developers to
  the runnable scripts in analysis/scripts/ they can reuse locally.
