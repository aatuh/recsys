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

## Buyer journey (2 minutes)

Prefer a single shareable page:

- Buyer journey (5-minute path): [Buyer journey](buyer-journey.md)
- Procurement pack (Security/Legal/IT/Finance): [Procurement pack](procurement-pack.md)

1. [What is RecSys?](../start-here/what-is-recsys.md) — outcomes, positioning, and where it fits.
2. [Pilot plan (2–6 weeks)](../start-here/pilot-plan.md) — timeline, deliverables, and exit criteria.
3. [Evidence (what “good outputs” look like)](evidence.md) — examples of logs, reports, and audit records.
4. [Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md) — what to do next and
   how to procure.
5. [Security pack](../security/security-pack.md) — self-serve pack for security/procurement review.

## Proof you can inspect (no calls required)

<div class="grid cards" markdown>

- **[Evidence (example logs + report excerpt)](evidence.md)**  
  What “good outputs” look like (response, exposure log, joined outcomes, report snippet).
- **[Baseline benchmarks](../operations/baseline-benchmarks.md)**  
  Reproducible “anchor numbers” and a template to record your own runs.
- **[Security pack](../security/security-pack.md)**  
  The artifacts a security review expects (posture, logging, and data handling).
- **[Known limitations](../start-here/known-limitations.md)**  
  Boundaries and non-goals (blunt) so you can decide fit quickly.

</div>

## Jump links

<div class="grid cards" markdown>

- **[Buyer journey (5-minute path)](buyer-journey.md)**  
  The recommended “what to read, in order” path you can forward to stakeholders.
- **[Procurement pack](procurement-pack.md)**  
  Role-based checklist (Security/Legal/IT/Finance) with canonical links only.
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
- **[Evidence (what “good outputs” look like)](evidence.md)**  
  Example response, logs, report excerpt, and audit record shape.
- **[Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)**  
  A single page that ties together evaluation, pricing, and licensing.
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

- Data contracts (schemas + examples): [Data contracts](../reference/data-contracts/index.md)
- How evaluation decisions are made: [Run eval and ship](../how-to/run-eval-and-ship.md)

## Read next

- Use cases (what to start with): [Use cases](use-cases.md)
- Success metrics (KPIs + guardrails): [Success metrics](success-metrics.md)
- Evaluation, pricing, and licensing (buyer guide): [Buyer guide](../pricing/evaluation-and-licensing.md)
