---
tags:
  - ops
  - performance
---

# Performance and capacity guide

This guide describes how to run reproducible load tests against recsys-service
and capture sizing data for production planning.

## Who this is for

- Lead developers and SREs sizing `recsys-service` for production
- Engineers running load tests before enabling new signals or data modes

## What you will get

- A runnable load-test harness
- The parameters that matter for repeatability
- A table format for recording sizing data over time

## 1) Preflight checklist

- Postgres is seeded with a tenant, config, and signal data.
- recsys-service is healthy (`/healthz` returns 200).
- Auth headers are configured (dev headers or a bearer token).

## 2) Run the load test

Use the built-in harness:

```bash
./scripts/loadtest.sh
```

Key parameters (env vars):

- `BASE_URL` (default: <http://localhost:8000>)
- `ENDPOINT` (default: /v1/recommend; set /v1/similar for similar-items)
- `TENANT_ID`, `SURFACE`, `K`
- `REQUESTS`, `CONCURRENCY`
- `DEV_HEADERS=true` (local) or set `BEARER_TOKEN` / `API_KEY`

Example:

```bash
BASE_URL=http://localhost:8000 \
ENDPOINT=/v1/recommend \
TENANT_ID=demo \
SURFACE=home \
REQUESTS=1000 \
CONCURRENCY=25 \
./scripts/loadtest.sh
```

Capture:

- `rps` (requests/sec)
- p50/p95/p99 latency
- error rate (non-2xx + timeouts)

!!! note
    If you see a lot of `429` responses locally, you may be hitting the dev stack’s safety rate limit. Either lower
    `CONCURRENCY`/`REQUESTS` or use the benchmark setup in
    [`operations/baseline-benchmarks.md`](baseline-benchmarks.md).

## 3) Record sizing data

Use this table as a **living record**. Fill with measured results from your
environment (hardware, cache settings, dataset size).

| Tier  | Target QPS | p95 Latency | CPU | Memory | Notes              |
| :---- | ---------: | ----------: | --: | -----: | ------------------ |
| dev   |            |             |     |        | local, seeded data |
| small |            |             |     |        | single tenant      |
| med   |            |             |     |        | multi-tenant       |
| large |            |             |     |        | dedicated cache    |

## 4) Tuning levers

- **Cache TTLs**: `RECSYS_CONFIG_CACHE_TTL`, `RECSYS_RULES_CACHE_TTL`
- **Backpressure**: `RECSYS_BACKPRESSURE_MAX_INFLIGHT`, `RECSYS_BACKPRESSURE_MAX_QUEUE`
- **Algorithm mode**: `RECSYS_ALGO_MODE` (`blend`, `popularity`, `cooc`, etc.)
- **Artifact mode**: `RECSYS_ARTIFACT_MODE_ENABLED` (affects S3/manifest latency)

## 5) Repeat after changes

Re-run the load test after:

- schema changes (new signals)
- algorithm changes
- cache or artifact mode changes
- infrastructure changes

## Read next

- Baseline benchmarks (anchor numbers): [`operations/baseline-benchmarks.md`](baseline-benchmarks.md)
- Production readiness checklist: [`operations/production-readiness-checklist.md`](production-readiness-checklist.md)
- Backpressure and limits: [`reference/config/recsys-service.md`](../reference/config/recsys-service.md)
