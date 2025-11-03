# Recsys Evaluation Remediation Tickets

## Description
This ticket set codifies the corrective work required to turn the latest evaluation findings (see `analysis_v2/report.md`) into production-ready improvements. Each epic clusters related fixes so we can parallelize engineering, modeling, and QA efforts while keeping the evaluation acceptance criteria in scope.

## Problem Statement
- Overall ranking quality lifts meet minimum thresholds, but the `new_users` cohort regresses on NDCG and MRR, jeopardizing onboarding experience (`analysis_v2/quality_metrics.json:26-45`).
- Catalog coverage stalls at 57.5% of the 320-item catalog (target ≥60%), limiting merchandising reach (`analysis_v2/quality_metrics.json:137-144`).
- Quality improvements must remain reproducible: scenario battery S1–S10 must stay green and determinism validated (`analysis_v2/scenarios.csv`, `analysis_v2/evidence/determinism_check.json`).

---

## EPIC-01 — New User Personalization Recovery
Restore ≥10% lift for the `new_users` segment across NDCG@10 and MRR@10 while preserving existing gains in other cohorts.

- [x] TKT-01A — Diagnose onboarding regressions  
  Collect feature usage, starter profile signals, and exploration parameters for `new_users`. Compare pre/post rollout traces using samples from `analysis_v2/evidence/recommendation_samples_after_seed.json` to identify missing affinity inputs or over-weighted diversity. Deliver a written root-cause doc with hypotheses to pursue and instrument gaps (e.g., lacking recency weighting, sparse profile merge).

- [ ] TKT-01B — Implement starter profile boosts  
  Ship a blended strategy (e.g., trait-driven signals + tempered popularity) for users with ≤3 post-split events. Include overrides for `profile_boost`, `profile_top_n`, or new fallback config, and capture before/after metrics with `analysis/scripts/run_quality_eval.py --limit-users 40 --sleep-ms 120`. Success criterion: `new_users` achieve ≥+10% lift on NDCG@10 and MRR@10 relative to baseline in `analysis_v2/quality_metrics.json`.

- [ ] TKT-01C — Validate scenario resilience for new users  
  Extend `analysis/scripts/run_scenarios.py` (S7) with assertions on minimum relevance (non-zero MRR) and diversity across ≥4 categories. Add automated regression tests that fail if `new_users` fall below the target lifts or diversity thresholds.

---

## EPIC-02 — Catalog Coverage & Exploration Uplift
Increase unique catalog exposure beyond 60% without sacrificing long-tail share or core relevance.

- [ ] TKT-02A — Analyze fan-out and pruning stages  
  Instrument candidate generation and pruning (trace `extras.candidate_sources`) for a representative user set to understand where coverage drops. Share a profiling report referencing `analysis_v2/evidence/scenario_s8_new_item.json` and deterministic replay outcomes.

- [ ] TKT-02B — Tune exploration knobs  
  Experiment with `popularity_fanout`, `mmr_lambda`, and bandit parameters to widen candidate breadth. Document experiments and resulting coverage measurements using `analysis/scripts/run_quality_eval.py`. Acceptance: `system_catalog_coverage` ≥ 0.60 with `system_long_tail_unique` ≥ baseline value (101) and no segment lift regression >5%.

- [ ] TKT-02C — Roll out coverage guardrails  
  Implement monitoring (Grafana/SLO or alert) that triggers if catalog coverage dips below 60% or long-tail share below 20% in production telemetry. Provide updated runbook entries in `docs/rules-runbook.md` covering remediation steps.

---

## EPIC-03 — Evaluation Automation & Regression Safety
Ensure the evaluation suite becomes part of CI/CD to prevent future policy regressions.

- [ ] TKT-03A — CI integration for scenario suite  
  Wrap `analysis/scripts/run_scenarios.py` in a CI job (GitHub Actions or internal pipeline) with deterministic seeds and sanitized secrets. Fail the build on any scenario regression, persist artifacts under `analysis_v2/evidence/`.

- [ ] TKT-03B — Automated quality metric checks  
  Add a pipeline step running `analysis/scripts/run_quality_eval.py` against staging data. Define acceptance thresholds matching the evaluation rubric (overall lifts ≥10%, coverage ≥60%, long-tail ≥20%), failing builds on violations. Publish results to an internal dashboard.

- [ ] TKT-03C — Determinism regression test  
  Codify the determinism probe (`analysis_v2/evidence/determinism_check.json`) into an integration test that compares repeated recommendation calls and asserts ≤1% variance in rank order for deterministic configs. Log anomalies with correlation IDs for quicker debugging.

---

## EPIC-04 — Documentation & Communication
Align stakeholders on remediation plans and keep evaluation knowledge current.

- [ ] TKT-04A — Publish remediation summary  
  Draft a concise memo linking `analysis_v2/report.md`, epics, and success metrics. Share with product, modeling, and ops teams, highlighting conditional pass status and timelines to reach PASS.

- [ ] TKT-04B — Update runbooks and onboarding docs  
  Incorporate new user onboarding steps, coverage targets, and monitoring instructions into `docs/rules-runbook.md` and relevant onboarding material. Ensure instructions point to `analysis_v2/` artifacts for reproducibility.

- [ ] TKT-04C — Schedule follow-up evaluation  
  Plan a re-run of the full evaluation suite post-fixes, including dataset reseed if needed. Record target date, required resources, and pass/fail criteria so we can document improvements in the next `analysis_v2/report.md` revision.
