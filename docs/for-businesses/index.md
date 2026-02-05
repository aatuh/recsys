---
tags:
  - overview
  - business
---

# For businesses (buyers and stakeholders)

## Who this is for

- Product owners evaluating RecSys as a product
- Engineering leaders planning a pilot and ownership model
- Stakeholders who want the “value / risk / cost / timeline” view first

## What you will get

- A short, business-first evaluation path (no technical rabbit holes)
- The minimum requirements to run a credible pilot (data + people)
- The trust story: security, operations, rollback, and how decisions are audited

## Start here (2-minute path)

<div class="grid cards" markdown>

- **[What is RecSys?](../start-here/what-is-recsys.md)**  
  What you get, where it fits, and what outcomes to expect.
- **[Pilot plan (2–6 weeks)](../start-here/pilot-plan.md)**  
  Timeline, deliverables, and exit criteria (includes a fast 2–4 week path).
- **[ROI and risk model](../start-here/roi-and-risk-model.md)**  
  A simple template for lift measurement, guardrails, and ownership boundaries.
- **[Operational reliability & rollback](../start-here/operational-reliability-and-rollback.md)**  
  What can be rolled back, how fast, and what “healthy” means in production.
- **[Security, privacy, compliance](../start-here/security-privacy-compliance.md)**  
  What we log/store and how to run this safely.
- **[Licensing & pricing](../licensing/index.md)**  
  Where the licenses apply and how to think about commercial use.

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

- Use cases (what to start with): [`for-businesses/use-cases.md`](use-cases.md)
- Success metrics (KPIs + guardrails): [`for-businesses/success-metrics.md`](success-metrics.md)
- Evidence (what “good outputs” look like): [`for-businesses/evidence.md`](evidence.md)
