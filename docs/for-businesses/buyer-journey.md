---
tags:
  - overview
  - business
---

# Buyer journey: evaluate RecSys in 5 minutes

## Who this is for

- Product and engineering leaders evaluating RecSys as a product
- Security/procurement stakeholders who want the “what evidence exists?” view first

## What you will get

- A short evaluation path (what to read, in order)
- The minimum requirements for a credible pilot
- Proof artifacts you can inspect (no calls required)
- Links to pricing, licensing, and procurement artifacts

!!! warning "Risk & limitations (read before you invest time)"
    RecSys is optimized for **deterministic, auditable operation**. It is not a managed service and it does not try to
    be an end-to-end “auto-ML recommender platform”.

    Common “surprises” to check early:

    - You run it in your own infrastructure (runbooks and checklists are included).
    - Some setup is still manual today (for example: tenant bootstrap is DB-only).
    - Some connectors are scaffolded (for example: Kafka ingestion is not implemented yet).

    Full list: [Known limitations](../start-here/known-limitations.md) and
    [Capability matrix](../explanation/capability-matrix.md).

## The 5-minute path (recommended order)

1. **Understand what RecSys is**
   - Stakeholder overview: [What is RecSys?](../start-here/what-is-recsys.md)
2. **See the pilot plan and exit criteria**
   - Pilot plan (2–6 weeks): [Pilot plan](../start-here/pilot-plan.md)
   - Define KPIs and guardrails: [Success metrics](success-metrics.md)
3. **Inspect evidence (what outputs look like)**
   - Example response, logs, and report excerpt: [Evidence](evidence.md)
4. **Confirm operational fit**
   - Known limitations (non-goals): [Known limitations](../start-here/known-limitations.md)
   - Rollback story: [Operational reliability & rollback](../start-here/operational-reliability-and-rollback.md)
   - Baseline benchmarks: [Baseline benchmarks](../operations/baseline-benchmarks.md)
5. **Confirm security and data posture**
   - Security pack: [Security pack](../security/security-pack.md)
6. **Decide licensing and procurement path**
   - Buyer guide (evaluation + procurement): [Buyer guide](../pricing/evaluation-and-licensing.md)
   - Licensing decision tree: [Licensing](../licensing/index.md)
   - Pricing overview (commercial plans): [Pricing](../pricing/index.md)

## Outcomes you should expect

By the end of a credible pilot, you should be able to show:

- At least one evaluation report comparing baseline vs candidate (plus a written ship/hold decision)
- An evidence trail you can audit later (exposures + outcomes joined by `request_id`)
- A rollback drill completed (so you trust the lever before you need it)
- A clear “next step” recommendation (ship to staging/production, expand to more surfaces, or stop)

See the canonical checklist: [Buyer guide](../pricing/evaluation-and-licensing.md).

## Pricing and how to buy (when you are ready)

- Pricing overview (commercial plans): [Pricing](../pricing/index.md)
- Licensing obligations (AGPL vs commercial): [Licensing](../licensing/index.md)
- Procurement flow and artifacts: [Buyer guide](../pricing/evaluation-and-licensing.md)
- Order form template: [Order form template](../licensing/order_form.md)

## What you need for a credible pilot (minimum)

RecSys pilots fail for a predictable reason: you can’t measure impact reliably.

Minimum requirements:

- 1 recommendation surface (for example: home feed, PDP similar-items)
- Stable join key: `request_id` present in exposures and outcomes
- Exposure logs (what was shown, with ranks)
- Outcome logs (what the user did), joinable by `request_id`
- Pseudonymous identifiers (avoid raw PII)

Canonical spec:

- Minimum instrumentation spec: [Minimum instrumentation](../reference/minimum-instrumentation.md)

## Proof you can inspect (no calls required)

<div class="grid cards" markdown>

- **[Evidence](evidence.md)**  
  Example response, exposure log, joined outcomes, and report excerpt.
- **[Baseline benchmarks](../operations/baseline-benchmarks.md)**  
  Reproducible “anchor numbers” and a template for your own runs.
- **[Security pack](../security/security-pack.md)**  
  The artifacts a security review expects.
- **[Known limitations](../start-here/known-limitations.md)**  
  Boundaries and non-goals.

</div>

## Read next

- For businesses hub: [For businesses](index.md)
- Procurement pack (Security/Legal/IT/Finance): [Procurement pack](procurement-pack.md)
- Start an evaluation (technical path): [Start an evaluation](../evaluate/index.md)
