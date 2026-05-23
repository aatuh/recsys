# Runbook: Roll Back Config and Rules

Use this when a tenant config or rules change causes empty recommendations, elevated errors, or a quality regression.

## Safety rules

- Prefer the admin API path because it validates payloads and records audit events.
- Capture the current failing response before changing state.
- Use `If-Match` with the current ETag when possible to avoid overwriting a concurrent update.
- Do not paste config, rules, audit details, or customer data into public issues.

## Capture current state

Set the affected tenant:

```bash
BASE_URL=${BASE_URL:-http://localhost:8000}
TENANT_ID=${TENANT_ID:-demo}
```

Fetch the current documents:

```bash
curl -i -fsS "$BASE_URL/v1/admin/tenants/$TENANT_ID/config" \
  -H "X-Org-Id: $TENANT_ID" \
  -H "X-Dev-User-Id: local-dev" \
  -H "X-Dev-Org-Id: $TENANT_ID"

curl -i -fsS "$BASE_URL/v1/admin/tenants/$TENANT_ID/rules" \
  -H "X-Org-Id: $TENANT_ID" \
  -H "X-Dev-User-Id: local-dev" \
  -H "X-Dev-Org-Id: $TENANT_ID"
```

Record the response body and `ETag` header for each document.

## Find the previous payload

List recent audit events:

```bash
curl -fsS "$BASE_URL/v1/admin/tenants/$TENANT_ID/audit?limit=50" \
  -H "X-Org-Id: $TENANT_ID" \
  -H "X-Dev-User-Id: local-dev" \
  -H "X-Dev-Org-Id: $TENANT_ID"
```

Look for `config.update` or `rules.update`. Use the prior payload from the audit detail only if your role is allowed to
read those fields. If audit details are not visible, recover the last-known-good JSON from deployment records, incident
notes, or a checked release artifact.

## Reapply the previous payload

Save the previous payload to a local file such as `/tmp/recsys-config-rollback.json`, then apply it:

```bash
curl -fsS -X PUT "$BASE_URL/v1/admin/tenants/$TENANT_ID/config" \
  -H "Content-Type: application/json" \
  -H "If-Match: W/\"current-config-etag\"" \
  -H "X-Org-Id: $TENANT_ID" \
  -H "X-Dev-User-Id: local-dev" \
  -H "X-Dev-Org-Id: $TENANT_ID" \
  --data-binary @/tmp/recsys-config-rollback.json
```

For rules, use the same pattern against `/v1/admin/tenants/$TENANT_ID/rules`.

If the API returns a version conflict, fetch the current document again, confirm no other operator is handling the same
incident, and retry with the new ETag.

## Invalidate caches

After config or rules rollback:

```bash
curl -fsS -X POST "$BASE_URL/v1/admin/tenants/$TENANT_ID/cache/invalidate" \
  -H "Content-Type: application/json" \
  -H "X-Org-Id: $TENANT_ID" \
  -H "X-Dev-User-Id: local-dev" \
  -H "X-Dev-Org-Id: $TENANT_ID" \
  -d '{"targets":["config","rules","popularity"]}'
```

`popularity` is relevant for artifact-related cache state and is harmless when no matching cache exists.

## Verification

1. Repeat the failing recommendation request.
2. Confirm `meta.config_version` and `meta.rules_version` match the expected rolled-back documents.
3. Confirm empty recommendation rate, error rate, and latency recover.
4. Add the rollback versions and request IDs to the incident record.

## Read next

- [Empty Recommendations](empty-recommendations.md)
- [Evaluation Decisions](../../evaluation-decisions.md)
- [API Reference](../../reference/api.md)
