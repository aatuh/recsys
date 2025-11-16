# Analysis Scripts Reference

Use this catalog to understand what lives under `analysis/scripts/`, when to reach for each tool, and what inputs/outputs to expect. Scripts intentionally keep flags consistent (`--base-url`, `--org-id`, `--namespace`) so you can swap them into the workflows described in [`GETTING_STARTED.md`](../GETTING_STARTED.md), [`docs/tuning_playbook.md`](tuning_playbook.md), and [`docs/simulations_and_guardrails.md`](simulations_and_guardrails.md).

> ⚠️ **Advanced topic**
>
> Read this after you have a basic integration running (see [`docs/quickstart_http.md`](quickstart_http.md)) and are comfortable running the stack locally.
>
> **Where this fits:** Observability & analysis.

## Seeding & Reset

- **`analysis/scripts/seed_dataset.py`** — Generate a realistic catalog/users/events set via the public API.
  - *Who / When:* Senior developers before running local evaluations or demos.
  - *Inputs & Outputs:* Requires a running API and credentials; writes evidence under `analysis/evidence/` plus seeded entities in the target namespace.
  - *Example:* `python analysis/scripts/seed_dataset.py --base-url http://localhost:8000 --org-id 00000000-0000-0000-0000-000000000001 --namespace demo --users 320 --events 20000`
- **`analysis/scripts/reset_namespace.py`** — Delete all items/users/events in a namespace.
  - *Who / When:* Dev/Ops when you need a clean slate for simulations or tuning.
  - *Inputs & Outputs:* Needs `--base-url`, `--org-id`, `--namespace`; no outputs except console log.
  - *Example:* `python analysis/scripts/reset_namespace.py --base-url http://localhost:8000 --org-id 00000000-0000-0000-0000-000000000001 --namespace demo --force`

## Evaluation & Scenarios

- **`analysis/scripts/run_quality_eval.py`** — Compare system recommendations vs. a baseline and compute lift/coverage/diversity metrics.
  - *Who / When:* Developers validating changes or running CI-quality checks.
  - *Inputs & Outputs:* Requires seeded namespace and running API; writes `analysis/quality_metrics.json` plus `{namespace}_warm_quality.json` under `analysis/results/`.
  - *Example:* `python analysis/scripts/run_quality_eval.py --base-url http://localhost:8000 --org-id 00000000-0000-0000-0000-000000000001 --namespace demo --min-segment-lift-ndcg 0.1`
- **`analysis/scripts/run_scenarios.py`** — Execute the full policy/regression scenario suite and capture evidence.
  - *Who / When:* PMs/developers demonstrating guardrail compliance.
  - *Inputs & Outputs:* Needs seeded namespace; writes `analysis/scenarios.csv` and JSON payloads under `analysis/evidence/`.
  - *Example:* `python analysis/scripts/run_scenarios.py --base-url http://localhost:8000 --org-id 00000000-0000-0000-0000-000000000001 --namespace demo --s7-min-avg-mrr 0.2 --s7-min-avg-categories 4`
- **`analysis/scripts/run_simulation.py`** — Full automation: configure env, reset, seed (optionally via fixtures), run quality eval & scenarios, and bundle reports.
  - *Who / When:* Solutions engineers running bespoke simulations for customers.
  - *Inputs & Outputs:* Consumes manifests/fixtures from `analysis/fixtures/`; produces bundles under `analysis/reports/` plus all standard evaluation artifacts.
  - *Example:* `python analysis/scripts/run_simulation.py --customer acme --batch-file analysis/fixtures/batch_simulations.yaml`
- **`analysis/scripts/exposure_dashboard.py`** — Build exposure metrics from a recommendation dump (power users vs long-tail).
  - *Who / When:* Analysts investigating coverage/exposure issues.
  - *Inputs & Outputs:* Needs a `recommendation_dump.json`; writes `analysis/results/exposure_dashboard.json`.
  - *Example:* `python analysis/scripts/exposure_dashboard.py --dump analysis/results/recommendation_dump.json`

## Environment & Profiles

- **`analysis/scripts/configure_env.py`** — Rewrite `api/.env` (or another env file) from profiles + overrides and capture history.
  - *Who / When:* Dev/Ops changing stack-level env vars.
  - *Inputs & Outputs:* Reads `api/env/*.env` and `config/profiles.yml`; writes back to `api/.env` plus history under `analysis/env_history/`.
  - *Example:* `python analysis/scripts/configure_env.py --profile dev --set API_PORT=8081 --note "staging tweak"`
- **`analysis/scripts/env_profile_manager.py`** — Fetch/apply recommendation configs via `/v1/admin/recommendation/config`.
  - *Who / When:* Tuning engineers managing namespace profiles.
  - *Inputs & Outputs:* Talks to the live API; stores JSON under `analysis/env_profiles/<namespace>/`.
  - *Example:* `python analysis/scripts/env_profile_manager.py --namespace demo --base-url https://api.customer.com --org-id 00000000-0000-0000-0000-000000000001 fetch --profile sweep_baseline`
- **`analysis/scripts/check_env_profiles.py`** — Ensure every `api/env/*.env` has the same keys as the base file.
  - *Who / When:* Anyone editing env files to catch drift.
  - *Inputs & Outputs:* Reads env files only; prints mismatches and exits non-zero with `--strict`.
  - *Example:* `python analysis/scripts/check_env_profiles.py --base api/.env --strict`

## Tuning & Guardrails

- **`analysis/scripts/tuning_harness.py`** — Apply profile overrides, run seeding + quality eval per parameter grid, and summarize metrics.
  - *Who / When:* Senior developers running blend/MMR sweeps per namespace or segment.
  - *Inputs & Outputs:* Requires running API & profile storage; writes run JSONs under `analysis/results/tuning_runs/<namespace_timestamp>/`.
  - *Example:* `python analysis/scripts/tuning_harness.py --base-url https://api.customer.com --org-id 00000000-0000-0000-0000-000000000001 --namespace tune_seg_power --profile-name sweep_baseline --segment power_users --alphas 0.32,0.38 --mmrs 0.2,0.3`
- **`analysis/scripts/ai_optimizer.py`** — Suggest next tuning parameters via surrogate modeling on past runs.
  - *Who / When:* Developers after collecting enough tuning runs.
  - *Inputs & Outputs:* Reads `analysis/results/tuning_runs/**/summary.json`; prints suggestions or writes JSON.
  - *Example:* `python analysis/scripts/ai_optimizer.py --namespace tune_seg_power --objective segment_ndcg_lift --suggestions 5`
- **`analysis/scripts/profile_coverage.py`** — Report coverage gaps per profile/segment to guide tuning focus.
  - *Who / When:* Analysts validating profile health.
  - *Inputs & Outputs:* Reads recorded metrics (quality JSON); prints coverage stats.
  - *Example:* `python analysis/scripts/profile_coverage.py --results analysis/results/tune_seg_coverage_warm_quality.json`
- **`analysis/scripts/check_guardrails.py`** — Validate tuning run summaries against `guardrails.yml` thresholds.
  - *Who / When:* CI and developers before promoting new configs.
  - *Inputs & Outputs:* Scans `analysis/results/tuning_runs/**/summary.json`; exits non-zero on violations.
  - *Example:* `python analysis/scripts/check_guardrails.py --namespace tune_seg_ --min-ndcg 0.1 --min-mrr 0.1`

## Reliability, Rules, and Load

- **`analysis/scripts/check_determinism.py`** — Replay a fixed request multiple times to ensure recommendation stability.
  - *Who / When:* Dev/Ops diagnosing inconsistent rankings; also runs in CI.
  - *Inputs & Outputs:* Needs a baseline request JSON (e.g., `analysis/results/determinism_baseline.json`); writes run results to `analysis/results/determinism_run.json`.
  - *Example:* `python analysis/scripts/check_determinism.py --base-url http://localhost:8000 --org-id 00000000-0000-0000-0000-000000000001 --namespace demo --baseline analysis/results/determinism_baseline.json`
- **`analysis/scripts/recommendations_load.py`** — Python load test that drives staged RPS against `/v1/recommendations`.
  - *Who / When:* Engineers needing a stress test without k6.
  - *Inputs & Outputs:* Requires a running API; writes `analysis/results/load_test_summary.json`.
  - *Example:* `python analysis/scripts/recommendations_load.py --base-url http://localhost:8000 --org-id 00000000-0000-0000-0000-000000000001 --namespace demo --rps-targets 10,50,100`
- **`analysis/scripts/test_rules.py`** — Evaluate merchandising/rule logic with canned scenarios.
  - *Who / When:* Rule authors verifying overrides, boosts, and pins.
  - *Inputs & Outputs:* Consumes rule fixtures; prints pass/fail per scenario.
  - *Example:* `python analysis/scripts/test_rules.py --base-url http://localhost:8000 --org-id 00000000-0000-0000-0000-000000000001 --namespace demo`
- **`analysis/scripts/chaos_toggle.py`** — Flip feature flags / chaos experiments for targeted namespaces.
  - *Who / When:* Ops experimenting with failure modes.
  - *Inputs & Outputs:* Calls admin API; prints status.
  - *Example:* `python analysis/scripts/chaos_toggle.py --base-url http://localhost:8000 --org-id 00000000-0000-0000-0000-000000000001 --namespace demo --enable`
