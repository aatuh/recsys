# How-to: operate recsys-pipelines

## Who this is for

- Data engineers and platform engineers operating `recsys-pipelines`
- SRE/on-call who need a concrete ship/rollback model for artifact mode

## Goal

Keep artifacts fresh and safely ship/rollback manifests for each `(tenant, surface)` pair.

## Prereqs

- Decide your data mode:
  - DB-only (signals in Postgres)
  - artifact/manifest mode ([artifacts](../project/glossary.md#artifact) in an object store + a
    [manifest](../project/glossary.md#manifest) pointer)
- If using artifact/manifest mode:
  - object store credentials are configured for pipelines and the service
  - `RECSYS_ARTIFACT_MANIFEST_TEMPLATE` points at your “current manifest” convention

## Steps

### Daily operation (artifact/manifest mode)

- Ingest exposure/outcome events.
- Validate and canonicalize.
- Build artifacts (start with popularity).
- Publish artifacts and swap the manifest pointer last.
- Monitor freshness, volume anomalies, and output sizes.

### Backfills

- Compute artifacts for explicit time windows.
- Publish and swap the manifest pointer to the new version.
- Keep prior manifests available for rollback.

### Rollback

- Swap the manifest pointer back to a last-known-good version.
- Invalidate service caches (or wait for TTL) to reduce “stale manifest” confusion.

## Local MinIO example (docker-compose default)

- Bucket: `${MINIO_BUCKET:-recsys-artifacts}`
- Manifest path convention: `registry/current/{tenant}/{surface}/manifest.json`
- Example manifest URI:

  `s3://recsys-artifacts/registry/current/demo/home/manifest.json`

### Service env (artifact mode)

```bash
RECSYS_ARTIFACT_MODE_ENABLED=true
RECSYS_ARTIFACT_MANIFEST_TEMPLATE=s3://recsys-artifacts/registry/current/{tenant}/{surface}/manifest.json
RECSYS_ARTIFACT_S3_ENDPOINT=minio:9000
RECSYS_ARTIFACT_S3_ACCESS_KEY=minioadmin
RECSYS_ARTIFACT_S3_SECRET_KEY=minioadmin
RECSYS_ARTIFACT_S3_REGION=us-east-1
RECSYS_ARTIFACT_S3_USE_SSL=false
```

## Verify

- Pipelines produced a manifest for your tenant/surface (local filesystem registry example):

  ```bash
  cat .out/registry/current/demo/home/manifest.json
  ```

- The service can read the manifest (no `artifact incompatible` errors) and returns non-empty results for a seeded
  tenant/surface.

## Pitfalls

### `registry_dir` location matters

- If `registry_dir` points to `s3://.../registry`, pipelines will write
  manifests directly to MinIO and you **do not** need a manual upload step.
  A local path (e.g. `registry`) requires uploading the manifest yourself.

## DB-only mode (simplest pilot)

- write signals into Postgres tables instead of publishing artifacts
- useful for local MVPs and popularity-only pilots
- seed examples: `reference/database/db-only-seeding.md`

## Read next

- Operational invariants (pipelines safety model): [`explanation/pipelines-operational-invariants.md`](../explanation/pipelines-operational-invariants.md)
- Artifacts and manifest lifecycle: [`explanation/artifacts-and-manifest-lifecycle.md`](../explanation/artifacts-and-manifest-lifecycle.md)
- Pipelines SLOs and freshness: [`recsys-pipelines/docs/operations/slos-and-freshness.md`](../recsys-pipelines/docs/operations/slos-and-freshness.md)
