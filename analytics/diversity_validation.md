# Diversity Validation Playbook

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

When running blend experiments, log diversity metrics alongside engagement:

- **Baseline** — Brand entropy 1.68, category entropy 1.42, caps hit rate 0.32, CTR 5.1%. Notes: reference.
- **Challenger** — Brand entropy 1.82, category entropy 1.56, caps hit rate 0.28, CTR 5.4%. Notes: +0.3pp CTR.

## 4. Remediation Steps

- **Entropy dips while caps hit rate spikes** — Relax caps (increase brand/category cap) or raise `MMR_LAMBDA`.
- **Entropy stable but CTR drops** — Examine retrieval mix vs. reranker impact; adjust sampling weights.
- **High overlap + low entropy** — Increase cold-start weight or diversify the session retriever.

## 5. Implementation Checklist

- [ ] Add daily batch job to compute entropy + overlap aggregates (dbt/notebook) and push metrics.
- [ ] Publish Grafana dashboard under `analytics/dashboards/diversity.json`.
- [ ] Hook alerts to `#ranking` Slack channel.
- [ ] Update runbooks to reference diversity checks before rolling out new blends.
