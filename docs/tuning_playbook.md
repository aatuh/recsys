# Tuning Playbook

Use this guide when you need to change blend weights, personalization knobs, caps, or other ranking parameters for a namespace or surface. It packages the reset → seed → tune → validate → ship workflow plus troubleshooting tips.

> **Who should read this?** Developers/Ops responsible for experiment design, coverage fixes, or customer onboarding. Pair it with `docs/simulations_and_guardrails.md` for safety checks and `docs/env_reference.md` for the knob catalog. Hosted API consumers who never run the stack can skip this doc and stick to `docs/quickstart_http.md`.

> **Need a script cheat sheet?** See `docs/analysis_scripts_reference.md` for a catalog of every tool under `analysis/scripts/` plus inputs/outputs.

### TL;DR

- **Purpose:** Walk-through of the reset → seed → tune → validate workflow using the analysis scripts and tuning harness.
- **Use this when:** You need to onboard a new tenant, adjust blend/MMR/personalization knobs, or reproduce guardrail failures locally.
- **Outcome:** A curated env profile plus evidence bundles (tuning summaries, guardrail checks) that you can commit or share with stakeholders.
- **Not for:** Simple HTTP integrations or read-only reviews—stick to `GETTING_STARTED.md` / `docs/quickstart_http.md` if you aren’t running the stack yourself.

---

## 1. When to run the tuning harness

- Onboard a new tenant or namespace.
- Investigate regressions flagged by guardrails (coverage, NDCG, starter profile MRR—see `docs/concepts_and_metrics.md` for definitions).
- Try a new signal mix (e.g., heavier embeddings, stricter diversity).
- Prior to exposing new rule bundles or surfaces to production traffic.

Tuning is iterative—expect to run multiple sweeps for different cohorts (power users, long-tail shoppers, cold-start).

---

## 2. Workflow (Reset → Ship)

Set common vars first:

> Local-only note: the commands that follow assume you are running the RecSys stack from this repo (Docker/Make + Python). Hosted API consumers can ignore them.

```bash
export BASE_URL=http://localhost:8000
export ORG_ID=00000000-0000-0000-0000-000000000001
export NS=demo
```

1. **Reset namespace (clean slate)**

```bash
python analysis/scripts/reset_namespace.py \
  --base-url $BASE_URL \
  --org-id $ORG_ID \
  --namespace $NS \
  --force
```

2. **Seed catalog/users/events**

```bash
python analysis/scripts/seed_dataset.py \
  --base-url $BASE_URL \
  --org-id $ORG_ID \
  --namespace $NS \
  --users 600 \
  --events 40000
```

3. **Fetch & edit an env profile**

```bash
python analysis/scripts/env_profile_manager.py \
  --namespace $NS \
  --base-url $BASE_URL \
  --org-id $ORG_ID \
  --profile sweep_baseline \
  fetch
```

- Edit `analysis/env_profiles/$NS/sweep_baseline.json` (blend weights, personalization knobs, per-segment bundles).
- Apply changes via `... --apply` once the JSON looks good.

4. **Run the tuning harness**

```bash
python analysis/scripts/tuning_harness.py \
  --base-url $BASE_URL \
  --org-id $ORG_ID \
  --namespace $NS \
  --profile-name sweep_baseline \
  --segment power_users \
  --samples 3 \
  --seed 2025 \
  --alphas 0.32,0.38 \
  --betas 0.44,0.50 \
  --gammas 0.18,0.24 \
  --mmrs 0.18,0.26 \
  --fanouts 450,650 \
  --profile-boosts 0.6,0.75 \
  --starter-blend-weights 0.7,0.9 \
  --reset-namespace \
  --sleep-ms 400 \
  --quality-limit-users 150 \
  --quality-request-timeout 180
```

MMR-related flags above (`--mmrs`, `MMR_LAMBDA`) refer to Maximal Marginal Relevance (MMR; see `docs/concepts_and_metrics.md` for definitions).

Outputs live in `analysis/results/tuning_runs/${NS}_<timestamp>/` with per-run JSON and summaries.

5. **Optional – AI-assisted suggestions**

```bash
python analysis/scripts/ai_optimizer.py \
  --namespace tune_seg_power_users \
  --objective segment_ndcg_lift \
  --suggestions 5 \
  --alpha-range 0.3 0.5 \
  --beta-range 0.4 0.6 \
  --gamma-range 0.1 0.3 \
  --mmr-range 0.15 0.35 \
  --fanout-range 400 800 \
  --output analysis/results/next_suggestions.json
```

Feed promising suggestions back into `tuning_harness.py` sweeps.

6. **Guardrail check**

```bash
python analysis/scripts/check_guardrails.py \
  --namespace tune_seg_power_users \
  --min-ndcg 0.1 \
  --min-mrr 0.1
```

These thresholds map to `guardrails.yml`. CI also runs this command; keep it fast locally to debug issues before pushing.

7. **Snapshot winning profile & evidence**

- Apply via `env_profile_manager.py --apply`.
- Commit `analysis/env_profiles/...` and `analysis/results/tuning_runs/.../summary.json`.
- Include top metrics (NDCG lift, coverage, long-tail share) in your PR or ticket.

---

## 3. Reading the results

- **`summary.json`**: aggregated metrics per sweep—look for gains in target segments while keeping coverage and long-tail share above guardrail floors.
- **`run_###.json`**: per-run telemetry including blend weights, overrides applied, guardrail verdicts, request/response samples.
- **Evidence bundle**: zipped traces + charts (if enabled) for sharing with PMs.

Key checks:

- Starter profile (scenario S7) meets minimum categories and MRR.
- `catalog_coverage` and `long_tail_share` stay within expected ranges.
- `policy_rule_blocked_items_total` counters behave (no unexpected spikes).

---

## 4. Troubleshooting

- **Segment guardrail failure:** Narrow the search space (e.g., `--fanouts 450,550 --mmrs 0.2,0.3`) and re-run for that cohort only.
- **Coverage shortfall:** Increase `POPULARITY_FANOUT`, lower `MMR_LAMBDA`, or raise long-tail weights in the env profile.
- **Seeding/connection errors:** Ensure Docker is running, rerun `seed_dataset.py`, or hit the proxy host with `--insecure`.
- **Slow harness runs:** Reduce `--samples`, trim override ranges, or temporarily disable AI optimizer suggestions.

---

## 5. Relationship to guardrails & simulations

- Guardrail failures should lead to **tuning changes**, not suppressing the guardrail itself, unless business requirements change.
- Use `docs/simulations_and_guardrails.md` to craft bespoke scenarios (e.g., customer-specific catalogs) that the tuning harness can borrow.
- Update `guardrails.yml` only after stakeholders agree on new thresholds and you have evidence from the harness + simulations.

---

## 6. Related documents

- `docs/env_reference.md` – knob definitions and override mapping.
- `docs/simulations_and_guardrails.md` – scenario fixtures, guardrail YAML, CI hooks.
- `docs/rules_runbook.md` – operational checklist once the tuned profile is live.
- `docs/concepts_and_metrics.md` – metric definitions used throughout the playbook.
