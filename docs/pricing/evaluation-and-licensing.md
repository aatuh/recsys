---
diataxis: how-to
tags:
  - overview
  - business
  - pricing
  - licensing
---
# Evaluation, pricing, and licensing (buyer guide)

Use this page to run a credible evaluation and procure the right license **without meetings**.

!!! info "Canonical page"
    This page is canonical for the **evaluation and procurement flow** (what to do next, what artifacts to produce).
    For plan definitions and prices, see [Pricing](index.md). For file-level license rules, see
    [Licensing](../licensing/index.md).

## Who this is for

- Buyers and stakeholders deciding whether to run a pilot and how to procure the right license.

## What you will get

- A step-by-step evaluation path (what to do first)
- The shortest path from **pilot → decision → procurement**
- A procurement checklist you can hand to Security/Legal/IT/Finance

## Step 0 — Decide the evaluation owner and success bar

Before you run anything, agree on:

- **One surface** to pilot (start small)
- **One primary KPI** and 2–5 guardrails
- **Who owns the decision** (ship / hold / rollback)

Start here:

- Pilot plan (2–6 weeks): [Pilot plan](../start-here/pilot-plan.md)
- Ownership (RACI): [Responsibilities](../start-here/responsibilities.md)
- KPIs and guardrails: [Success metrics](../for-businesses/success-metrics.md)

## Step 1 — Run a pilot that produces auditable evidence

A credible pilot proves the measurement loop is real:

- You can serve non-empty recommendations
- You log what was shown (**exposures**) and what happened (**outcomes**)
- You can join them reliably by a stable `request_id`
- You can produce a report that supports a decision

Start here:

- Tutorial (end-to-end): [Local end-to-end](../tutorials/local-end-to-end.md)
- Suite workflow: [Run eval and ship](../how-to/run-eval-and-ship.md)
- What “good outputs” look like: [Evidence](../for-businesses/evidence.md)

## Step 2 — Confirm operational fit for your environment

Confirm the suite fits your constraints **before** procurement:

- Boundaries and non-goals: [Known limitations](../start-here/known-limitations.md)
- Data handling and posture: [Security, privacy, compliance](../start-here/security-privacy-compliance.md)
- Rollback story: [Operational reliability & rollback](../start-here/operational-reliability-and-rollback.md)

## Step 3 — Choose plan scope and license path

### Plan mapping (quick)

- **Commercial Evaluation (30 days)**: a time-boxed pilot under commercial terms (recommended if you want to avoid AGPL
  uncertainty during the pilot).
- **Starter**: one production deployment for one tenant, plus a small number of surfaces.
- **Growth**: a few tenants and/or deployments, with faster async response expectations.
- **Enterprise**: multi-region HA, OEM/resale, regulated environments, or custom terms.

Canonical plan definitions and prices: [Pricing](index.md).

### Do you need commercial terms? (quick decision tree)

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

## How to procure (fast path)

1. Run a pilot on one surface and produce at least one report:
   [Run eval and ship](../how-to/run-eval-and-ship.md)
2. Confirm fit for your environment:
   [Known limitations](../start-here/known-limitations.md) and [Security pack](../security/security-pack.md)
3. Choose the plan and scope (tenants/deployments/support expectations):
   [Pricing](index.md)
4. Use the self-serve procurement path for standard plans:
   [Self-serve procurement](../for-businesses/self-serve-procurement.md)
5. Use the order form + contact path only for Enterprise/custom terms:
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

## Procurement checklist

Hand this to Security/Legal/IT/Finance as a Definition of Done:

- [Procurement checklist](../for-businesses/procurement-checklist.md)
- [Decision readiness matrix](../for-businesses/decision-readiness.md)

[pricing_contact]: mailto:contact@recsys.app?subject=RecSys%20Commercial%20Evaluation

## Read next

- Pricing overview (commercial plans): [Pricing](index.md)
- Self-serve path: [Self-serve procurement](../for-businesses/self-serve-procurement.md)
- Licensing decision tree: [Licensing](../licensing/index.md)
- What “good outputs” look like: [Evidence](../for-businesses/evidence.md)
- Security pack: [Security pack](../security/security-pack.md)
