# Blend Weights Online Experiment Plan (RT-3B)

## Objective
Validate that updated blend weights (alpha/beta/gamma) improve engagement and monetisation metrics without hurting guardrails. The experiment compares the current “baseline” blend from RT-3A against one or more challenger configurations derived from the offline harness.

## Scope
- **Surfaces**: Shop home widget, PDP recommendations, cart cross-sell (all namespace `default` initially).
- **Traffic split**: 50/50 control vs. single challenger. Future iterations can extend to multi-armed or bandit-style assignment once impact is proven.
- **Duration**: Minimum 14 full days to capture weekday/weekend patterns or until statistical power reached.

## Experiment Setup
1. **Candidate Selection**  
   - Use the offline harness (`go run ./api/cmd/blend_eval …`) to shortlist two challenger configs with higher hit-rate/MRR and acceptable coverage.
   - Pick one challenger for the first online run to reduce complexity.
2. **Assignment**  
   - Leverage the existing configuration service / feature flagging (RecSys config service or LaunchDarkly).  
   - Key: `blend_weights_v2` with variants `{baseline, challenger_a}`.  
   - Stickiness: userID hash to ensure consistent experience.
   - Exposure logging: add variant name to request context (e.g. `ctx.variant=blend_challenger_a`) so decision traces capture cohort.
3. **Runtime Overrides**  
   - Extend recommendation service to read the flag and inject override weights (`Overrides.Blend*`).  
   - Guard with kill switch to fall back to baseline instantly.

## Metrics
| Category | Metric | Description | Target |
|----------|--------|-------------|--------|
| Primary | CTR | Clicks / impressions per surface | +3% relative |
| Primary | Add-to-cart rate | Adds / impressions | ≥ baseline |
| Secondary | Revenue per mille | GMV per 1000 impressions | +2% relative |
| Personalisation | Personalised share | % of recs with personalised reasons | ≥ baseline |
| Guardrail | Out-of-stock rate | % recs for unavailable items | ≤ baseline |
| Guardrail | Candidate coverage | Average unique items per 1k impressions (from `candidate_sources`) | ≥ baseline |

## Instrumentation
- **Decision trace**: already includes `candidate_sources`, `segment_id`, etc. Add `experiment_variant` field to `Trace.Extras`.
- **Shop telemetry**: ensure click/add/purchase events include `variant` tag (extend `shop/src/components/ClickTelemetry.tsx` pipeline).
- **Dashboards**:  
  - Grafana panels comparing CTR/Add/Rev across variants (time series + relative delta).  
  - Loki queries for guardrails (personalisation hit rate, coverage).  
  - Alert when CTR delta < -5% for 2 consecutive hours.
- **Data warehouse**: land nightly aggregates into `experiments.blend_weights_daily` for deeper analysis (use existing ETL job pattern).
- **Config management**: leverage `BLEND_WEIGHTS_OVERRIDES` in the API env to seed the challenger weights, with feature-flag service toggling namespace assignments at runtime.
- **Personalisation dashboard**: wire the new `personalization_dashboard` view (RT-3D) into Grafana so experiment analysis is co-located with live guardrails.

## Analysis
1. Check power assumptions (baseline CTR ~5%, desired lift 3%, α=0.05 → ~600k impressions/arm).  
2. Monitor sequential metrics daily for safety; final decision after full horizon.  
3. Statistical test: two-sample z-test for proportions (CTR/Add), Welch t-test for revenue.  
4. Segment cuts: new vs. returning users, device type, surface.

## Rollout Timeline
1. **T-7 days**: Finalise challenger weights, configure feature flag, deploy runtime override support.  
2. **T-2 days**: Dry run in staging (`make dev`) to confirm logging and dashboards.  
3. **Day 0**: Start experiment with 1% ramp, validate metrics.  
4. **Day 1**: Ramp to 50% if guardrails OK.  
5. **Day 14**: Analyse results; if positive, plan staged rollout to 100% and archive experiment.

## Risks & Mitigations
- **Cold-start regressions**: monitor cold-start dashboards (RT-6/RT-8 work) for traffic drops; revert to baseline if hit.  
 - **Segment bias**: ensure assignment hash includes namespace to avoid cross-surface contamination.  
- **Operational**: document rollback steps (flip flag to `baseline` + redeploy).

## Owners
- Experiment lead: Relevance PM  
- Data analyst: Growth analytics  
- On-call engineer: Ranking / infrastructure team
