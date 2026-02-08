---
diataxis: reference
tags:
  - ops
  - performance
  - benchmark
---
# Baseline benchmarks (anchor numbers)
This page is the canonical reference for Baseline benchmarks (anchor numbers).


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

## Reproducing these numbers

To reproduce the baseline measurements on your own hardware and dataset, follow:
- **[How to reproduce the baseline benchmarks](../how-to/reproduce-benchmarks.md)**
