# Rules & Overrides Runbook

This runbook explains how to verify, monitor, and troubleshoot the rule/override
pipeline in production.

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

2. **Prometheus counters** (exposed via `/metrics`)
   - `policy_rule_actions_total{namespace, surface, action}`
   - `policy_rule_exposure_total{namespace, surface, action}`
   - `policy_rule_zero_effect_total{namespace, surface, action}`
   - `policy_include_filter_*`, `policy_explicit_exclude_hits_total`, etc., for
     constraint hygiene.

Dashboards and alerts should watch the zero-effect counter and the ratio of
`rule_*_exposure` to `rule_*_count` so merchandising regressions are detected
quickly.

## 3. Coverage guardrails

Every response now contributes to coverage telemetry so you can alert when
catalog breadth shrinks:

- `policy_item_served_total{namespace, surface, item_id}` – increments once per
  item delivered.
- `policy_coverage_bucket_total{namespace, surface, bucket}` – `bucket` is
  `all` or `long_tail`, letting you track exposure mix.
- `policy_catalog_items_total{namespace}` – gauge of the available catalog size.

Use PromQL (tweak the look-back window to taste):

```promql
# Unique items shown in the past hour
unique_items =
  count(count_over_time(policy_item_served_total{namespace="default"}[1h]) > 0)

# Catalog coverage ratio (target ≥ 0.60)
coverage_ratio = unique_items / max(policy_catalog_items_total{namespace="default"})

# Long-tail share (target ≥ 0.20)
long_tail_share =
  sum(rate(policy_coverage_bucket_total{namespace="default", bucket="long_tail"}[1h])) /
  sum(rate(policy_coverage_bucket_total{namespace="default", bucket="all"}[1h]))
```

Alert if `coverage_ratio < 0.60` or `long_tail_share < 0.20` for your chosen
window.

**If a guardrail fires**:

1. Re-run the quality suite or `analysis/scripts/profile_coverage.py` against
   the affected environment to confirm the regression.
2. Inspect recent deployments for changes to `POPULARITY_FANOUT`,
   `BLEND_*`, `MMR_LAMBDA`, or rule overrides that may concentrate exposure.
3. Temporarily raise exploration knobs (fan-out, ALS weight) and re-check the
   telemetry. When the targets recover, bake the updated settings into the
   environment (`COVERAGE_LONG_TAIL_HINT_THRESHOLD`, `COVERAGE_CACHE_TTL`) and
   document the change.

## 4. New-user onboarding playbook

1. **Seed starter data** via `analysis/scripts/seed_dataset.py` using the same org/namespace as production; this mirrors the catalog/users captured in `analysis_v2/evidence/seed_manifest.json`.
2. **Run scenario S7** (`make scenario-suite …`) and review `analysis_v2/evidence/scenario_s7_cold_start.json` to confirm ≥4 categories and personalization reasons for the first item.
3. **Compare segment lifts** with `analysis_v2/quality_metrics.json`; new_users should stay ≥+10% on NDCG@10/MRR@10 after any rollout (use `analysis/scripts/run_quality_eval.py` to regenerate metrics).
4. **Audit determinism** using `.github/workflows/determinism.yml` (or run `analysis/scripts/check_determinism.py --baseline analysis_v2/evidence/determinism_check.json`) before exposing a new surface.

## 5. Operational checklist

- **Before a campaign**
  - Confirm rule creation through the admin list endpoints (`/v1/admin/rules`,
    `/v1/admin/manual_overrides`).
  - Hit `/v1/admin/recommendation/presets` if you need the curated `mmr_lambda`
    presets for a surface.
  - Run `make scenario-suite` (or `python analysis/scripts/run_scenarios.py`)
    pointing at the target environment to revalidate S1–S10.

- **When something looks wrong**
  1. Check `policy_rule_zero_effect` warnings for the surface/namespace in
     question. Zero exposure means the rule matched but had no effect.
  2. Inspect the corresponding Prometheus counters to gauge impact across
     traffic.
  3. Use `/v1/audit/decisions` (with `include_reasons=true` in the original
     request) to inspect the stored trace – the policy summary and rule effects
     are persisted.
  4. Re-run the scenario suite (especially S3/S5/S8/S9) against the
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
