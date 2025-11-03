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

## 3. Operational checklist

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

## 4. Useful commands

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
