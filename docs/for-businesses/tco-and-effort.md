---
diataxis: explanation
tags:
  - business
  - tco
  - planning
---
# TCO and effort model
This page explains TCO and effort model and how it fits into the RecSys suite.


## Who this is for

- Buyers and stakeholders planning time, roles, and risk.
- Lead developers estimating integration effort.

## What you will get

- A decomposition of effort into concrete work packages
- What changes effort up/down (so you can plan realistically)
- A checklist you can use in procurement and planning

## The effort buckets

### 1) Integration effort

What you do:

- Implement `POST /v1/recommend` integration
- Send tenant scope and request IDs correctly
- Emit exposure logs (minimum viable spec)

Docs:

- Integration tutorial: [How-to: integrate recsys-service into an application](../how-to/integrate-recsys-service.md)
- Minimum instrumentation spec: [Minimum instrumentation spec (for credible evaluation)](../reference/minimum-instrumentation.md)

Effort drivers:

- Number of surfaces and item types
- How clean your existing event/identity model is
- Whether you already log “what was shown” consistently

### 2) Data readiness effort

What you do:

- Provide item metadata (IDs, tags, price, availability, …)
- Provide outcome signals (click, add-to-cart, purchase) if available

Docs:

- Data contracts: [Data contracts](../reference/data-contracts/index.md)

Effort drivers:

- Are event schemas consistent across clients/platforms?
- Can you produce join keys reliably? (request IDs, user/session IDs)

### 3) Ops readiness effort

What you do:

- Define deployment shape (pilot vs production-like)
- Decide DB-only vs artifact/manifest mode
- Put runbooks and rollback in place

Docs:

- Pilot deployment options: [Pilot deployment options](../start-here/pilot-deployment-options.md)
- Ops checklist: [Production readiness checklist (RecSys suite)](../operations/production-readiness-checklist.md)

Effort drivers:

- Existing platform maturity (observability, deployment, on-call)
- Need for strict SLOs and rollback windows

### 4) Optimization effort (ongoing)

What you do:

- Tune weights, rules, and constraints per tenant/surface
- Add or improve signals via pipelines (optional)
- Evaluate changes and decide ship/hold/rollback

Docs:

- Customization map: [Customization map](../explanation/customization-map.md)
- Evaluation workflow: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)

Effort drivers:

- How many stakeholders need to sign off on changes
- How fast you need to iterate vs how risk-averse you are

## A simple planning template

Use this as a procurement-ready checklist:

- [ ] Surfaces in scope (list them)
- [ ] Item ID and item metadata source confirmed
- [ ] Exposure logging confirmed (format, storage, access)
- [ ] Outcome logging confirmed (optional but recommended)
- [ ] Tenant model confirmed (one tenant vs many)
- [ ] Data mode chosen (DB-only vs artifact/manifest)
- [ ] Pilot environment defined (local / staging / prod-like)
- [ ] Success metrics agreed (business + offline/online)

## Read next

- Pilot plan: [Pilot plan (2–6 weeks)](../start-here/pilot-plan.md)
- Success metrics: [Success metrics (KPIs, guardrails, and exit criteria)](success-metrics.md)
- Evaluation and licensing: [Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)
