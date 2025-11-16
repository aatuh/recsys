# Rules & Overrides Runbook

This runbook explains how to verify, monitor, and troubleshoot the rule/override
pipeline in production.

> ⚠️ **Advanced topic**
>
> Read this after you have at least one surface integrated and want to layer in manual control safely.
>
> **Who should read this?** Business stakeholders and engineers responsible for merchandising overrides and guardrail monitoring (terms are defined in `docs/concepts_and_metrics.md`). Sections that mention `make`/`analysis/scripts` assume you run the RecSys stack; hosted API-only consumers can stick to `docs/quickstart_http.md`.
> Need exact commands or inputs? The shared script catalog (`docs/analysis_scripts_reference.md`) lists every helper under `analysis/scripts/`.
>
> **Where this fits:** Guardrails & safety.

## 1. Rule precedence refresher

When a recommendation request arrives, the engine evaluates rules in priority
order (descending). For each matching item:

1. **BLOCK** removes the item from the candidate list (pins/boosts no longer
   apply).
2. **PIN** reserves a slot (subject to `RULES_MAX_PIN_SLOTS`) and places the item
   ahead of the ranked list.
3. **BOOST** adjusts the score (multiplicative with additive fallback).

Conflicting rules are resolved by priority; if priorities tie, creation order is
used (earlier rule wins).

The policy summary emitted for every request captures the resulting counts:
`rule_block_count`, `rule_pin_count`, `rule_boost_count`,
`rule_boost_injected`, `rule_boost_exposure`, and `rule_pin_exposure`.

## 2. Day-to-day telemetry

Two sources expose runtime health:

1. **Structured logs**
   - `policy_rule_actions` – boost/pin/block counts plus the rule IDs that fired.
   - `policy_rule_exposure` – how many boosted or pinned items actually made the
     final list.
   - `policy_rule_zero_effect` – warns when a boost or pin rule matched but did
     not influence the response.
   - `manual_override_activity` – lists every manual override (override ID + rule ID)
     that matched the request, including which items were blocked, boosted, pinned, or
     ultimately served.
   - Every recommendation trace now exposes a `manual_overrides` array under
     `trace.extras`, mirroring the log payload so investigators can see override impact
     directly in stored responses.

2. **Prometheus counters** (exposed via `/metrics`)
   - `policy_rule_actions_total{namespace, surface, action}`
   - `policy_rule_exposure_total{namespace, surface, action}`
   - `policy_rule_zero_effect_total{namespace, surface, action}`
   - `policy_rule_blocked_items_total{namespace, surface, rule_id}` – lets you alert when a specific block rule starts removing too many (or too few) items.
   - `policy_override_matches_total{namespace, surface, override_id, action}` and
     `policy_override_exposure_total{namespace, surface, override_id, action}` track
     manual override matches/exposures per override ID so you can alert when a
     specific override stops firing.
   - `policy_constraint_leak_total{namespace, surface, reason}`, `policy_include_filter_*`,
     `policy_explicit_exclude_hits_total`, etc., for constraint hygiene. The `reason`
     label distinguishes between include-tag misses, price-band violations, stale items,
     and other constraint failures so dashboards can point directly to the faulty guardrail.

## 2a. Manual overrides in the pipeline

Creating a manual override (`/v1/admin/manual_overrides`) transparently writes a
high-priority rule under the hood. That rule flows through the same evaluator as
block/pin/boost rules, but the engine now remembers which override fired:

- Realtime traces/logs include the `override_id`, `rule_id`, and the affected
  items (`blocked`, `boosted`, `pinned`, `served`).
- `/metrics` exposes both match counts and exposure counts per override ID so
  you can alert on individual campaigns.
- `analysis/scripts/test_rules.py` seeds sample data, issues a boost/pin/block
  override, and asserts that the telemetry increments—run it whenever you need a
  quick end-to-end override sanity check.

Dashboards and alerts should watch the zero-effect counter and the ratio of
`rule_*_exposure` to `rule_*_count` so merchandising regressions are detected
quickly.

## 3. Coverage telemetry (ops view)

Track breadth via Prometheus rather than duplicating guardrail definitions here:

- `policy_item_served_total{namespace, surface, item_id}` – shows unique exposure counts.
- `policy_coverage_bucket_total{namespace, surface, bucket}` – `bucket=all` vs `bucket=long_tail` for exposure mix.
- `policy_catalog_items_total{namespace}` – denominator for coverage ratios.

Sample PromQL:

```promql
# Unique items shown in the past hour
count(count_over_time(policy_item_served_total{namespace="default"}[1h]) > 0)

# Catalog coverage ratio (target ≥ 0.60)
count(count_over_time(policy_item_served_total{namespace="default"}[1h]) > 0)
  / max(policy_catalog_items_total{namespace="default"})

# Long-tail share (target ≥ 0.20)
sum(rate(policy_coverage_bucket_total{namespace="default", bucket="long_tail"}[1h])) /
sum(rate(policy_coverage_bucket_total{namespace="default", bucket="all"}[1h]))
```

If ratios dip below guardrail targets, coordinate with the tuning team (`docs/tuning_playbook.md`) and consult `docs/simulations_and_guardrails.md` for the formal thresholds enforced in CI. Ops should focus on alerting + escalation, not editing guardrail YAML.

## 4. New-user onboarding playbook

1. **Seed starter data** via `analysis/scripts/seed_dataset.py` using the same org/namespace as production; this mirrors the catalog/users captured in `analysis/evidence/seed_manifest.json`.
   _Local stack note: commands such as `make scenario-suite` assume repository access with Docker/Make._
2. **Run the starter-profile scenario** (`make scenario-suite …`) and review `analysis/evidence/scenario_s7_cold_start.json` to confirm ≥4 categories and personalization reasons for the first item.
3. **Compare segment lifts** with `analysis/quality_metrics.json`; new_users should stay ≥+10% on NDCG@10/MRR@10 after any rollout (use `analysis/scripts/run_quality_eval.py` to regenerate metrics).
4. **Audit determinism** using `.github/workflows/determinism.yml` (or run `analysis/scripts/check_determinism.py --baseline analysis/evidence/determinism_check.json`) before exposing a new surface.

> For the YAML guardrails, simulation steps, and CI wiring, see `docs/simulations_and_guardrails.md`.

## 5. Operational checklist

- **Before a campaign**
  - Confirm rule creation through the admin list endpoints (`/v1/admin/rules`,
    `/v1/admin/manual_overrides`).
  - Hit `/v1/admin/recommendation/presets` if you need the curated `mmr_lambda`
    presets for a surface.
  - Run `make scenario-suite` (or `python analysis/scripts/run_scenarios.py`)
    pointing at the target environment to revalidate the full scenario suite.

- **When something looks wrong**
  1. Check `policy_rule_zero_effect` warnings for the surface/namespace in
     question. Zero exposure means the rule matched but had no effect.
  2. Inspect the corresponding Prometheus counters to gauge impact across
     traffic.
  3. Use `/v1/audit/decisions` (with `include_reasons=true` in the original
     request) to inspect the stored trace – the policy summary and rule effects
     are persisted.
  4. Re-run the scenario suite (especially the segmentation, pin/boost, and exposure-focused flows) against the
     environment to confirm fixes.

- **If zero-effect persists**
  - Verify the rule target (tags/IDs) still exists in the catalog and that the
    item is available.
  - Check for competing higher-priority rules (e.g., a block rule with the same
    target).
- Ensure boosts are sensible; remember manual boosts are relative to the
    item’s base score.

## 6. Useful commands

```bash
# Run the scenario suite against the local stack
make scenario-suite \
  SCENARIO_BASE_URL=http://localhost:8000 \
  SCENARIO_ORG_ID=00000000-0000-0000-0000-000000000001

# Tail rule-specific logs (example using jq on structured JSON logs)
journalctl -u recsys-api.service | jq 'select(.msg=="policy_rule_actions")'

# Curl the presets endpoint for UI tooling
curl -s https://api.example.com/v1/admin/recommendation/presets
```

Keep this runbook close by when launching campaigns or triaging override bugs;
the combination of telemetry and automated scenarios should keep you ahead of
regressions.
