---
tags:
  - ops
  - performance
  - benchmark
---

# Baseline benchmarks (anchor numbers)

## Who this is for

- Operators and engineers who want “order of magnitude” expectations before running their own load tests
- Anyone who wants a reproducible harness to compare changes over time (signals, caches, data modes)

## What you will get

- One **sample baseline** benchmark result (reproducible demo dataset)
- A **command to reproduce** the same run in your environment
- A template to record your own numbers over time

## Sample baseline (demo dataset, local Docker)

This is intentionally a small, reproducible baseline. Treat it as an anchor, not a promise:
numbers depend heavily on hardware, Docker resource limits, dataset size, and enabled signals.

### Environment

- OS: Linux (x86_64)
- CPU: AMD Ryzen 7 3700X (8C/16T)
- Memory: 31 GiB

### Dataset + config

- Dataset: `./scripts/demo.sh` (10 events ingested; 4 items in popularity; 4 co-occurrence rows)
- Endpoint: `POST /v1/recommend`
- `k=20`
- Algorithm: `RECSYS_ALGO_MODE=popularity`
- Data mode: `RECSYS_ARTIFACT_MODE_ENABLED=true` (manifest in local MinIO)

### Result (2026-02-06)

From `./scripts/loadtest.sh` (5000 requests, concurrency 50):

- Throughput: **3557 rps**
- Latency: **p50=12.8ms, p95=22.8ms, p99=39.3ms**
- Errors: **0** (200: 5000)

## How to reproduce

### 1) Start a benchmark-friendly local stack

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

### 2) Seed reproducible demo data/artifacts

```bash
RECSYS_API_ENV_FILE=./api/.env.benchmark ./scripts/demo.sh
```

### 3) Run the load test

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

- Performance and capacity guide: [`operations/performance-and-capacity.md`](performance-and-capacity.md)
- Backpressure and limits: [`reference/config/recsys-service.md`](../reference/config/recsys-service.md)
