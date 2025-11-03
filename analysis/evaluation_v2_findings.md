• Built a repeatable evaluation harness, seeded the live API-only dataset, and captured all required deliverables plus evidence for every scenario.

Artifacts

- analysis/scripts/seed_dataset.py:1 seeds 320 items / 120 users / 5.2k events via official ingest endpoints with fixed seed metadata.
- analysis/scripts/run_quality_eval.py:1 computes baseline vs system metrics and persists analysis/quality_metrics.json:1.
- analysis/scripts/run_scenarios.py:1 automates S1–S10 and writes analysis/scenarios.csv:1 plus per-scenario evidence blobs.
- analysis/config_matrix.md:1 maps each configurability control to Supported / Partial / Missing status.
- analysis/report.md:1 contains the executive FAIL verdict, evidence index, and remediation priorities.

Key Findings

- Overall lift over popularity exists but absolute quality is weak (NDCG@10 = 0.076) and new users regress sharply (−35% NDCG, −30% MRR) per analysis/
quality_metrics.json:2-46.
- Strict policy controls fail: include filters, block rules, manual boosts, pins, and brand whitelists all return FAIL in analysis/scenarios.csv:2-10
with supporting traces (see config matrix rows analysis/config_matrix.md:3-11).
- Diversity knob (mmr_lambda) is the lone working override (S4 PASS), but merchandising/exposure knobs show no effect, leaving business teams without
governance levers.

Next Steps

1. Repair request-time constraint enforcement so constraints.include_* and rule-based blocks actually gate candidates.
2. Reconnect the merchandising rule + manual override pipeline to the final re-ranker (boosts, pins, trade-offs), then re-run the scenario suite.
3. Address new-user onboarding so quality meets baseline (e.g., seeded profiles or curated starters) before attempting another evaluation.
