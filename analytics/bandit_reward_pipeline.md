# Bandit Reward Feedback

This doc captures the mechanics for wiring user telemetry back into the bandit service so Thompson/UCB posteriors stay fresh.

## Event Flow
1. Recommendation response now includes `bandit` metadata (policy, bucket, request id, explore flag) and the shop UI wires it into anchor/button `data-*` attributes.
2. `ClickTelemetry` / `AddToCartButton` emit events with meta fields:
   - `bandit_policy_id`
   - `bandit_request_id`
   - `bandit_algorithm`
   - `bandit_bucket`
   - `bandit_explore`
3. `/api/events` stores the metadata and `forwardEventsBatch` forwards events to Recsys **and** calls `/v1/bandit/reward` for clicks/adds/purchases.

## Configuration
- `SHOP_BANDIT_ENABLED` toggles the entire pipeline.
- `SHOP_BANDIT_POLICY_IDS` optional allowlist of eligible policies.
- `SHOP_BANDIT_REWARD_ON_{CLICK,ADD,PURCHASE}` control which actions generate a reward.

## Monitoring
- `bandit_rewards_log` table in Recsys stores every reward.
- Add Grafana panels:
  - Reward count per surface/policy.
  - Reward latency (event time → API call).
  - Reward success rate (HTTP 2xx vs failures).
- Alert if reward stream is empty for >30 minutes while exploration is enabled.

## Next Steps
- Feed reward magnitudes (e.g., +3 for add, +5 for purchase) by extending the API payload when Recsys supports weighted rewards.
- Join bandit reward events with CTR/add/purchase dashboards to verify lift.
