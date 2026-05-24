# Observability

RecSys exposes Prometheus metrics at `/metrics` when the service runs with the default toolkit middleware. The
repository includes starter monitoring assets:

- `observability/prometheus-alerts.yaml`
- `observability/grafana-dashboard.json`

Treat these as templates. Tune thresholds to your catalog, traffic, latency budget, and rollback tolerance.

## Key Signals

| Signal | Metric |
| --- | --- |
| Request rate and outcomes | `recsys_recommendation_requests_total` |
| Latency | `recsys_recommendation_latency_seconds` |
| Returned item count | `recsys_recommendation_returned_items` |
| Warning count | `recsys_recommendation_warnings` |
| Artifact load failures | `recsys_artifact_load_failures_total` |
| Manifest freshness | `recsys_artifact_manifest_age_seconds` |

The built-in labels intentionally avoid tenant IDs, request IDs, user IDs, and artifact URIs. Use logs with request IDs
for detailed incident reconstruction.

## First Alerts

Start with alerts for:

- error or overload rate above the agreed guardrail,
- empty recommendation rate above the agreed guardrail,
- p95 latency regression,
- stale manifests,
- artifact load failures.

When an alert fires, use the operations runbooks for empty recommendations, stale manifests, and service readiness.
