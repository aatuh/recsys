# Runbook: Stale Artifact Manifest

Use this when artifact mode is enabled and recommendations do not reflect recently published pipeline outputs.

## Fast triage

Set the affected tenant and surface:

```bash
TENANT_ID=${TENANT_ID:-demo}
SURFACE=${SURFACE:-home}
BASE_URL=${BASE_URL:-http://localhost:8000}
```

Confirm artifact mode is intended for the environment:

```bash
rg "RECSYS_ARTIFACT_" api/.env api/.env.example
```

The service reads the current manifest from `RECSYS_ARTIFACT_MANIFEST_TEMPLATE`. The local/default shape is:

```text
s3://recsys-artifacts/registry/current/{tenant}/{surface}/manifest.json
```

## Decision flow

1. If artifact mode is off, this is a DB-only freshness issue. Check source data and skip this runbook.
2. If the current manifest is missing or unreadable, fix object-store path, credentials, DNS, or TLS.
3. If the manifest is old, check whether `recsys-pipelines` published successfully.
4. If the manifest is current but serving is stale, wait for manifest TTL expiry or invalidate service caches.
5. If invalidation temporarily fixes the issue, revisit `RECSYS_ARTIFACT_MANIFEST_TTL` and the publish workflow.

## Local checks

For local filesystem proof-kit output, inspect the current manifest directly:

```bash
test -s tmp/commercial-proof-kit/pipelines/registry/current/demo/home/manifest.json
python3 -m json.tool tmp/commercial-proof-kit/pipelines/registry/current/demo/home/manifest.json
```

For local MinIO or S3-compatible deployments, fetch the path that matches your configured manifest template. Do not
paste credentials into tickets or public issues.

## Cache invalidation

When the manifest is known-good and the service is allowed to read it, invalidate popularity/artifact-related caches:

```bash
curl -fsS -X POST "$BASE_URL/v1/admin/tenants/$TENANT_ID/cache/invalidate" \
  -H "Content-Type: application/json" \
  -H "X-Org-Id: $TENANT_ID" \
  -H "X-Dev-User-Id: local-dev" \
  -H "X-Dev-Org-Id: $TENANT_ID" \
  -d "{\"targets\":[\"popularity\"],\"surface\":\"$SURFACE\"}"
```

## Verification

1. Fetch the manifest and confirm `updated_at` or current artifact URIs changed.
2. Call `POST /v1/recommend` for the affected tenant and surface.
3. Confirm response quality, warning rates, and empty recommendation rate recover.
4. Record whether the recovery required cache invalidation, TTL expiry, or a new pipeline publish.

## Read next

- [Artifacts and Pipelines](../../artifacts-and-pipelines.md)
- [Empty Recommendations](empty-recommendations.md)
- [Configuration Reference](../../reference/config.md)
