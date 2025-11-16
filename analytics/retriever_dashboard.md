# Retriever Coverage & Latency Dashboard

This dashboard validates that every retriever contributes healthy candidate sets and makes it easy to spot regressions after deploys.

## Metrics Source
- API emits structured log line `candidate_source_metrics` with fields: `source`, `count`, `duration_ms`, `surface`, `namespace`, `k`.
- Decision trace payloads now include `extras.candidate_sources` with the same counts/durations for long-term audit retention.
- Logs flow through Promtail → Loki (existing infra). Export to Prometheus via `loki_push_api` if SLO-style alerting is required.

## Grafana Panels
1. **Coverage by Source (Stacked Area)**: `sum by (source)` of `count` over time, filtered by `namespace` or `surface`.
2. **Latency by Source (Heatmap)**: `quantile_over_time(0.95, duration_ms)` to catch spikes; color by source.
3. **Personalization Coverage**: table highlighting zero-count windows for `collaborative`, `content`, or `session`.
4. **Merge vs. Post-Exclusion Counts**: compare `merged` vs `post_exclusion` averages to validate seen-item filters and caps.
5. **Recent Trace Samples**: dynamic table linking to audit records when `count` deviates +/-20% from previous hour.

## Alerting Rules
- **Collaborative Gap**: alert if `avg(count{source="collaborative"})` == 0 for 5 minutes on any active namespace.
- **Latency Spike**: alert when `p95(duration_ms)` for any source exceeds 120 ms for 10 minutes.
- **Merge Drop**: warn if `post_exclusion` < `k` for more than 3 consecutive evaluation periods.

## Implementation Checklist
- [ ] Update Promtail config with JSON label extraction for `candidate_source_metrics`.
- [ ] Create Grafana dashboard JSON and commit under `analytics/dashboards/retriever.json`.
- [ ] Hook alerts to PagerDuty `#recsys-rotation` and Slack `#ranking`.
- [ ] Backfill historical traces (if needed) by replaying audit logs through the metrics pipeline.

Refer to `api/docs/candidate-audit.md` for the qualitative analysis flow; this dashboard adds quantitative observability.
