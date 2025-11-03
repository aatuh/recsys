# Recsys Evaluation Report

**Verdict: CONDITIONAL PASS**

## Executive Summary
- Overall lift versus the popularity baseline clears targets (NDCG@10 +16%, Recall@20 +53%, MRR@10 +40%) but the new_users cohort regresses on NDCG/MRR (`analysis/quality_metrics.json`), signalling onboarding gaps.
- Catalog coverage reaches 57.5% of the 320-item catalog (target ≥60%) while long-tail share hits 49.6%; diversity knobs help, yet broader coverage remains a stretch goal.
- Policy controls (filters, block/boost/pin, whitelists) now behave deterministically with clear trace evidence (`analysis/scenarios.csv`, `analysis/config_matrix.md`).
- Business knobs (MMR lambda, margin boosts) exhibit predictable trade-offs and do not break relevance, enabling campaign tuning.
- Cold-start responses are varied and explainable, and deterministic replay holds, but production READINESS hinges on fixing the new-user regression and nudging catalog coverage past the threshold.

## Quality Findings
- Averaged metrics: system NDCG@10 = 0.079, Recall@20 = 0.133, MRR@10 = 0.287 vs popularity baseline 0.068 / 0.086 / 0.205 (`analysis/quality_metrics.json`), meeting ≥10% lift requirements.
- Segment view: trend_seekers, power_users, niche_readers, and weekend_adventurers each exceed +16% lift; new_users drops −27.6% NDCG and −26.0% MRR, pointing to missing onboarding features or overly aggressive exploration.
- Coverage and novelty: 184 unique items recommended (57.5% catalog coverage) with 101 long-tail uniques (49.6% share). Diversity remains strong (intra-sim 0.154) but coverage window should reach ≥60%.
- Diversity override (`overrides.mmr_lambda=0.1`) lowers similarity from 0.109→0.070 while improving NDCG (+12.7 points) (`analysis/evidence/scenario_s4_diversity.json`), confirming balanced diversification.

## Configurability & Policy Controls
- Tag includes/excludes, brand whitelists, and manual overrides all pass scenario checks with zero leakage (`analysis/evidence/scenario_s1_response.json`, `.../scenario_s2_block_high_margin.json`, `.../scenario_s6_whitelist.json`).
- Boosts and pins respond monotonically (rank 2→0), and multi-objective experiments show smooth margin gain (0.386→0.408) with manageable NDCG drift (`analysis/evidence/scenario_s3_boost.json`, `.../scenario_s9_tradeoff.json`).
- New-item exposure boosts raise recommendation incidence from 82% to 100%, and trace counters reflect applied rules (`analysis/evidence/scenario_s8_new_item.json`).

## Cold-Start, New Items & Determinism
- Cold-start user receives 10 items spanning 8 categories with starter profile telemetry exposed (`analysis/evidence/scenario_s7_cold_start.json`).
- Determinism check across five identical recommendation calls returned the same ranked list (`analysis/evidence/determinism_check.json`), supporting replay/audit workflows.
- New item treatments combined with boosts yield controllable exposure without constraint leakage (S8 evidence).

## Explainability & Traceability
- `include_reasons=true` with `explain_level=full` returns model_version, per-item explain blocks, and segment/profile identifiers (`analysis/evidence/scenario_s10_explainability.json`).
- Recommendation samples (`analysis/evidence/recommendation_samples_after_seed.json`) show consistent reason weights and trace extras for rule execution, aiding audit trails.

## Required Remediation for PASS
1. **Fix new-user regression**: Introduce starter profiles or blended popularity traits so `new_users` recover ≥10% lift on NDCG/MRR; validate with segment-level replay (`analysis/quality_metrics.json`).
2. **Increase catalog coverage**: Tune exploration or fan-out to exceed the ≥60% coverage target while keeping long-tail share above 20%.
3. **Institutionalize regression checks**: Automate S1–S10 payloads plus determinism replay in CI to prevent future policy regressions.

## Evidence Index
- Dataset seed manifest & samples — `analysis/evidence/seed_manifest.json`, `analysis/evidence/seed_samples.json`.
- Quality metrics (aggregate & per-segment) — `analysis/quality_metrics.json`.
- Scenario outcomes table — `analysis/scenarios.csv`; detailed payloads — `analysis/evidence/scenario_*.json`.
- Determinism validation — `analysis/evidence/determinism_check.json`.
- Config matrix snapshot — `analysis/config_matrix.md`.
