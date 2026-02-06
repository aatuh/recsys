---
tags:
  - overview
  - business
  - pricing
  - licensing
---

# Evaluation, pricing, and licensing (buyer guide)

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

## Next steps

- Pricing overview (commercial plans): [Pricing](index.md)
- Licensing decision tree: [Licensing](../licensing/index.md)
- Security pack: [Security pack](../security/security-pack.md)
