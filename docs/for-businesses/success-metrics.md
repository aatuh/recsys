---
tags:
  - overview
  - business
  - evaluation
---

# Success metrics (KPIs, guardrails, and exit criteria)

## Who this is for

- Stakeholders defining “what success means” for a RecSys pilot
- Product + analytics teams setting measurement and guardrails

## What you will get

- A practical KPI + guardrail template you can reuse
- A minimal set of exit criteria that prevents “shipping on broken data”
- Links to the exact evaluation workflow pages

## The template (recommended)

For each surface you pilot:

1. Pick **one primary KPI** that represents business value.
1. Pick **2–4 guardrails** that must not regress.
1. Define **success criteria** and **rollback criteria** up front.

Example KPI choices:

- ecommerce: conversion rate, revenue per session, add-to-cart rate
- content: completion rate, time spent, return rate

Common guardrails:

- latency and error rate
- empty-recs rate
- join-rate / instrumentation integrity (for measurement)

## Exit criteria (minimum to call the pilot “credible”)

- You can validate logs against schemas (no “unknown fields” surprises).
- Join integrity is sane (broken joins invalidate metrics).
- You can produce at least one report that compares baseline vs candidate.
- You have practiced rollback once (config/rules and/or manifest pointer).

## How we measure (suite-level workflow)

Start here:

- How to run evaluation and decide ship/hold/rollback: [Run eval and ship](../how-to/run-eval-and-ship.md)
- Ship/hold/rollback decision playbook: [Decision playbook](../recsys-eval/docs/decision-playbook.md)

Then use the `recsys-eval` workflow pages:

- [Offline gate in CI](../recsys-eval/docs/workflows/offline-gate-in-ci.md)
- [Online A/B in production](../recsys-eval/docs/workflows/online-ab-in-production.md)
- [Interpretation cheat sheet](../recsys-eval/docs/workflows/interpretation-cheat-sheet.md)

## Next steps

- Evidence (what outputs look like): [Evidence](evidence.md)
- ROI and risk model: [ROI and risk model](../start-here/roi-and-risk-model.md)
- Evaluation, pricing, and licensing (buyer guide): [Buyer guide](../pricing/evaluation-and-licensing.md)
