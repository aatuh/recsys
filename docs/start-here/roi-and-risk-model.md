---
tags:
  - overview
  - business
  - evaluation
---

# ROI and risk model

## Who this is for

- Stakeholders evaluating whether a RecSys pilot is “worth doing”
- Product and analytics teams who need a simple measurement plan
- Engineering leads who want to de-risk ownership and rollout

## What you will get

- A lightweight ROI template you can adapt to your domain
- A concrete “what to measure” checklist (with links to the right docs)
- A risk checklist with mitigations and escalation cues

## ROI (template, not a promise)

Recommendations only create value if they move a **business KPI** while keeping **guardrails** healthy.

Start with one primary KPI per surface:

- ecommerce: conversion rate, revenue per session, add-to-cart rate
- content: time spent, return rate, completion rate

Then define 2–4 guardrails:

- latency / error rate
- empty-recs rate
- user complaints or negative feedback signals
- diversity / coverage constraints (if applicable)

A simple ROI framing:

- **Incremental value** = (KPI lift) × (eligible traffic) × (value per action)
- **Cost** = engineering time + operational load + infrastructure

Your pilot goal is to decide whether the incremental value is large enough to justify a production rollout.

## Measurement plan (what we need from you)

To measure lift reliably, you need consistent logging and joins:

- Exposure logs: what was shown (with ranks)
- Outcome logs: what the user did
- Stable join IDs: `request_id` and a pseudonymous `user_id` or session id

Start here:

- Data contracts (canonical schemas + examples): [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
- Pilot plan (deliverables + exit criteria): [`start-here/pilot-plan.md`](pilot-plan.md)
- How to run evaluation and decide ship/hold/rollback:
  [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)

## Risks and mitigations (practical)

- **Bad instrumentation (joins low, SRM warnings, “impossible” lift)**
  - Mitigation: validate schemas early; fix logging before trusting metrics.
  - Docs: [`recsys-eval/docs/troubleshooting.md`](../recsys-eval/docs/troubleshooting.md)
- **Operational risk (bad publish affects users)**
  - Mitigation: use reversible rollouts; practice rollback once.
  - Docs: [`start-here/operational-reliability-and-rollback.md`](operational-reliability-and-rollback.md)
- **Data quality drift (late data, spikes, schema surprises)**
  - Mitigation: validation gates + guardrails; alert on freshness.
  - Docs: [`recsys-pipelines/docs/operations/slos-and-freshness.md`](../recsys-pipelines/docs/operations/slos-and-freshness.md)
- **Privacy / compliance risk**
  - Mitigation: log only pseudonymous IDs; treat schemas as strict; minimize PII.
  - Docs: [`start-here/security-privacy-compliance.md`](security-privacy-compliance.md)

## Read next

- Pilot plan: [`start-here/pilot-plan.md`](pilot-plan.md)
- Stakeholder overview: [`start-here/what-is-recsys.md`](what-is-recsys.md)
- Security, privacy, compliance: [`start-here/security-privacy-compliance.md`](security-privacy-compliance.md)
- Interpreting eval results: [`recsys-eval/docs/interpreting_results.md`](../recsys-eval/docs/interpreting_results.md)
