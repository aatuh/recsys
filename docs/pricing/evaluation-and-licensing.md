---
tags:
  - overview
  - business
  - pricing
  - licensing
---

# Evaluation, pricing, and licensing (buyer guide)

!!! info "Canonical page"
    This page is canonical for the **evaluation and procurement flow** (what to do next, what artifacts to produce).
    For plan definitions and prices, see [pricing](index.md). For file-level license rules, see
    [licensing](../licensing/index.md).

## Who this is for

- Buyers and stakeholders deciding whether to run a pilot and how to procure the right license.

## What you will get

- a recommended evaluation path (what to do first)
- links to canonical pricing definitions and licensing terms
- a practical checklist for procurement/security review

## The evaluation path (recommended)

1. **Run a pilot on one surface**

   - Instrument exposures + outcomes with stable `request_id`
   - Produce at least one report that compares baseline vs candidate

   Start here:

   - Pilot plan (2–6 weeks): [Pilot plan](../start-here/pilot-plan.md)
   - Success metrics (KPIs + guardrails): [Success metrics](../for-businesses/success-metrics.md)
   - Evidence (what outputs look like): [Evidence](../for-businesses/evidence.md)

2. **Confirm operational fit**

   - Known limitations and non-goals: [Known limitations](../start-here/known-limitations.md)
   - Security, privacy, compliance:
     [Security, privacy, compliance](../start-here/security-privacy-compliance.md)
   - Rollback story:
     [Operational reliability & rollback](../start-here/operational-reliability-and-rollback.md)

3. **Decide your purchase scope**

   Typical scope questions:

   - How many tenants and deployments do we need?
   - Do we require commercial terms for production use?
   - What support expectations do we need (response time, channels, escalation)?

## Do you need commercial terms? (quick decision tree)

This is not legal advice. If licensing affects your business, involve counsel.

1. Are you only using `recsys-eval/**`?
   - Yes → it is Apache-2.0 (commercial license not needed; comply with Apache-2.0 conditions).
   - No → continue.
2. Are you deploying or offering network access to the serving stack (`api/**`, `recsys-algo/**`, `recsys-pipelines/**`)?
   - Yes → AGPLv3 applies unless you have commercial terms.
   - No → your obligations depend on how you use/distribute the code.
3. Can you comply with AGPLv3 terms for your deployment (including source-offer obligations for modifications)?
   - Yes → use under AGPLv3.
   - No / we need modifications private → request a commercial license.

See the canonical decision tree and file-level rules: [Licensing](../licensing/index.md).

## Plan mapping (quick)

- **Commercial Evaluation (30 days)**: a time-boxed pilot under commercial terms (recommended if you want to avoid AGPL
  uncertainty during the pilot).
- **Starter**: one production deployment for one tenant, plus a small number of surfaces.
- **Growth**: a few tenants and/or deployments, with faster async response expectations.
- **Enterprise**: multi-region HA, OEM/resale, regulated environments, or custom terms.

Canonical plan definitions and prices: [Pricing](index.md).

## How to procure (fast path)

1. Run a pilot on one surface and produce at least one report:
   [Run eval and ship](../how-to/run-eval-and-ship.md)
2. Confirm fit for your environment:
   [Known limitations](../start-here/known-limitations.md) and [Security pack](../security/security-pack.md)
3. Choose the plan and scope (tenants/deployments/support expectations): [Pricing](index.md)
4. Use the order form template and contact us:
   [Order form template](../licensing/order_form.md) and [`contact@recsys.app`][pricing_contact]

## Outputs and exit criteria

By the end of a successful evaluation, you should have:

- At least one evaluation report comparing baseline vs candidate:
  [Run eval and ship](../how-to/run-eval-and-ship.md)
- A default evaluation pack agreed (primary KPI + guardrails):
  [Default evaluation pack](../recsys-eval/docs/default-evaluation-pack.md)
- An evidence trail you can audit later (exposures + outcomes joined by `request_id`):
  [Evidence](../for-businesses/evidence.md)
- A written decision (ship/hold/rollback) with the supporting report links
- One rollback drill completed (so you trust the lever before you need it):
  [Operational reliability & rollback](../start-here/operational-reliability-and-rollback.md)

## Who does what (typical)

- Product owner: chooses KPIs/guardrails and owns the final ship/hold decision
- Lead developer: integrates one surface end-to-end and owns the rollout plan
- Data/analytics: validates join-rate and reads the evaluation reports
- Security: reviews the security pack and data posture

See also:

- Responsibilities (RACI): [Responsibilities](../start-here/responsibilities.md)
- Customer onboarding checklist: [Customer onboarding checklist](../start-here/customer-onboarding-checklist.md)

## Timeline vs evaluation license term

- Pilot plans commonly take **2–6 weeks**: [Pilot plan](../start-here/pilot-plan.md)
- The commercial evaluation license term is **30 days** (from first access), unless extended in writing:
  [Evaluation license](../licensing/eval_license.md)

If your pilot scope exceeds 30 days:

- Request an evaluation term extension in writing **before** the term expires, and/or
- Start procurement earlier so licensing does not block measurement.

## Pricing (canonical)

For current plan definitions and pricing, see:

- Pricing overview (commercial plans): [Pricing](index.md)
- Legal pricing definitions: [Pricing definitions](../licensing/pricing.md)

## Licensing (plain language)

This repository is multi-license:

- `recsys-eval/**` is Apache-2.0 (permissive).
- The serving stack (`api/**`, `recsys-algo/**`, `recsys-pipelines/**`) is AGPLv3 unless you purchase commercial terms.

For the canonical decision tree and file-level rules, see:

- Licensing overview: [Licensing](../licensing/index.md)
- Commercial use & how to buy: [Commercial use](../licensing/commercial.md)

## What to hand to procurement/security

- Security pack: [Security pack](../security/security-pack.md)
- Known limitations: [Known limitations](../start-here/known-limitations.md)
- Support model (includes expectations by plan): [Support](../project/support.md)
- Order form template: [Order form template](../licensing/order_form.md)

## Procurement checklist (Definition of Done)

- [ ] We ran a pilot and produced at least one evaluation report.
- [ ] Our pilot timeline fits the chosen license term (or we secured an extension in writing).
- [ ] We reviewed known limitations and confirmed fit for our current stage.
- [ ] We chose plan scope (tenants/deployments) and support expectations.
- [ ] We confirmed license obligations (AGPL vs commercial terms).
- [ ] Security reviewed the security pack.

## Read next

- Pricing overview (commercial plans): [Pricing](index.md)
- Licensing decision tree: [Licensing](../licensing/index.md)
- Security pack: [Security pack](../security/security-pack.md)

[pricing_contact]: mailto:contact@recsys.app?subject=RecSys%20Commercial%20Evaluation
