# Diversity Validation Playbook (RT-4D)

This document covers the analysis and simulation workflow to prove that the diversity constraints (MMR, caps) keep variety high without hurting engagement.

## 1. Offline Simulation Harness

- Extend the blend evaluation harness (`api/cmd/blend_eval`) with `-diversity-report` flag (future enhancement) to compute:
  - Brand/category entropy per surface (`H = -Σ p_i log p_i`).
  - Average pairwise similarity within top-K (using existing tag vectors).
  - Diversity penalty utilisation (`MMRInfo`, `CapsInfo` hit rates).
- For now, run the SQL below against decision traces to produce the same metrics:

```sql
WITH latest AS (
  SELECT namespace,
         surface,
         jsonb_array_elements(trace->'final_items') AS item,
         trace->'extras' AS extras
  FROM analytics.decision_traces
  WHERE ts >= NOW() - INTERVAL '24 hours'
)
SELECT namespace,
       surface,
       AVG((item->>'brand_entropy')::float) AS brand_entropy,
       AVG((item->>'category_entropy')::float) AS category_entropy,
       AVG((extras->>'mmr_applied')::int) AS mmr_applied_rate,
       AVG((extras->>'caps_applied')::int) AS caps_applied_rate
FROM latest
GROUP BY 1,2;
```

## 2. Grafana Dashboard (Live Metrics)

- Panels: entropy per surface, caps hit rate, similarity heatmap (bucketed by surface & user cohort).
- Alerts: entropy < baseline -10%, caps hit rate >80% (over-constraining), similarity >0.25 (homogenised list).
- Reuse `analytics/personalization_dashboard.md` data pipeline for daily aggregates; add Prometheus gauges:
  - `diversity_entropy_brand{surface}`
  - `diversity_entropy_category{surface}`
  - `diversity_caps_hit_rate{surface}`

## 3. Experiment Evaluation

- When running blend experiments (RT-3B), include diversity metrics in the results table:

| Variant | Brand Entropy | Category Entropy | Caps Hit Rate | CTR | Notes |
|---------|----------------|------------------|---------------|-----|-------|
| Baseline | 1.68 | 1.42 | 0.32 | 5.1% | Reference |
| Challenger | 1.82 | 1.56 | 0.28 | 5.4% | +0.3pp CTR |

## 4. Remediation Steps

| Issue | Action |
|-------|--------|
| Entropy dips while caps hit rate spikes | Relax caps (increase brand/category cap) or tweak MMR lambda upwards.
| Entropy stable but CTR drops | Investigate retrieval mix (RT-1 sources) vs. reranker impact; adjust sampling weights.
| High overlap + low entropy | Increase cold-start weight or session retriever diversity.

## 5. Implementation Checklist

- [ ] Add daily batch job to compute entropy + overlap aggregates (dbt/notebook) and push metrics.
- [ ] Publish Grafana dashboard under `analytics/dashboards/diversity.json`.
- [ ] Hook alerts to `#ranking` Slack channel.
- [ ] Update runbooks (RT-3B plan) to reference diversity checks before rolling out new blends.
