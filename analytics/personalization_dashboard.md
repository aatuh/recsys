# Personalisation & Overlap Monitoring (RT-3D)

This dashboard closes RT-3D by surfacing trends that prove the blend automation is keeping personalised lift healthy while guarding against homogenised results.

## Metrics & Sources

- **Personalised Impression Share** (`personalized_impression_rate`): proportion of served items marked as personalised (from decision trace extras `personalized_items` or HTTP response reasons). Ingest via existing log pipeline (`candidate_source_metrics`).
- **Reason Coverage** (`reason_coverage_rate`): % of items with at least one reason surfaced to the UI, highlighting any explainability regressions.
- **Overlap Index** (`catalog_overlap_pct`): rolling Jaccard overlap of top-K recommendations between randomly sampled users – computed offline via daily job, exported to Prometheus/Loki.
- **Brand/Category Entropy**: reuse diversity reranker logs (`CapsInfo`, `MMRInfo`) to calculate entropy per surface; higher entropy indicates variety.
- **Guardrails**: `candidate_sources.collaborative.count` zero streaks, cold-start exposure (`cold_start_impression`) for context.

## Grafana Panels

1. **Personalised Share by Surface** – stacked area chart splitting home/PDP/cart. Alert when drop >5% vs baseline.
2. **Reason Coverage** – line chart with target band (>= 80%).
3. **Top-K Overlap Heatmap** – table comparing Jaccard overlap percentiles for cohorts (new vs returning, device). Use data from nightly batch job stored in `experiments.blend_weights_daily`.
4. **Entropy Widgets** – bar chart of brand/category entropy per surface (RT-4 work).
5. **Correlated Metrics** – multi-axis chart overlaying personalised share with CTR to confirm causal lift.

## Data Pipeline

1. Extend decision tracer to emit `extras.personalized_items_count` and `extras.reason_items_count` (already tracked via TraceData). Shipping as metrics via Loki JSON log and OTEL counter exporter (`personalized_impression_rate`).
2. Batch Job (`notebook or dbt`) reads daily decision trace snapshots, computes overlap and entropy aggregates, writes to `analytics.personalization_daily` table, and pushes summary metrics to Prometheus via pushgateway.
3. Grafana dashboard JSON saved under `analytics/dashboards/personalization.json` once configured.

## Alerts

- **Personalisation Drop**: if personalised share falls below 0.55 (baseline 0.60) for 30 minutes → PagerDuty `#recsys-rotation`.
- **Overlap Spike**: if Jaccard@K > 0.35 for any surface 2 days in row → Slack `#ranking`.
- **Reason Coverage**: alert when <60% for >1 hour (likely instrumentation regression).

## Action Playbook

1. Personalisation drop but CTR steady: inspect blend overrides – roll back via `BLEND_WEIGHTS_OVERRIDES` flag.
2. Overlap spike accompanies CTR drop: deploy diversity reranker adjustments (RT-4 knobs) or reduce popularity weight.
3. Coverage alert: check shop telemetry ingestion, rerun `make codegen` if contracts changed.
