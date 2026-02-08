---
diataxis: explanation
tags:
  - ops
  - troubleshooting
  - evaluation
---
# Failure modes & diagnostics (baseline)
This page explains Failure modes & diagnostics (baseline) and how it fits into the RecSys suite.


## Who this is for

- On-call engineers triaging “recs are broken” incidents
- Teams rolling out a new ranking configuration and needing rollback triggers
- Analysts seeing suspicious evaluation results (bad joins, strange cliffs)

## What you will get

- A set of common failure modes with: symptom → cause → diagnosis → fix → prevention
- Links to the deeper runbooks and reference pages

## First triage (fast)

1. Is the service healthy?
   - `GET /healthz` and `GET /readyz`
2. Are responses empty or full of warnings?
   - check `warnings[]` and `meta.request_id`
3. Are you in DB-only mode or artifact/manifest mode?
   - stale manifests and missing artifacts are artifact-mode-only failures
4. Is evaluation data trustworthy?
   - schema validation + join-rate

## Failure modes

### 1) Low join-rate (evaluation data is not trustworthy)

- **Symptom**
  - `recsys-eval` reports low exposure/outcome join rates
  - KPI swings look “too good / too bad” and vary wildly by slice
- **Likely causes**
  - outcomes missing `request_id`
  - `request_id` generated twice (API call vs downstream logging)
  - the same `request_id` reused for multiple renders
  - surface/tenant keys mismatch between logs and slice keys
- **Diagnosis**
  - run `recsys-eval validate` on exposures/outcomes (and assignments if experiments)
  - compute join-rate by `surface` (and platform) from raw logs
- **Fix**
  - propagate `request_id` from recommend → render → outcome event
  - add an automated integration test that asserts “same `request_id` everywhere”
- **Prevention**
  - enforce the invariants in: [Minimum instrumentation spec (for credible evaluation)](../reference/minimum-instrumentation.md)
  - keep `request_id` generation in one place (shared middleware/client)

See: [Event join logic (exposures ↔ outcomes ↔ assignments)](../reference/data-contracts/join-logic.md)

### 2) Empty recommendations

- **Symptom**
  - response `items[]` is empty (or much shorter than `k`)
- **Likely causes**
  - no candidate data (empty popularity table in DB-only mode)
  - surface/namespace mismatch (writing signals to `home` but requesting `home_feed`)
  - constraints or allow-lists filtered everything
  - missing artifacts / stores (signal unavailable) in artifact mode
- **Diagnosis**
  - check `warnings[]` (`SIGNAL_UNAVAILABLE`, `CONSTRAINTS_FILTERED`, `CANDIDATES_INCLUDE_EMPTY`)
  - confirm tenant + surface config exists (admin bootstrap)
  - if DB-only: verify seed tables contain data for the namespace
- **Fix**
  - follow the runbook: [Runbook: Empty recs](runbooks/empty-recs.md)
- **Prevention**
  - integration checklist (one surface): [How-to: Integration checklist (one surface)](../how-to/integration-checklist.md)

### 3) Stale manifest / stale artifacts (artifact mode)

- **Symptom**
  - recommendations do not change after pipeline runs
  - results look “stuck” on an old model/version
- **Likely causes**
  - pipelines did not publish a new manifest pointer
  - object store credentials/paths misconfigured
  - service caches not invalidated after shipping
- **Diagnosis**
  - check manifest timestamp/version in the registry
  - check pipeline job logs for publish steps
  - confirm the service can read the manifest path and objects
- **Fix**
  - follow the runbook: [Runbook: Stale manifest (artifact mode)](runbooks/stale-manifest.md)
- **Prevention**
  - add a “ship verification” step: publish → invalidate cache → smoke test one request

### 4) `SIGNAL_UNAVAILABLE` / `SIGNAL_PARTIAL` warnings

- **Symptom**
  - responses include signal warnings (and quality regresses vs expectations)
- **Likely causes**
  - store backend does not implement an optional port (feature unavailable)
  - artifacts not built or not found (wrong tenant/surface path)
  - partial signal failures per-anchor (timeouts, missing neighbors)
- **Diagnosis**
  - inspect `warnings[]` details (which signal is unavailable)
  - compare requested mode/weights vs available stores
- **Fix**
  - build the missing signal/artifact (pipelines) or disable the signal in config
- **Prevention**
  - keep a baseline mode (popularity-only) available for safe fallback
  - monitor warning rates by surface

### 5) Latency spikes / timeouts

- **Symptom**
  - p95/p99 latency increases, timeouts, or elevated 5xx/429
- **Likely causes**
  - artifact/object-store calls on the hot path without cache warmth
  - backpressure thresholds too high/low for current traffic
  - expensive candidate fanout or high `k`
- **Diagnosis**
  - run the load test harness and compare p95/p99 before/after
  - check service logs/metrics for timeouts and store error rates
- **Fix**
  - reduce fanout, enable caching, tune backpressure
  - roll back the manifest/config while you investigate
- **Prevention**
  - keep a repeatable capacity baseline: [Performance and capacity guide](performance-and-capacity.md)

## Read next

- Decision playbook (ship/hold/rollback): [Decision playbook: ship / hold / rollback](../recsys-eval/docs/decision-playbook.md)
- Operations runbooks: [Operations](index.md)
- Minimum instrumentation spec: [Minimum instrumentation spec (for credible evaluation)](../reference/minimum-instrumentation.md)
