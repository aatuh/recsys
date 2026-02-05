---
tags:
  - overview
  - business
---

# Business (buyers and stakeholders)

RecSys is an auditable recommendation system suite with deterministic ranking and versioned ship/rollback.

## Who this is for

- Product owners evaluating RecSys as a product
- Engineering leaders planning a pilot and ownership model
- Security/compliance reviewers doing due diligence

## What you will get

- A 2-minute evaluation path (pilot → metrics → trust → pricing)
- The minimum instrumentation required to measure impact credibly
- Links to the trust center, licensing, and support model

## Start here (2-minute path)

<div class="grid cards" markdown>

- **[Pilot plan (2–6 weeks)](../start-here/pilot-plan.md)**  
  Timeline, deliverables, and exit criteria.
- **[Success metrics](../for-businesses/success-metrics.md)**  
  KPIs + guardrails and what needs to be logged.
- **[Trust center](../security/index.md)**  
  Security, privacy, compliance, and operational readiness.
- **[Pricing](../pricing/index.md)**  
  Commercial evaluation + annual tiers.
- **[Known limitations](../start-here/known-limitations.md)**  
  What the suite does not try to solve (so you can scope honestly).

</div>

## What you need for a credible pilot

RecSys pilots fail for a predictable reason: you can’t measure impact reliably.

Minimum requirements:

- Stable join keys (`request_id` plus a pseudonymous `user_id` or session id)
- Exposure logs (what was shown, with ranks)
- Outcome logs (what the user did)

Start here:

- Data contracts (schemas + examples): [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
- How evaluation decisions are made: [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)

## Read next

- Use cases (where to start): [`for-businesses/use-cases.md`](../for-businesses/use-cases.md)
- ROI and risk model: [`start-here/roi-and-risk-model.md`](../start-here/roi-and-risk-model.md)
- Responsibilities (RACI): [`start-here/responsibilities.md`](../start-here/responsibilities.md)
- Operational reliability & rollback: [`start-here/operational-reliability-and-rollback.md`](../start-here/operational-reliability-and-rollback.md)
- Licensing: [`licensing/index.md`](../licensing/index.md)
