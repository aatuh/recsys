---
diataxis: explanation
tags:
  - business
  - benchmarks
  - trust
  - performance
---
# Benchmarks and methodology

Benchmarks are credibility tools, not marketing numbers.

This page explains **what you can reasonably measure**, **how to measure it**, and **how to record results** so they are comparable over time.

## What we benchmark

We focus on three benchmark categories that matter during procurement:

1. **Serving performance** (latency/throughput for `POST /v1/recommend`)
2. **Pipelines performance** (how long artifacts and manifests take to build)
3. **Evaluation runtime** (how long offline reports take to produce)

## What you should not expect

- These numbers are not “your production numbers.”
- These numbers do not imply business lift.
- Cross-company comparisons are misleading unless environments are comparable.

## Reproducible baseline benchmarks

The suite includes baseline runs and a template to record your own results:

- Baseline benchmarks (ops): [Baseline benchmarks (anchor numbers)](../operations/baseline-benchmarks.md)
- Performance and capacity: [Performance and capacity guide](../operations/performance-and-capacity.md)

### Minimal benchmark protocol (recommended)

Run this protocol in a clean local environment first (10–20 minutes):

1. Run the local end-to-end tutorial:
   - [local end-to-end (service → logging → eval)](../tutorials/local-end-to-end.md)
2. Record:
   - host specs (CPU / RAM / disk)
   - docker versions
   - dataset size (items/users)
   - data mode (DB-only vs artifact/manifest)
3. Run the included baseline scripts and write down results.

!!! tip "How to share results internally"
    Paste your recorded results into your evaluation document and link to the exact git commit + manifest id.

## Benchmark validity checklist

Use this checklist to keep your benchmarks meaningful:

- [ ] You know the workload (surface count, `k`, filters)
- [ ] You capture p50/p95/p99 latency (not only averages)
- [ ] You record the dataset size (items/users)
- [ ] You record artifact versions and config versions
- [ ] You record cache behavior (cold vs warm)

## How benchmarks connect to procurement

Benchmarks should answer:

- “Will this fit inside our latency budget?”
- “What will it cost us to run?”
- “How hard is it to operate?”

See also:

- TCO and effort: [TCO and effort model](tco-and-effort.md)
- Procurement pack: [Procurement pack (Security, Legal, IT, Finance)](procurement-pack.md)
- Evidence (example outputs): [Evidence (what “good outputs” look like)](evidence.md)

## Read next

- TCO and effort: [TCO and effort model](tco-and-effort.md)
- Evidence: [Evidence (what “good outputs” look like)](evidence.md)
- Limitations: [Known limitations and non-goals (current)](../start-here/known-limitations.md)
