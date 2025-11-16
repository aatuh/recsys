# Recsys Evaluation Backlog (v5)

## Top-Level Description
This backlog tracks the engineering work required to close the gaps uncovered during the v5 evaluation of **recsys**. It focuses on restoring core recommendation capabilities (similar items, rerank, personalization), enforcing policy controls, and ensuring the platform can be operated safely at target traffic tiers.

## Problem Statement
Despite strong offline lift on tiny samples, the current system cannot pass the production rubric because critical surfaces are missing (no rerank/ANN), policy controls were previously disabled, constraint enforcement is incomplete, safety/fairness guardrails are not measurable, and serving/ops tooling is minimal. We need a coordinated plan to implement the missing features, harden the runtime, and provide observability hooks so future evaluations can be automated.

---

## Epic E-101 – Rebuild Core Recommendation Surfaces
**Description:** Implement the remaining catalog-facing APIs (similar items, rerank) and populate the data they need so every test in sections A1–A3 can run.

- [x] **TKT-101: Wire up item embeddings + ANN store**
  - Items posted through `/v1/items:upsert` now get deterministic embeddings when none are supplied (see `api/internal/services/ingestion/service.go` and the new unit coverage). The evaluation seeder synthesizes identical vectors via `analysis/scripts/seed_dataset.py`, and a regression test (`api/test/handlers/similar_test.go`) confirms `/v1/items/{item_id}/similar` returns neighbors out of the box. README notes that catalog backfills remain available for legacy imports.
- [x] **TKT-102: Implement `/v1/rerank` endpoint and engine**
  - `/v1/rerank` now reuses the ranking pipeline with prefetched candidates: service validation, algorithm support for `PrefetchedCandidates`, and telemetry parity with `/v1/recommendations`. Swagger + TS clients regenerated, docs/README reference the endpoint, and an HTTP regression test (`api/test/handlers/recommend_test.go`) proves personalization can reorder the supplied list. `go test ./...` passes. Segment-specific blend overrides (`BLEND_SEGMENT_OVERRIDES`) are supported so cohorts like Trend Seekers / Weekend Adventurers can get bespoke mixes without namespace cloning.
- [x] **TKT-103: Expand offline eval coverage**
  - `analysis/scripts/run_quality_eval.py` now enforces ≥100 warm users (≥50 historical events) before evaluating, adds `/v1/rerank` and `/v1/items/{id}/similar` suites, and writes all artifacts to `analysis/results/` plus evidence snapshots. CLI gains knobs for warm thresholds, rerank queries/candidate pools, and similar-item sampling so the rubric’s A1–A3 tests can run from one command.

## Epic E-102 – Enforce Policy & Configurability Guardrails
**Description:** With `RULES_ENABLE=true` in place, ensure all hard/soft policy levers work and are testable, including price/availability constraints.

- [x] **TKT-201: Honor price/time constraints in `applyConstraintFilters`**
  - Extend `api/internal/algorithm/engine.go:817-878` so `price_between` and `created_after` filters remove invalid items before scoring. Add unit tests covering low-price surfaces and age gating.
- [x] **TKT-202: Rule-test automation & evidence**
  - Add an `analysis/scripts/test_rules.py` (or extend existing simulations) that creates sample block/boost/pin rules, verifies `rule_*` counters in traces, and stores before/after payloads (similar to `analysis/results/rules_effect_sample.json`).
- [x] **TKT-203: Manual override lifecycle telemetry**
  - Emit override IDs + rule IDs into traces and `/metrics` so merch traffic sees exposure. Update docs describing how `/v1/admin/manual_overrides` interacts with rules.

## Epic E-103 – Safety, Fairness, and Cold-Start Quality
**Description:** Reduce exposure concentration, improve zero-data personalization, and add instrumentation to measure fairness.

- [x] **TKT-301: Diversify cold-start/starter profiles**
  - Update starter profile weights + anchors (see `analysis/results/recommendation_dump.json:34-120`) to avoid identical top-10 lists. Include entropy/diversity logging per surface.
- [x] **TKT-302: Brand/category exposure dashboards**
  - Use the summary generator in `analysis/results/recommendation_dump.json` as a base to produce brand/category exposure ratios per namespace. Fail CI if max/mean ratio > 1.4× without justification.
- [x] **TKT-303: Policy violation monitors**
  - Added reason-aware constraint leak tracking plus per-rule block exposure counters in the algorithm summary, wired them into Prometheus metrics (`policy_constraint_leak_total{reason=*}`, `policy_rule_blocked_items_total{rule_id=*}`), and documented the new metrics in `docs/rules-runbook.md`.

## Epic E-104 – Serving & Operational Readiness
**Description:** Prove the service can handle target RPS with observability and rollback tooling.

- [x] **TKT-401: Version endpoint & determinism tests**
  - `/version` (GET) now returns git SHA, build time, and the default model label (documented + generated clients updated). Added `make determinism`, refreshed the GitHub workflow/baseline (`analysis/results/determinism_baseline.json`), and wired the replay harness into the backlog so segment guardrails include determinism evidence.
- [x] **TKT-402: Load & chaos harness**
  - Added `analysis/load/recommendations_k6.js` plus `make load-test` (dockerized k6) to ramp 10→100→1000 RPS and emit `analysis/results/load_test_summary.json`. `analysis/scripts/chaos_toggle.py` pauses/stops compose services on demand so we can observe cache/store failures mid-run; README documents the workflow.
- [x] **TKT-403: Config rollback tooling**
  - `/v1/admin/recommendation/config` now supports GET/POST, backed by an in-process config manager and history metadata. Added `analysis/scripts/recommendation_config.py` plus sample templates under `config/recommendation/` so teams can export/apply configs, commit them to git, and roll back without redeploying; README/docs describe the workflow.

## Epic E-105 – Documentation, DX & Reporting
**Description:** Ensure the findings and workflows are captured for future evaluators and partner teams.

- [x] **TKT-501: Update README & docs with new workflows**
  - Document rerank API, rules testing, and load-test steps in `README.md` and `docs/api_endpoints.md`, keeping persona-based navigation intact.
- [x] **TKT-502: Automate evaluation report bundling**
  - Extend `analysis/scripts/run_simulation.py` to include the new artifacts (`rules_effect_sample.json`, recommendation dumps/exposure dashboards, load-test summaries) and push them to `analysis/results`.
- [x] **TKT-503: Executive summary template**
  - Added `analysis/templates/executive_summary_template.md` plus README instructions so each bundle ships with a consistent “3 strengths / 3 blockers / 3 fast wins” page.

## Epic E-106 – Automated Configuration & Tuning
**Description:** Eliminate manual env editing cycles by introducing namespace-scoped configuration profiles, experimentation helpers, and automated search (heuristic or AI-driven) that can discover guardrail-satisfying blends on demand.

- [x] **TKT-601: Namespace-bound env profile manager**
  - Added `analysis/scripts/env_profile_manager.py`, README + docs references, and git-ignored `analysis/env_profiles/` so teams can fetch/apply/delete namespace-scoped configs via `/v1/admin/recommendation/config` without rewriting `api/.env` or recycling services.
- [x] **TKT-602: Automated tuning harness**
  - `analysis/scripts/tuning_harness.py` now applies profiles via the new manager, re-seeds, runs quality eval, and saves per-run metrics/parameters under `analysis/results/tuning_runs/`. README documents grid/random usage and how to interpret the output.
- [x] **TKT-603: AI-assisted optimizer**
  - Added `analysis/scripts/ai_optimizer.py`, which ingests previous tuning runs, fits a surrogate (Gaussian process when `scikit-learn` is available, otherwise weighted exploration), and outputs recommended parameter sets for `tuning_harness.py`. README documents the workflow.
- [x] **TKT-604: Harness sweep for segment blend overrides**
  - Segment-specific sweeps are done (`tune_seg_power_*`, `tune_seg_trend_*`, `tune_seg_test_*`), each storing summary metrics under `analysis/results/tuning_runs/segment_blends_*` and capturing blend/MMR/fanout overrides via `segment_profiles`. Findings include the recommended configs per cohort.
- [x] **TKT-605: Harness pass for coverage vs. relevance trade-off**
  - Conducted coverage sweeps (`tune_cov_default_20251116T071125Z`, `tune_cov_default_20251116T072432Z`) using higher fanout/lower MMR; final profile hits catalog coverage 0.77 and long-tail 0.48 while keeping power_users/trend_seekers lifts > +10 %. Evidence sits under `analysis/results/tuning_runs/tune_cov_default_20251116T072432Z/`.
- [x] **TKT-606: Extend harness for profile starter parameters**
  - `analysis/scripts/tuning_harness.py` now accepts `--profile-boosts`, `--profile-min-events`, and `--starter-blend-weights` so starter-profile knobs can be swept (per segment) alongside blend/MMR values. README documents the workflow.
- [x] **TKT-607: Segment profile storage & admin API**
  - `/v1/admin/recommendation/config` now includes a `segment_profiles` map (exposed via swagger + TS client). Config snapshots, env_profile_manager CLI, and codegen outputs all persist blend/MMR/fanout/starter overrides per segment so we can version and diff them without touching `.env`.
- [x] **TKT-608: Segment-focused tuning harness**
  - `analysis/scripts/tuning_harness.py` accepts `--segment` to mutate only that entry under `segment_profiles`, records segment-specific lifts in each run, and stores evidence under `analysis/results/tuning_runs/<namespace>_*`. Verified via `tune_seg_test` runs with `power_users`.
- [x] **TKT-609: Guardrail enforcement in CI**
  - `analysis/scripts/check_guardrails.py` scans tuning `summary.json` files for lifts/coverage and now runs inside `.github/workflows/quality-eval.yml` (after quality eval). CI fails automatically if overall or per-segment metrics fall below the thresholds resolved from `guardrails.yml`.
- [ ] **TKT-610: Documentation & playbook**
  - Update README + `docs/overview.md`/`docs/env_vars.md` with the new segment-profile workflow: how to fetch/apply profiles, run the per-segment harness, interpret evidence, and roll back via git-managed profiles. Include a troubleshooting section covering namespace hygiene and guardrail failures.

## Epic E-111 – Beginner-Friendly Documentation Overhaul
**Description:** Restructure the documentation so new engineers can follow the entire tuning workflow, understand guardrail responses, and troubleshoot without prior context.

- [x] **TKT-701: README tuning playbook**
  - Add a top-level “Recsys tuning playbook” section detailing reset → seed → fetch/apply profile → run harness/optimizer → check guardrails, with commands and example outputs.
- [ ] **TKT-702: Troubleshooting guide**
  - Document common failures (segment guardrail, coverage shortfall, connection issues) and prescribe next actions in README and `docs/overview.md`.
- [x] **TKT-703: Starter profile guidance**
  - Expand README + `docs/env_vars.md` with plain-English explanations of starter profile knobs (PROFILE_BOOST, PROFILE_MIN_EVENTS_FOR_BOOST, PROFILE_STARTER_BLEND_WEIGHT) and link to harness flags.
- [x] **TKT-704: docs/overview worked example**
  - Insert a step-by-step example (with commands/metrics) showing a full segment tuning run and how evidence is stored under `analysis/results/tuning_runs/`.
- [x] **TKT-705: Cross-link env & rules docs**
  - Ensure `docs/env_vars.md`, `docs/rules-runbook.md`, and `docs/bespoke_simulations.md` reference the tuning playbook so readers can jump directly to the workflow.
