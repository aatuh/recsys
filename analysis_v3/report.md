# Recsys Evaluation Report

**Verdict: PASS**

## Executive Summary
- Overall lift vs. the popularity baseline now meets the PASS rubric with margin: NDCG@10 +40.8%, Recall@20 +38.5%, MRR@10 +77.1% (`analysis/quality_metrics.json`).
- The `new_users` cohort recovered to +23.9% NDCG lift and +28.4% MRR lift after enabling starter-profile blending plus diversity guardrails.
- Catalog coverage holds at 65.9% of the 320-item catalog with 49.2% long-tail share and low intra-list similarity, showing broad-yet-relevant exposure.
- Policy controls (filters, boosts, pins, whitelists, diversity knobs) remain deterministic with full trace evidence (`analysis/scenarios.csv`, `analysis/config_matrix.md`), and cold-start guardrails in S7 now enforce ≥0.2 MRR and ≥4 categories per list.
- Explainability, determinism, and new-item exposure artifacts confirm the system is auditable and configurable for real business use.

## Quality Findings
- System metrics: NDCG@10 0.0953, Recall@20 0.1196, MRR@10 0.3632 vs baseline 0.0677 / 0.0864 / 0.2051, comfortably above the ≥10% lift thresholds.
- Segments: all five cohorts exceed +20% NDCG lift (`new_users`: +23.9%, `trend_seekers`: +47.7%, `power_users`: +22.0%, `niche_readers`: +68.3%, `weekend_adventurers`: +64.3%). MRR lift ranges from +28% to +133%.
- Coverage/diversity: 211 unique items (65.9% catalog coverage) with 115 long-tail uniques (49.2% share); intra-list similarity@10 = 0.105 and novelty@20 = 0.521 show balanced variety.
- Diversity override (`mmr_lambda=0.0`) reduced similarity 0.079→0.057 while boosting NDCG 0.000→0.150 (`analysis/evidence/scenario_s4_diversity.json`), confirming controlled diversity budgets.

## Configurability & Policy Controls
- Tag includes/excludes, brand whitelists, and block rules show zero leakage with auditable traces (`analysis/evidence/scenario_s1_response.json`, `analysis/evidence/scenario_s2_block_high_margin.json`, `analysis/evidence/scenario_s6_whitelist.json`).
- Manual boosts/pins remain monotonic (rank 2→0) and the margin-vs-relevance curve stays smooth (margin share 0.385→0.406 with NDCG within ±0.01) per `analysis/evidence/scenario_s3_boost.json` and `analysis/evidence/scenario_s9_tradeoff.json`.
- New-item exposure boosts raise surfaced lists from 82% to 90% with rule counters logging the injection (`analysis/evidence/scenario_s8_new_item.json`).

## Cold-Start, New Items & Determinism
- Cold-start batch hit avg MRR 0.250 and 5.6 categories/list while satisfying new harness thresholds (S7 evidence) and holdout metrics now show positive lift.
- Determinism check across five identical requests still yields the same slate and trace IDs (`analysis/evidence/determinism_check.json`).
- Boosted new items appear without constraint leakage (S8), maintaining exploration capability.

## Explainability & Traceability
- `include_reasons=true` + `explain_level=full` responses include model version, per-item reason codes, and rule applications (`analysis/evidence/scenario_s10_explainability.json`).
- Recommendation samples (`analysis/evidence/recommendation_samples_after_seed.json`) and scenario traces capture config IDs plus source metrics for audits.

## Evidence Index
- Dataset seed manifest & samples — `analysis/evidence/seed_manifest.json`, `analysis/evidence/seed_samples.json`.
- Quality metrics (aggregate & per-segment) — `analysis/quality_metrics.json`.
- Scenario outcomes table — `analysis/scenarios.csv`; detailed payloads — `analysis/evidence/scenario_*.json`.
- Determinism validation — `analysis/evidence/determinism_check.json`.
- Config matrix snapshot — `analysis/config_matrix.md`.
