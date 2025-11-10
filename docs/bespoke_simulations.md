# Bespoke Simulation Guide

Use this guide when you need to replay a customer’s dataset through Recsys, tune
algorithm knobs safely, and collect evidence for stakeholders.

## 1. Prepare environment profiles

1. Copy the closest profile (`api/env/dev.env`, `test.env`, etc.) into a new
   file (e.g., `api/env/customer_a.env`). Update the algorithm env vars to match
   the customer’s defaults (fan-out, MMR presets, starter profile weights, etc.).
2. Use `analysis/scripts/configure_env.py --profile customer_a --dry-run` to
   confirm the values and add the profile to version control so teammates can
   reuse it.
3. Optional: add the namespace/profile pairing to `guardrails.yml` so the CI
   workflows and `run_simulation.py` enforce the correct thresholds automatically.

## 2. Build a fixture

- Start from one of the templates in `analysis/fixtures/templates/` (marketplace,
  media, retail) or from `analysis/fixtures/sample_customer.json`.
- Populate the `items`, `users`, and `events` lists with real catalog IDs (or
  anonymised placeholders), ensuring the segments align with your evaluation
  cohorts (`new_users`, `power_users`, etc.).
- Validate the fixture locally:

```bash
python analysis/scripts/seed_dataset.py \
  --base-url http://localhost:8000 \
  --namespace customer_a \
  --org-id "$RECSYS_ORG_ID" \
  --fixture-path analysis/fixtures/customers/customer_a.json
```

Review `analysis/evidence/seed_segments.json` to confirm the segment counts,
then commit the fixture under `analysis/fixtures/customers/`.

## 3. Run the simulation

For a single customer:

```bash
python analysis/scripts/run_simulation.py \
  --customer customer-a \
  --base-url http://localhost:8000 \
  --namespace customer_a \
  --org-id "$RECSYS_ORG_ID" \
  --env-profile customer_a \
  --fixture-path analysis/fixtures/customers/customer_a.json
```

Key points:

- `run_simulation.py` loads the env profile, resets the namespace, seeds the
  fixture, runs `run_quality_eval.py` and `run_scenarios.py`, and stores the
  results under `analysis/reports/customer-a/<timestamp>/`.
- Guardrails (segment lifts, S7 thresholds) come from `guardrails.yml`; override
  via `--guardrails-file` if you need a temporary config.
- Each report folder contains `simulation_metadata.json`, the bundled artifacts
  (quality metrics, scenario summary, seed manifests), and a Markdown summary
  you can share with reviewers.

For multiple customers, create a manifest (see
`analysis/fixtures/batch_simulations.yaml`) and run:

```bash
python analysis/scripts/run_simulation.py \
  --batch-file analysis/fixtures/batch_simulations.yaml \
  --batch-name pilot-rollout
```

This produces per-customer reports plus a batch summary under
`analysis/reports/batches/`.

## 4. Interpret evidence

- **Quality metrics** (`analysis/reports/.../artifacts/quality_metrics.json`):
  check segment lifts, catalog coverage, and long-tail share against guardrails.
  Failures appear in the simulation log/CI output.
- **Scenario summary** (`.../scenario_summary.json`): confirm S1–S10 pass,
  especially S7 (cold-start), S8/S9 (boost/trade-off), and any customer-specific
  guardrails.
- **Seed manifests** (`.../seed_manifest.json`, `seed_segments.json`): verify
  the seeded catalog/users/events match expectations if coverage guardrails fail.

Share the entire report folder with stakeholders or upload to cloud storage; the
embedded Markdown summary includes links to each artifact and a snapshot of key
metrics.
