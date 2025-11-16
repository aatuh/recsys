# Recsys Overview (Personas & Lifecycle)

This guide explains Recsys from three vantage points—business stakeholders,
integration engineers, and developers. Each section links to the deeper docs
so you can jump directly into the material relevant to you.

> **Who should read this?** Everyone. Start here if you want to understand how the pieces fit together before diving into specific docs.

## Business / Product Stakeholders

**Goal:** Understand what Recsys delivers, which guardrails exist, and how to
audit results during campaigns.

1. **Seed & configure** – Confirm catalog/users/events are loaded for your org
   (`analysis/scripts/seed_dataset.py` or `/v1/items:upsert`, etc.). Check
   `analysis/evidence/seed_segments.json` for segment counts.
2. **Review guardrails** – Use the scenario suite (`make scenario-suite`) and
   quality eval (`analysis/scripts/run_quality_eval.py`) to ensure:
   - Segment lifts (new_users, trend_seekers, etc.) ≥ targets
   - Coverage guardrails (catalog coverage ≥0.60, long-tail share ≥0.20)
   - Policy telemetry (boost/pin exposure, leakage warnings) is healthy
3. **Override campaigns** – Merchants manage rules via `/v1/admin/rules` and
   `/v1/admin/manual_overrides`. The traces and Prometheus metrics show whether
   boosts/pins took effect.
4. **Audit & reporting** – Pull decision traces via `/v1/audit/decisions` to
   inspect what the engine produced for VIP accounts or campaigns.

> Start with: `docs/rules-runbook.md`, `docs/api_endpoints.md` (rules/overrides
sections), and the guardrail sections in `README.md`.

## Integration Engineers

**Goal:** Wire your data pipelines into Recsys and keep namespaces healthy.

1. **Ingest data** – Use `/v1/items:upsert`, `/v1/users:upsert`, `/v1/events:batch`.
   - Reference: `docs/api_endpoints.md` (Ingestion & Data Management)
   - Schema details: `docs/database_schema.md`
2. **Configure algorithms** – Edit env profiles (`api/env/*.env`, `config/profiles.yml`)
   and run `analysis/scripts/configure_env.py --namespace <ns>` to apply changes.
   - Reference: `docs/env_vars.md`
3. **Validate** – Run `analysis/scripts/run_simulation.py` or the CI workflows to
   ensure guardrails stay green before pushing changes.
4. **Maintain namespaces** – Use `analysis/scripts/reset_namespace.py` and the
   SQL snippets in `docs/database_schema.md` when you need to clean data or re-seed.

## Developers / Ops

**Goal:** Extend Recsys, debug production traffic, and monitor telemetry.

1. **API consumption** – Build clients using `docs/api_endpoints.md` (ranking,
   bandit, explainability) and the generated Swagger clients (`web/src/lib/api-client`).
2. **Telemetry & audits** – Monitor Prometheus counters (policy coverage, guardrails)
   and query decision traces (`rec_decisions`) for root-cause analysis.
3. **Simulations & CI** – Use `analysis/scripts/run_simulation.py` to reproduce
   customer scenarios locally or in CI. Artifacts live under `analysis/reports/...`.
4. **Deployment hygiene** – Each change to env profiles or rules should include:
   - Guardrail evidence (scenario + quality runs)
   - README/config matrix updates
   - Links to `docs/env_vars.md` / `docs/database_schema.md` if new knobs/tables appear

## Quick Links

| Need                  | Docs / Commands                                                     |
|-----------------------|---------------------------------------------------------------------|
| API behavior          | `docs/api_endpoints.md`, `/docs` Swagger UI                         |
| Env/algorithm knobs   | `docs/env_vars.md`, README configuration section                    |
| Database schema       | `docs/database_schema.md`                                           |
| Guardrails & runbooks | README (Onboarding & Coverage checklist), `docs/rules-runbook.md`   |
| Seeding/Simulation    | `docs/bespoke_simulations.md`, `analysis/scripts/run_simulation.py` |

Keep this overview handy when onboarding new teammates or explaining Recsys to
external partners—it shows where each persona should dive deeper.

## Lifecycle Checklist

1. **Seed data** – Load items/users/events for the target namespace via the ingestion APIs or fixtures (`analysis/scripts/seed_dataset.py --fixture-path ...`). Confirm with `/v1/items`, `/v1/users`, `/v1/events` or `analysis/evidence/seed_segments.json`.
2. **Configure env/profile** – Edit `api/env/*.env` or create a namespace-specific profile in `config/profiles.yml`, then apply with `analysis/scripts/configure_env.py --namespace <ns>`. Reference `docs/env_vars.md` for guidance.
3. **Run simulations & guardrails** – Execute `analysis/scripts/run_simulation.py --customer <name>` (or the CI workflows) to reset, seed, and run quality/scenario guardrails. Inspect `analysis/reports/<customer>/<timestamp>/` for evidence.
4. **Deploy / Update rules** – Use `/v1/admin/rules` or `/v1/admin/manual_overrides` for merchandising campaigns; check `docs/rules-runbook.md` for dry-run/testing tips.
5. **Monitor & audit** – Watch Prometheus metrics (coverage, policy telemetry), and pull decision traces via `/v1/audit/decisions` or the `rec_decisions` table to troubleshoot issues. Repeat the cycle whenever catalog, env knobs, or overrides change.
6. **Tune segments (example)**  
   ```bash
   # Reset + seed
   python analysis/scripts/reset_namespace.py --base-url https://api.pepe.local --org-id $ORG --namespace demo --force
   python analysis/scripts/seed_dataset.py --base-url https://api.pepe.local --org-id $ORG --namespace demo --users 600 --events 40000
   # Fetch/apply profile
   python analysis/scripts/env_profile_manager.py --namespace demo fetch --base-url https://api.pepe.local --org-id $ORG --profile sweep
   # Run segment sweep
   python analysis/scripts/tuning_harness.py --base-url https://api.pepe.local --org-id $ORG \
     --namespace demo --profile-name sweep --segment power_users \
     --samples 3 --alphas 0.32,0.38 --betas 0.44,0.50 --gammas 0.18,0.24 \
     --mmrs 0.18,0.26 --fanouts 450,650 --quality-limit-users 150
   # Check metrics under analysis/results/tuning_runs/demo_<timestamp>/summary.json
   # Enforce guardrails
   python analysis/scripts/check_guardrails.py --namespace tune_seg_ --min-ndcg 0.1 --min-mrr 0.1
   ```
6. **Tune segments** – Run `analysis/scripts/tuning_harness.py --segment <name>` to optimize per-cohort knobs via segment profiles, then verify with `analysis/scripts/check_guardrails.py` (or CI) that the stored sweeps continue to meet the ≥+10 % lift targets.
