# Tutorial: local end-to-end (pipelines -> service -> eval)

Goal: run a tiny dataset through pipelines, serve recommendations, and generate
an evaluation report.

Prereqs:

- Docker + docker compose
- curl
- POSIX shell

## 1. Start local dependencies

Bring up Postgres and recsys-service. Ensure migrations are applied.

## 2. Load demo tenant + config + rules

Create:

- tenant: demo
- config and rules (versioned, current pointer updated)

See: `reference/api/examples/admin-config.http`
See also: `reference/api/admin.md` (auth + bootstrap details, tenant insert SQL)

## 3. Load a tiny dataset

Use `tutorials/datasets/tiny/`:

- catalog.csv
- exposures.jsonl
- interactions.jsonl

In production, pipelines ingest raw events. For this tutorial you can import the
files into your canonical store.

## 4. Run pipelines to publish artifacts

### Local artifact bootstrap (MinIO)

If you want artifact/manifest mode locally, use the bundled MinIO in
`docker-compose` and publish manifests to the default bucket:

- Bucket: `recsys-artifacts`
- Manifest path: `registry/current/{tenant}/{surface}/manifest.json`
- Example URI: `s3://recsys-artifacts/registry/current/demo/home/manifest.json`

Service env (local):

```bash
RECSYS_ARTIFACT_MODE_ENABLED=true
RECSYS_ARTIFACT_MANIFEST_TEMPLATE=s3://recsys-artifacts/registry/current/{tenant}/{surface}/manifest.json
RECSYS_ARTIFACT_S3_ENDPOINT=minio:9000
RECSYS_ARTIFACT_S3_ACCESS_KEY=minioadmin
RECSYS_ARTIFACT_S3_SECRET_KEY=minioadmin
RECSYS_ARTIFACT_S3_REGION=us-east-1
RECSYS_ARTIFACT_S3_USE_SSL=false
```

Run the minimum pipelines jobs to produce non-empty recs (e.g. popularity +
manifest), then verify the manifest is present at the expected MinIO path.

See: `how-to/operate-pipelines.md`

## 5. Call the API

- POST /v1/recommend
- POST /v1/similar (optional)
- POST /v1/recommend/validate (lint requests)

See: `reference/api/examples/`

## 6. Run evaluation

Run offline regression and (optionally) experiment analysis.

See: [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)
