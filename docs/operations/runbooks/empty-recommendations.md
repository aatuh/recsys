# Runbook: Empty Recommendations

Use this when `POST /v1/recommend` succeeds but returns no items, or fewer items than expected.

## Fast triage

Set the local defaults:

```bash
BASE_URL=${BASE_URL:-http://localhost:8000}
TENANT_ID=${TENANT_ID:-demo}
SURFACE=${SURFACE:-home}
```

Validate the request shape before debugging data:

```bash
curl -fsS "$BASE_URL/v1/recommend/validate" \
  -H "Content-Type: application/json" \
  -H "X-Org-Id: $TENANT_ID" \
  -H "X-Dev-User-Id: local-dev" \
  -H "X-Dev-Org-Id: $TENANT_ID" \
  -d "{\"surface\":\"$SURFACE\",\"k\":10,\"user\":{\"anonymous_id\":\"debug-user\"}}"
```

Then call recommend and keep the full response for the incident record:

```bash
curl -fsS "$BASE_URL/v1/recommend" \
  -H "Content-Type: application/json" \
  -H "X-Org-Id: $TENANT_ID" \
  -H "X-Dev-User-Id: local-dev" \
  -H "X-Dev-Org-Id: $TENANT_ID" \
  -d "{\"surface\":\"$SURFACE\",\"k\":10,\"user\":{\"anonymous_id\":\"debug-user\"}}"
```

Preserve `meta.request_id`, `meta.config_version`, `meta.rules_version`, and `warnings`.

## Decision flow

1. If validation fails, fix request shape first. Empty results from an invalid payload are not a ranking incident.
2. If `warnings` includes `CANDIDATES_INCLUDE_EMPTY`, remove or correct `candidates.include_ids`.
3. If `warnings` includes `CONSTRAINTS_FILTERED`, relax required tags, forbidden tags, or per-tag caps.
4. If `warnings` includes `SIGNAL_UNAVAILABLE` or `SIGNAL_PARTIAL`, check the data mode:
   - DB-only mode: confirm source tables contain signal rows for the tenant and surface.
   - Artifact mode: check the current manifest and artifact object paths.
5. If there are no useful warnings, inspect tenant config, tenant rules, and service logs by request ID.

## Common causes

| Cause | Check | Remediation |
| --- | --- | --- |
| Surface mismatch | The request uses `home`, but signals were written under another namespace. | Align the integration surface with data production. |
| Candidate allow-list removed everything | Response warning is `CANDIDATES_INCLUDE_EMPTY`. | Remove the allow-list or use known-good item IDs. |
| Constraints removed everything | Response warning is `CONSTRAINTS_FILTERED`. | Relax constraints and retry with small `k`. |
| Missing signal data | Response warning is `SIGNAL_UNAVAILABLE` or `SIGNAL_PARTIAL`. | Rebuild or republish the missing signal, or disable the unavailable signal in config. |
| Bad rules/config rollout | Versions changed recently and empty results started afterward. | Roll back config or rules with [Rollback Config and Rules](rollback-config-rules.md). |

## Verification

After remediation:

1. Repeat the same recommendation request.
2. Confirm `items` is non-empty for the affected tenant and surface.
3. Confirm `warnings` no longer identify the root cause.
4. Keep the before/after response metadata in the incident record.

## Read next

- [Operations](../../operations.md)
- [Stale Artifact Manifest](stale-artifact-manifest.md)
- [API Reference](../../reference/api.md)
