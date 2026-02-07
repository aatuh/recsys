---
tags:
  - how-to
  - evaluation
  - business
  - recsys-eval
---

# Decision playbook: ship / hold / rollback

## Who this is for

- Teams making **shipping decisions** based on RecSys evaluation reports
- Engineers on-call for a ranking rollout who need clear rollback triggers
- Stakeholders agreeing on “what good looks like” before running experiments

## What you will get

- A concrete decision table for **ship / hold / rollback**
- Example threshold starting points (KPI + guardrails + join-rate)
- “What to do if…” branches for common failure patterns

## The decision table (recommended baseline)

Use this in order. Don’t skip step 0: bad joins make all metrics untrustworthy.

### 0) Data integrity gate (must pass)

If any of these fail: **HOLD** and fix logging before interpreting results.

- Schemas validate (`recsys-eval validate` passes).
- Join integrity is sane:
  - Example threshold: **join-rate ≥ 95%** for the slices you care about.
  - If join-rate is lower, the most common causes are: missing/unstable `request_id`,
    wrong surface/tenant keys, or dropped events.

### 1) Safety guardrails (must not breach)

If any guardrail breaches: **ROLL BACK** (or hold with an immediate mitigation plan).

Example starting points (tune to your product/SLOs):

- Error rate: no worse than **+0.1–0.5% absolute**
- Latency: p95 no worse than **+10–20%**
- Empty-recs: no worse than **+0.2–1.0% absolute**

### 2) KPI effect (ship vs hold vs roll back)

If guardrails hold and data is valid:

- **SHIP** when primary KPI improves beyond your minimum detectable effect and results are stable across key slices.
  - Example threshold: **+1–3% relative** on your primary KPI sustained for N days.
- **HOLD** when results are inconclusive (underpowered, too noisy, conflicting slices).
  - Example threshold: KPI is within **±1% relative** (or confidence interval includes 0).
- **ROLL BACK** when primary KPI regresses meaningfully (even if some slices improved).
  - Example threshold: **≤ −1–2% relative** on your primary KPI.

## What to do if…

### KPI improved, but join-rate is low

**HOLD.** Fix instrumentation before shipping. Otherwise you risk “shipping on broken data”.

Checklist:

- Verify `request_id` is present and stable in both exposure + outcome logs.
- Confirm `tenant_id` and `surface` match between logs and evaluation slice keys.
- Re-run validation and re-compute join-rate.

### KPI improved, but latency/error/empty-recs regressed

Default action: **ROLL BACK** (or hold only if you can mitigate quickly with a safe change).

Next actions:

- Reduce blast radius (segment/surface or ramp down traffic).
- Run the relevant runbook (empty recs, stale manifest, backpressure).
- Re-run a load test and check p95/p99 before retrying the rollout.

### KPI regressed, but guardrails held

Default action: **ROLL BACK.**

If you suspect underpowering or a slice-specific mismatch, **HOLD** only long enough to:

- verify sample size / exposure volume
- check key slices (new vs returning, segments, devices, locales)
- confirm you didn’t change candidate sources/data modes unintentionally

### Offline gate failed in CI

**HOLD** the rollout. Fix the regression or update the baseline only if the change is intentional and reviewed.

## Rollback levers (mechanics)

- **Artifacts/manifest rollback (artifact mode):** roll back the manifest pointer and invalidate caches.
- **Config/rules rollback (DB-only or control-plane changes):** roll back config/rules versions and invalidate caches.

## Read next

- Suite how-to: [`how-to/run-eval-and-ship.md`](../../how-to/run-eval-and-ship.md)
- Default quality gate contract (thresholds + overrides): [`default-quality-gate.md`](default-quality-gate.md)
- Offline gates: [`recsys-eval/docs/workflows/offline-gate-in-ci.md`](workflows/offline-gate-in-ci.md)
- Interpreting results: [`recsys-eval/docs/interpreting_results.md`](interpreting_results.md)
- Failure modes & diagnostics: [`operations/failure-modes.md`](../../operations/failure-modes.md)
- Operations runbooks: [`operations/index.md`](../../operations/index.md)
