# Runbook: Stale manifest (artifact mode)

## Symptoms

- Recommendations look stale even after pipelines published new artifacts
- The “current manifest” `updated_at` is older than expected for a tenant/surface
- `POST /v1/admin/tenants/{tenant_id}/cache/invalidate` with `popularity` fixes it temporarily

## Decision tree (fast path)

```mermaid
flowchart TD
  A[Serving stale artifacts] --> B{Artifact mode enabled?}
  B -->|No| C[DB-only mode: check DB signal freshness]
  B -->|Yes| D{Can you fetch the current manifest from object store?}
  D -->|No| E[Fix object store access (DNS/egress/creds/TLS)]
  D -->|Yes| F{Manifest updated recently?}
  F -->|No| G[Pipeline publish failed or scheduler not running]
  F -->|Yes| H{Service still stale after manifest TTL?}
  H -->|No| I[Wait for TTL expiry or invalidate caches]
  H -->|Yes| J[Invalidate caches; check service logs for fetch errors]
```

## Quick triage (copy/paste)

Set:

```bash
TENANT_ID=demo
SURFACE=home
BASE_URL=${BASE_URL:-http://localhost:8000}
```

1. Confirm the service is in artifact mode:

- `RECSYS_ARTIFACT_MODE_ENABLED=true` is required.
- If you’re not sure, check the service config for `RECSYS_ARTIFACT_*` variables and object store settings.

1. Fetch the current manifest from object storage.

The manifest location is defined by:

- `RECSYS_ARTIFACT_MANIFEST_TEMPLATE` (default:
  `s3://recsys-artifacts/registry/current/{tenant}/{surface}/manifest.json`)

Example with AWS CLI:

```bash
aws s3 cp "s3://recsys-artifacts/registry/current/${TENANT_ID}/${SURFACE}/manifest.json" -
```

Local dev (MinIO via docker compose):

```bash
docker compose run --rm --entrypoint sh minio-init -c \
  "mc alias set local http://minio:9000 minioadmin minioadmin >/dev/null && \
   mc cat local/recsys-artifacts/registry/current/${TENANT_ID}/${SURFACE}/manifest.json | head"
```

1. If the manifest is new but service output is still old, invalidate caches:

```bash
curl -fsS -X POST "$BASE_URL/v1/admin/tenants/${TENANT_ID}/cache/invalidate" \
  -H "Content-Type: application/json" \
  -H "X-Org-Id: $TENANT_ID" \
  -d "{\"targets\":[\"popularity\"],\"surface\":\"${SURFACE}\"}"
```

## Likely causes and safe remediations

- Manifest pointer not updated
  - Check pipeline scheduler health and recent pipeline runs.
  - See pipelines runbook: [`recsys-pipelines/docs/operations/runbooks/stale-artifacts.md`](../../recsys-pipelines/docs/operations/runbooks/stale-artifacts.md)
- TTLs are too long for your workflow
  - Tune `RECSYS_ARTIFACT_MANIFEST_TTL` and `RECSYS_ARTIFACT_CACHE_TTL`.
- Object store connectivity problems
  - Validate endpoint/creds (`RECSYS_ARTIFACT_S3_*`) from the service network.

## Verification

- Fetch the current manifest again and confirm `updated_at` advanced.
- Call `/v1/recommend` twice (after TTL expiry or cache invalidation) and confirm outputs reflect the new artifacts.

## Read next

- Artifacts and manifest lifecycle: [`explanation/artifacts-and-manifest-lifecycle.md`](../../explanation/artifacts-and-manifest-lifecycle.md)
- Data modes: [`explanation/data-modes.md`](../../explanation/data-modes.md)
- Pipelines rollback guide: [`recsys-pipelines/docs/how-to/rollback-manifest.md`](../../recsys-pipelines/docs/how-to/rollback-manifest.md)
