# Cold-Start Performance Dashboard (RT-8C)

## Metrics
- `cold_start_impression` (counter) with labels: `surface`, `widget`, `user_segment` (future), `item_id` (top N for table).
- `cold_start_click`, `cold_start_add`, `cold_start_purchase` counters mirroring the label set.
- Derived KPIs: CTR = clicks/impressions, Add-to-Cart Rate = adds/impressions, CVR = purchase/impressions.
- Inventory coverage: unique cold-start SKUs shown per day.

## Data Sources
- Structured logs emitted by shop API (`[metrics] cold_start_* ...`).
- Shipping plan: Promtail → Loki for log ingestion; optional Prometheus exporter later.
- Combine with existing event warehouse tables for blended reporting.

## Dashboard Panels
1. **Exposure Funnel**: stacked area (impression/click/add/purchase) over time.
2. **CTR by Surface**: per-surface lines with target annotations.
3. **Top Cold-Start Items**: table sorted by impressions + CTR, links to product detail.
4. **Coverage Heatmap**: calendar heatmap of distinct cold-start items / day.
5. **Alert thresholds**: highlight when CTR falls below baseline or impressions drop >30% WoW.

## Alerting
- Grafana alert rule: if 6h rolling CTR < 0.02 or impressions < 100 → Slack `#merchandising`.
- Secondary alert: Add-to-cart rate zero for 2 hours.

## Implementation To-Dos
- Configure Promtail scraping shop logs with label `service=shop`.
- Define Loki queries for each metric (regex extract JSON payload).
- Export dashboard JSON and store under `analytics/dashboards/` once Grafana configured.
- Coordinate with data engineering to persist metrics in warehouse for historical reporting.
