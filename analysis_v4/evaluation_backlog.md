# Recsys Automation & Simulation Backlog

## Problem Statement
Operating Recsys across multiple customers requires more than the “golden path” evaluation suite. Each tenant has bespoke catalog/user/event distributions, service configuration drift (env vars vs. request overrides), and different guardrails. Today we can seed the reference dataset and run scenarios/quality, but we lack:
- An automation-friendly way to seed & populate arbitrary datasets.
- A workflow that can tune env-based algorithm knobs (starter profiles, fan-out, MMR presets, etc.) without manual edits and restarts.
- A simulation harness that strings seeding + tuning + scenarios/quality into repeatable experiments for each customer or rollout.

The following epics define the work required to build that tooling.

---

## Epics & Tickets

### AP-201: Bespoke Seeding & Population
Provide scripts + configuration to ingest customer-specific catalog/users/events, including traits and segment metadata, and capture evidence for audits.

- [x] AP-201A — Customer fixture schema & loader  
  *Define a YAML/JSON schema (catalog fields, segment metadata, event distributions) and extend `analysis/scripts/seed_dataset.py` (or a new module) to read fixtures per customer and ingest via /v1/items:upsert, /v1/users:upsert, /v1/events:batch. Include CLI flags like `--fixture-path` and write evidence under `analysis/customer_data/<customer>/`.*  
  - Added `--fixture-path` plus `load_fixture` support (see `analysis/scripts/seed_dataset.py`); sample fixture lives at `analysis/fixtures/sample_customer.json`, and the README now documents how to use it.
- [x] AP-201B — Segment/profile alignment  
  *Ensure bespoke seeds populate segment profiles + traits so the scenario harness (S7) and quality eval know which users belong to which cohorts. Provide validation output (e.g., counts per segment, sample trait dumps).*  
  - `analysis/scripts/seed_dataset.py` now records per-segment counts and sample traits in `analysis/evidence/seed_segments.json`, so bespoke fixtures can be validated quickly after ingestion (see updated README note).
- [x] AP-201C — Data reset utilities  
  *Add scripts/Make targets to wipe a namespace (delete items/users/events) before re-seeding, preventing cross-contamination between customer runs.*  
  - Introduced `analysis/scripts/reset_namespace.py` plus `make reset-namespace`, which call the delete endpoints in order (events → users → items) and emit `analysis/evidence/reset_<timestamp>.json` for auditing prior to bespoke seeding.

### AP-202: Automated Env Var Tuning & Restarts
Enable scripted edits to `api/.env` (or a namespace-specific env file), trigger service restarts, and track which env set was used for each experiment.

- [x] AP-202A — Env patcher CLI  
  *Create `scripts/configure_env.py` or similar that accepts a profile name or key=value overrides, rewrites `api/.env`, and logs the change (e.g., `analysis/env_history/<timestamp>.json`).*  
  - Implemented `analysis/scripts/configure_env.py` with `--profile`/`--set` support; it rewrites `api/.env`, prints the diff, and persists `analysis/env_history/<timestamp>.json` containing before/after + SHA so experiments stay traceable.  
- [x] AP-202B — Restart orchestration  
  *Add a helper (`scripts/restart_api.py` or Make target) that runs `docker compose up -d --force-recreate api` (or the relevant command for non-Compose environments) and blocks until health checks pass. Integrate with the env patcher.*  
  - Added `analysis/scripts/restart_api.py` plus `make restart-api`; the script executes the compose restart (optionally with custom files/project) and polls `/health` at the supplied base URL, so env changes from AP-202A can be applied + validated automatically.
- [x] AP-202C — Tracking & provenance  
  *Extend the harness to capture which env profile was active when quality/scenario evidence was produced (e.g., embed `env_hash` inside `analysis/scenarios.csv` / `analysis/quality_metrics.json`).*  
  - `analysis/scripts/env_utils.py` computes the SHA of `api/.env`, `run_quality_eval.py` and `run_scenarios.py` gained `--env-file` flags, and their outputs now include `env_file`/`env_hash` metadata (`analysis/quality_metrics.json`, `analysis/scenarios.csv`, `scenario_summary.json`), giving every evidence artifact clear provenance.

### AP-203: Simulation & Reporting Suite
Combine bespoke seeding, env tuning, and the existing quality/scenario harness into a configurable “simulation run” CLI that outputs a comprehensive report per customer/experiment.

- [x] AP-203A — Simulation runner CLI  
  *Implement `scripts/run_simulation.py` that orchestrates: (1) configure env (optional), (2) restart API, (3) seed customer data, (4) run quality eval with guardrail flags, (5) run scenario suite with S7 thresholds. Allow per-customer configs referencing fixtures and env profiles.*  
  - Delivered `analysis/scripts/run_simulation.py` with `--customer`, env profile/override support, reset + seed orchestration, quality/scenario toggles, and provenance output under `analysis/reports/<customer>/<timestamp>/simulation_metadata.json`. README now documents usage/example invocation.
- [x] AP-203B — Report bundling  
  *After a simulation run, bundle evidence (`analysis_v3/…` snapshots, env profile, CLI args) into `analysis/reports/<customer>/<timestamp>/` and produce a human-readable summary (Markdown or HTML) linking to the artifacts.*  
  - `run_simulation.py` now copies quality metrics, scenario summary/CSV, and seeding manifests into each report folder, emits `simulation_metadata.json`, and writes a Markdown summary (`README.md`) that lists lifts + scenario pass rates with links to the bundled artifacts.
- [x] AP-203C — Batch simulations  
  *Add support for running multiple simulations back-to-back (e.g., iterate over customers/profiles) and summarize pass/fail status across the batch. Useful for regression testing before releases.*  
  - `run_simulation.py` accepts `--batch-file` (YAML/JSON) plus `--batch-name`; each entry can override env/fixture settings, and the script now writes `analysis/reports/batches/<name>_<timestamp>.json` summarizing report paths and scenario outcomes for every run.

### AP-204: Guardrail Extensibility
Extend the current guardrails (segment lift, scenario S7 thresholds) so they can be tuned per customer or pipeline.

- [x] AP-204A — Guardrail config file  
  *Define a `guardrails.yml` where each customer/namespace declares S7 thresholds, segment lift minimums, and coverage targets. The orchestrator reads this file and passes flags to the scripts.*  
  - Added root `guardrails.yml` plus `analysis/scripts/guardrails.py`; `run_simulation.py` now loads the file (or per-run overrides) and applies the thresholds to both quality and scenario harnesses, including new catalog/long-tail guardrails enforced by `run_quality_eval.py`.
- [x] AP-204B — CI integration  
  *Update CI workflows to read `guardrails.yml` so the right thresholds apply per branch/customer (for example, a “beta” namespace might tolerate lower coverage while seeding).*  
  - Both `.github/workflows/quality-eval.yml` and `scenario-suite.yml` now resolve guardrails via `analysis/scripts/guardrails.py`, pass the resulting thresholds to `run_quality_eval.py` / `run_scenarios.py`, and install PyYAML for the loader. The Makefile exposes `S7_MIN_AVG_*` knobs so CI can honor per-customer S7 guardrails automatically.
- [x] AP-204C — Documentation update  
  *Add a guardrail reference section (README + docs/rules-runbook.md) explaining how to define thresholds, run simulations per customer, and interpret failures.*  
  - README now contains a dedicated “Guardrail configuration” section covering `guardrails.yml`, simulation overrides, and CI behavior; `docs/rules-runbook.md` documents the workflow for editing guardrails and validating them via simulation/CI. 

### AP-205: Bespoke Fixture Library & Samples
Ship example fixtures and documentation so customers (or internal teams) can clone, edit, and simulate their own data locally.

- [x] AP-205A — Fixture templates  
  *Provide templates under `analysis/fixtures/<customer>/` covering common patterns (marketplace, media, retail). Document required fields, optional overrides, and sample size guidance.*  
  - Added `analysis/fixtures/templates/{marketplace,media,retail}.json` plus `analysis/fixtures/README.md`; each template includes ready-to-seed catalog/users/events and the README explains required fields, optional props, and how to validate via `seed_dataset.py` or batch simulations.
- [x] AP-205B — Sample scripts  
  *Add `scripts/examples/` showing how to run the full simulation for a sample customer (configure env → seed fixture → run quality/scenarios → report).*  
  - Introduced `analysis/scripts/examples/run_marketplace_simulation.sh` (plus README pointers) to demonstrate the configure → seed → simulate flow with guardrails, namespace selection, and fixture templates.
- [x] AP-205C — Onboarding guide  
  *Write a customer-facing guide (docs/bespoke_simulations.md) describing how to craft fixtures, run the simulation suite, and interpret the evidence.*  
  - Authored `docs/bespoke_simulations.md` covering env profiles, fixture creation, simulation commands (single + batch), and how to read the resulting artifacts.

### AP-206: Configuration Profiles & Runtime Overrides
Make environment profiles first-class so simulations/config tooling can switch namespaces without restarts, and document which knobs are safe to tweak at runtime.

- [x] AP-206A — Env profile audit & namespace linkage  
  *Ensure every algorithm env var exists across `api/env/{dev,test,prod,ci}.env`, add a validation script that diff-checks profile files vs. `api/.env`, and design how profiles tie to namespaces (e.g., `profiles.yml` keyed by namespace). Output should make it possible to associate a new namespace with its own env profile so simulations can swap profiles instead of restarting services.*  
  - `api/env/dev|test|prod.env` now include the bandit experiment, coverage, and MMR preset knobs that were missing. Added `analysis/scripts/check_env_profiles.py` (diffs `api/.env` vs. `api/env/*.env` with a `--strict` mode) and `config/profiles.yml` to document namespace-to-profile intent.
- [x] AP-206B — Runtime override matrix  
  *Catalog all request-level overrides exposed by `recommendation_config.go` (MMR presets, starter-profile weights, bandit toggles, etc.), wire through any missing “useful” algorithm knobs, and update README/config docs with an explicit list of supported on-the-fly parameters. Anything not on the list stays env-only.*  
  - Verified the handler already threads the override fields; README now documents the `overrides` object (popularity/co-vis/MMR, profile knobs, blend weights, bandit algo) with an example payload so tooling knows exactly which params can be tuned on-the-fly.
- [x] AP-206C — Simulation profile support  
  *Extend `analysis/scripts/run_simulation.py` (and configure_env) so each run can select an env profile + namespace pairing without restarts. If profiles are bound to namespaces, the orchestrator should simply pick the correct profile before seeding; otherwise, add a light-weight mechanism to swap profiles per run. Document the new flow and highlight that this eliminates the need to restart/clean the API between simulations.*  
  - `config/profiles.yml` defines namespace→profile mappings. `configure_env.py` and `run_simulation.py` accept `--namespace` / `--profiles-file` and automatically pick the right env profile before applying overrides, so simulations can swap namespaces without restarting the API.

---

Each ticket should update the relevant automation scripts, documentation, and evidence directories to keep the workflow reproducible. This backlog can live alongside the existing `analysis_v3/` evaluation artifacts to track progress toward customer-ready automation. 
