---
diataxis: how-to
tags:
  - recsys-algo
  - ranking
  - tuning
  - evaluation
---
# How-to: tune ranking safely

Use this playbook to adjust ranking behavior **without losing auditability**.

## Before you start

- Decide the goal: one primary KPI + 2–5 guardrails
- Ensure you can measure (stable `request_id` and exposure/outcome logs)

Start points:

- Success metrics: [Success metrics](../for-businesses/success-metrics.md)
- Minimum instrumentation spec: [Minimum instrumentation spec](../reference/minimum-instrumentation.md)

## Step 1 — Understand the current ranking contract

Read the core behavior and determinism rules:

- Ranking & constraints reference: [Ranking & constraints reference](../recsys-algo/ranking-reference.md)
- Scoring model specification: [Scoring model spec](../recsys-algo/scoring-model.md)

## Step 2 — Choose the smallest knob

Prefer the smallest change that can be rolled back cleanly:

1. **Config/rules change (no code)**
   - Use weights/limits/flags per tenant and merchandising constraints.
   - Best for quick iteration and safe rollback.

2. **Pipeline change (signals/data)**
   - Add or adjust a signal end-to-end.

3. **Ranking code change**
   - Use only when the scoring/merge logic must change.
   - Requires a stricter evaluation and review.

## Step 3 — Create a candidate and keep it reproducible

- Record the baseline version (config/rules/algo versions)
- Record the candidate version (exact diffs)
- Keep artifacts immutable (avoid "silent" rewrites)

Helpful reading:

- Artifacts and manifest lifecycle: [Artifacts and manifest lifecycle](../explanation/artifacts-and-manifest-lifecycle.md)

## Step 4 — Run offline evaluation gates

Run an evaluation report and interpret it as a decision artifact:

- Workflow: [Run eval and ship](run-eval-and-ship.md)
- Interpretation orientation: [Interpreting metrics and reports](../explanation/metric-interpretation.md)

## Step 5 — Validate determinism and joinability

These prevent "it worked on my laptop" outcomes:

- Verify determinism: [Verify determinism](../tutorials/verify-determinism.md)
- Verify joinability: [Verify joinability](../tutorials/verify-joinability.md)

## Step 6 — Ship with rollback discipline

- Run at least one rollback drill before you need it.
- Write the decision and link it to the report and evidence kit.

Start here:

- Rollback model: [Operational reliability & rollback](../start-here/operational-reliability-and-rollback.md)
- Evidence kit template: [Evidence](../for-businesses/evidence.md)

## Read next

- RecSys engineering hub: [RecSys engineering hub](../recsys-engineering/index.md)
- Add a signal end-to-end: [Add a signal end-to-end](add-signal-end-to-end.md)
- Run eval and make ship decisions: [Run eval and ship](run-eval-and-ship.md)
- Ranking reference: [Ranking & constraints reference](../recsys-algo/ranking-reference.md)
