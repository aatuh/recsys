# Deploy with Helm (production-ish)

This chart installs the **recsys-service** and optionally a **pipelines CronJob**.
Postgres and MinIO are **disabled by default** so you can bring your own.

## 1) Install (BYO Postgres + S3)

```bash
RECSYS_ARTIFACT_MANIFEST_TEMPLATE='s3://recsys-artifacts/registry/current/{tenant}/{surface}/manifest.json'

helm install recsys ./charts/recsys \
  --set api.env.DATABASE_URL='postgres://user:pass@db:5432/recsys?sslmode=disable' \
  --set api.env.RECSYS_ARTIFACT_MANIFEST_TEMPLATE="${RECSYS_ARTIFACT_MANIFEST_TEMPLATE}" \
  --set api.env.RECSYS_ARTIFACT_S3_ENDPOINT='s3.example.com' \
  --set api.env.RECSYS_ARTIFACT_S3_ACCESS_KEY='***' \
  --set api.env.RECSYS_ARTIFACT_S3_SECRET_KEY='***'
```

## 2) Local demo (enable bundled Postgres + MinIO)

```bash
# kind
./scripts/helm-local.sh --kind

# minikube
./scripts/helm-local.sh --minikube
```

## 3) Enable pipelines CronJob

```bash
helm upgrade --install recsys ./charts/recsys \
  --set pipelines.enabled=true \
  --set pipelines.schedule='0 2 * * *'
```

The CronJob reads `pipelines.configJson` from a ConfigMap. Override it in
`values.yaml` for your tenant, surfaces, and storage endpoints.

## 4) Verify

```bash
kubectl get deploy,svc,cronjob
kubectl logs deploy/recsys-api
```

## Notes

- The chart uses the `DATABASE_URL` env var as defined in `recsys/api/.env.example`.
- If you disable bundled Postgres/MinIO, you **must** provide external endpoints.
- The pipelines job requires object store + registry configured in its config.
