# Deploy with Helm (production-ish)

## Who this is for

- Platform engineers deploying `recsys-service` to Kubernetes
- Teams running artifact/manifest mode with external Postgres + S3/MinIO

## Goal

Install `recsys-service` and optionally a `recsys-pipelines` CronJob.
Postgres and MinIO are **disabled by default** so you can bring your own.

## Prereqs

- Helm 3 + `kubectl`
- A Postgres database (or enable the chart’s bundled Postgres for local demos)
- An S3-compatible bucket (or enable the chart’s bundled MinIO for local demos)

## Steps

### 1) Install (BYO Postgres + S3)

```bash
RECSYS_ARTIFACT_MANIFEST_TEMPLATE='s3://recsys-artifacts/registry/current/{tenant}/{surface}/manifest.json'

helm install recsys ./charts/recsys \
  --set api.env.DATABASE_URL='postgres://user:pass@db:5432/recsys?sslmode=disable' \
  --set api.env.RECSYS_ARTIFACT_MANIFEST_TEMPLATE="${RECSYS_ARTIFACT_MANIFEST_TEMPLATE}" \
  --set api.env.RECSYS_ARTIFACT_S3_ENDPOINT='s3.example.com' \
  --set api.env.RECSYS_ARTIFACT_S3_ACCESS_KEY='***' \
  --set api.env.RECSYS_ARTIFACT_S3_SECRET_KEY='***'
```

### 2) Local demo (enable bundled Postgres + MinIO)

```bash
# kind
./scripts/helm_local.sh --kind

# minikube
./scripts/helm_local.sh --minikube
```

### 3) Enable pipelines CronJob

```bash
helm upgrade --install recsys ./charts/recsys \
  --set pipelines.enabled=true \
  --set pipelines.schedule='0 2 * * *'
```

The CronJob reads `pipelines.configJson` from a ConfigMap. Override it in
`values.yaml` for your tenant, surfaces, and storage endpoints.

## Verify

```bash
kubectl get deploy,svc,cronjob
kubectl logs deploy/recsys-api
```

## Pitfalls

- The chart uses the `DATABASE_URL` env var as defined in `api/.env.example`.
- If you disable bundled Postgres/MinIO, you **must** provide external endpoints.
- The pipelines job requires object store + registry configured in its config.

## Read next

- Operational invariants (pipelines safety model): [`explanation/pipelines-operational-invariants.md`](../explanation/pipelines-operational-invariants.md)
- Artifacts and manifest lifecycle: [`explanation/artifacts-and-manifest-lifecycle.md`](../explanation/artifacts-and-manifest-lifecycle.md)
- Production readiness checklist: [`operations/production-readiness-checklist.md`](../operations/production-readiness-checklist.md)
