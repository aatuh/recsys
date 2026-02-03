# How-to: operate recsys-pipelines

Daily:

- ingest exposures + outcomes
- validate/canonicalize
- build artifacts (start with popularity)
- publish atomically (update manifest pointer)
- monitor freshness, volume anomalies, output sizes

Backfills:

- compute for explicit time windows
- publish new manifest version
- keep prior manifest for rollback

Rollback:

- pointer swap to last good manifest
- invalidate service caches

Local MinIO example (docker-compose default):

- Bucket: `${MINIO_BUCKET:-recsys-artifacts}`
- Manifest path convention: `registry/current/{tenant}/{surface}/manifest.json`
- Example manifest URI:

  `s3://recsys-artifacts/registry/current/demo/home/manifest.json`

Service env (artifact mode):

```bash
RECSYS_ARTIFACT_MODE_ENABLED=true
RECSYS_ARTIFACT_MANIFEST_TEMPLATE=s3://recsys-artifacts/registry/current/{tenant}/{surface}/manifest.json
RECSYS_ARTIFACT_S3_ENDPOINT=minio:9000
RECSYS_ARTIFACT_S3_ACCESS_KEY=minioadmin
RECSYS_ARTIFACT_S3_SECRET_KEY=minioadmin
RECSYS_ARTIFACT_S3_REGION=us-east-1
RECSYS_ARTIFACT_S3_USE_SSL=false
```

Tip:

- If `registry_dir` points to `s3://.../registry`, pipelines will write
  manifests directly to MinIO and you **do not** need a manual upload step.
  A local path (e.g. `registry`) requires uploading the manifest yourself.

DB-only mode:

- write signals into Postgres tables instead of publishing artifacts
- useful for local MVPs and popularity-only pilots
- seed examples: `reference/database/db-only-seeding.md`
