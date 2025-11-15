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
- [ ] **TKT-603: AI-assisted optimizer**
  - Layer a Bayesian/LLM-guided optimizer on top of the tuning harness so it proposes the next parameter set based on prior runs, aiming to meet segment guardrails with minimal evaluations. Surface suggestions + confidence intervals in `analysis/findings.md` and expose a “one-click” tuning recipe for future evaluators.
- [ ] **TKT-604: Harness sweep for segment blend overrides**
  - Use `analysis/scripts/tuning_harness.py` to sweep `BLEND_ALPHA/BETA/GAMMA`, `MMR_LAMBDA`, and `POPULARITY_FANOUT` ranges that previously required manual `.env` edits for Trend Seekers / Weekend Adventurers / Power Users. Store runs under `analysis/results/tuning_runs/segment_blends_*` and document the winning settings in `analysis_v5/findings.md`.
- [ ] **TKT-605: Harness pass for coverage vs. relevance trade-off**
  - Configure the harness to explore higher `POPULARITY_FANOUT` + lower `MMR_LAMBDA` combinations (the manual cycles we ran to satisfy coverage guardrails). Compare catalog coverage / long-tail metrics across runs and capture the recommended env set in `analysis/results/tuning_runs/coverage_*`.
- [ ] **TKT-606: Extend harness for profile starter parameters**
  - Add optional overrides for `PROFILE_BOOST`, `PROFILE_MIN_EVENTS_FOR_BOOST`, and `PROFILE_STARTER_BLEND_WEIGHT`, then run the harness to replace the ad-hoc cold-start tuning loops. Store evidence under `analysis/results/tuning_runs/profile_*` once the extension is in place.
