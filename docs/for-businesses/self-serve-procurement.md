---
diataxis: how-to
tags:
  - business
  - procurement
  - pricing
  - licensing
---
# Self-serve procurement (minimum requests)

Use this page when you want the shortest path from evaluation to commercial activation.

## Goal

Minimize procurement back-and-forth to a single commercial intake request for standard plans.

## Scope

This flow is designed for:

- Commercial Evaluation
- Starter
- Growth
- Fixed-scope service packages listed on [Pricing](../pricing/index.md#fixed-scope-services-and-add-ons)
- Private first-year bundle examples captured in an Order Form:
  [Private first-year bundles](../pricing/index.md#private-first-year-bundles-order-form-examples)

Enterprise is intentionally custom and may require negotiation.

## One-request flow

1. Choose plan and scope from canonical pricing:
   [Pricing](../pricing/index.md)
2. Choose any private first-year bundle or fixed-scope service package that reduces a concrete adoption risk:
   [Private first-year bundles](../pricing/index.md#private-first-year-bundles-order-form-examples) and
   [Fixed-scope services and add-ons](../pricing/index.md#fixed-scope-services-and-add-ons)
3. Validate legal/security artifacts from published docs:
   [Commercial procurement artifacts](../security/commercial-procurement-artifacts.md)
4. Fill the order form template using standard defaults:
   [Order form template](../licensing/order_form.md)
5. Send a single procurement request with completed order form, selected plan, selected bundle if any, and selected
   service package(s).

## Required artifact bundle (all published)

- Licensing decision tree: [Licensing](../licensing/index.md)
- Commercial agreement: [Commercial license](../licensing/commercial_license.md)
- Evaluation deliverables pack:
  [Evaluation deliverables pack](evaluation-deliverables-pack.md)
- DPA/SCC baseline: [DPA and SCC terms](../security/dpa-and-scc.md)
- Subprocessor/distribution disclosure:
  [Subprocessors and distribution details](../security/subprocessors.md)
- Standard support schedule: [SLA and support schedule](../security/sla-schedule.md)
- Security pack: [Security pack](../security/security-pack.md)

## What remains non-self-serve

Enterprise negotiation only, including:

- OEM/resale or third-party hosting rights
- Multi-region HA and regulated-environment commitments
- Custom SLA/service credits
- Custom DPA/SCC/legal riders
- Custom liability cap or other redlines

Enterprise buyers should review the evidence map before relying on custom operational, support, HA, or legal/security
commitments: [Enterprise readiness evidence](enterprise-readiness-evidence.md).

## Guided-evaluation intake

Use this when you want guided help after the local proof kit. RecSys does not currently offer a hosted evaluation
sandbox; the supported path is local proof kit first, then a guided customer-environment pilot.

Include:

- Surface name and user journey
- Tenant and deployment target
- Expected event types, approximate volume, and pilot time window
- Links or paths to proof-kit output: recommendation response, manifest, eval report, and decision note
- Primary success metric and 2-5 guardrails
- Rollback or safety constraints
- Security/procurement constraints and required review deadlines

Canonical intake page: [Guided evaluation](../evaluate/guided-evaluation.md).

## What to include for first-year bundles

If you want a private first-year bundle, include:

- Bundle name from [Pricing](../pricing/index.md#private-first-year-bundles-order-form-examples)
- Selected plan scope: tenants, deployments, production recommendation surfaces, and support expectations
- Fixed-scope service package milestone and review inputs
- Confirmation that the bundle is first-year Order Form packaging and does not change renewal pricing unless stated

## What to include for service packages

For each selected fixed-scope service package, include:

- The package name from [Pricing](../pricing/index.md#fixed-scope-services-and-add-ons)
- The tenant, deployment, and surface in scope
- The milestone to review (pilot readiness, production readiness, or security/procurement review)
- Links to the relevant docs, configs, reports, or runbooks you want reviewed

These packages are advisory/review packages. They do not include managed hosting, production on-call, or unlimited
custom development unless a custom Enterprise Order Form says otherwise.

## Read next

- Buyer guide (end-to-end): [Evaluation, pricing, and licensing](../pricing/evaluation-and-licensing.md)
- Evaluation deliverables pack: [Evaluation deliverables pack](evaluation-deliverables-pack.md)
- Guided evaluation: [Guided evaluation](../evaluate/guided-evaluation.md)
- Procurement checklist: [Procurement checklist](procurement-checklist.md)
- Enterprise readiness evidence: [Enterprise readiness evidence](enterprise-readiness-evidence.md)
- Decision readiness matrix: [Decision readiness matrix](decision-readiness.md)
