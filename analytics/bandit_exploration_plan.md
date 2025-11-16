# Bandit Exploration Framework

Goal: define the policy and data plumbing necessary to explore new items without harming core metrics.

## 1. Scope
- Surfaces: home widget + PDP cross-sell (namespace `default` initially).
- Slot count: reserve 1–2 slots per impression for exploration (configurable via feature flag `bandit_explore_slots`).
- Candidate pool: filtered cold-start/fresh items (`freshness_index`) with minimum availability and price sanity checks.

## 2. Policy
- Algorithm: Thompson Sampling over Beta priors.
- Reward signal: click → +1, add-to-cart → +3, purchase → +5 (configurable weights).
- Prior parameters: α=1, β=1 (uninformative) per item, per surface.
- Decay: exponential decay with half-life 14 days to drop stale feedback.

## 3. Data Flow
1. `shop` emits event telemetry with `variant` + `bandit_policy_id`.
2. Ingestion pipeline writes bandit rewards into `bandit_rewards` table (existing).
3. Batch job (`bandit_posterior_update`) runs every 15 minutes:
   - Aggregates rewards per item/surface.
   - Updates posteriors (α, β).
   - Writes to Redis/feature store for low-latency reads.
4. Recommendation API:
   - When exploration slots active, requests top N candidates from bandit service.
   - Injects selected item(s) into recommendation response with flag `explore=true`.

## 4. Logging & Monitoring
- Decision tracer already captures `bandit` context; extend extras with `explore_slot` index.
- Metrics:
  - `bandit_explore_impressions{surface}`
  - `bandit_explore_reward_sum{surface}`
  - `bandit_explore_ctr{surface}`
- Alerts: exploration CTR < baseline -50% for 1 hour.

## 5. Controls
- Feature flag `bandit_exploration_enabled` scoped by namespace.
- Kill switch automatically reverts to exploitation if reward < threshold.
- Guard against duplicate exposure by tracking recent explored items per user for 1 hour.

## 6. Rollout
1. Dry-run in staging using synthetic traffic (ensure rewards ingest).
2. Start with 10% of traffic, 1 slot.
3. Monitor metrics/dashboards for 48 hours; adjust reward weights if variance high.

## 7. Ownership
- Ranking platform: bandit service + API integration.
- Data science: reward weights, prior tuning.
- Product analytics: dashboard + success criteria.

## 8. Controlled Experiment
- Use `BANDIT_EXPERIMENT_*` env vars to configure holdout control traffic (e.g., 10% of `home,cart`).
- Bandit decisions/rewards now log `experiment` + `variant` metadata for downstream dashboards.
- Decision tracer extras surface the same fields for quick inspection (`bandit_experiment`, `bandit_variant`).
- Roll back by setting `BANDIT_EXPERIMENT_HOLDOUT_PERCENT=0` or `BANDIT_EXPERIMENT_ENABLED=false`.
