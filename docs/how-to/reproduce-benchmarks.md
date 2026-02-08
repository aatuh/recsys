---
diataxis: how-to
tags:
  - ops
  - performance
  - benchmark
---
# How to reproduce the baseline benchmarks

This guide shows how to reproduce the baseline measurements in a repeatable way, so you can compare changes and environments.

## 1) Start a benchmark-friendly local stack

The dev stack includes a small global rate limit (30 burst / 15 rps) for safety. For benchmarking, you can bypass it
via a skip header.

Create a benchmark env file:

```bash
test -f api/.env || cp api/.env.example api/.env
cp api/.env api/.env.benchmark
```

Then update `api/.env.benchmark`:

- Disable per-tenant rate limiting:
  - `TENANT_RATE_LIMIT_ENABLED=false`
- Enable global rate limit bypass **for local benchmarking only**:
  - `RATE_LIMIT_SKIP_ENABLED=true`
  - `RATE_LIMIT_SKIP_HEADER=X-RateLimit-Skip`
  - `RATE_LIMIT_ALLOW_DANGEROUS_DEV_BYPASSES=true`
  - `TRUSTED_PROXIES=0.0.0.0/0,::/0`

!!! warning
    Do not enable the global rate limit bypass in production. It is intended only for local/test environments.

Start the stack:

```bash
RECSYS_API_ENV_FILE=./api/.env.benchmark make cycle
```

## 2) Seed reproducible demo data/artifacts

```bash
RECSYS_API_ENV_FILE=./api/.env.benchmark ./scripts/demo.sh
```

## 3) Run the load test

`X-RateLimit-Skip: true` must be sent on requests when the bypass is enabled.

```bash
API_KEY_HEADER=X-RateLimit-Skip API_KEY=true \
BASE_URL=http://localhost:8000 ENDPOINT=/v1/recommend \
TENANT_ID=demo SURFACE=home K=20 \
REQUESTS=5000 CONCURRENCY=50 \
./scripts/loadtest.sh
```

## Recording template

Use this table as a living record (commit it as a PR comment or internal doc when you run the benchmark):

| Date | Env | Dataset | Endpoint | k | c | n | rps | p95 | Notes |
| :--- | :-- | :------ | :------- | -: | -: | -: | --: | :-- | :---- |
| YYYY-MM-DD | local docker | demo | /v1/recommend | 20 | | | | | |
| YYYY-MM-DD | staging/prod | real | /v1/recommend | 20 | | | | | |

## Read next

- [Baseline benchmarks (anchor numbers)](../operations/baseline-benchmarks.md)
- [Run locally](run-docs-locally.md)
- [Deploy with Helm](deploy-helm.md)
- [Production readiness checklist](../operations/production-readiness-checklist.md)
- [Offline gate in CI](../recsys-eval/docs/workflows/offline-gate-in-ci.md)
