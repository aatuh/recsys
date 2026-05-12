---
diataxis: how-to
tags:
  - evaluation
  - business
  - procurement
---
# Guided evaluation

Use this path after you have run the local proof kit and want guided help validating RecSys in your own environment.

!!! info "Current evaluation posture"
    RecSys does not currently offer a hosted evaluation sandbox. The supported path is:
    **local proof kit first**, then a guided pilot in the buyer's customer-controlled environment.

## When to use this

- You have produced the commercial proof-kit outputs locally.
- You want to pilot one real recommendation surface with your own events, catalog, and guardrails.
- You need help reviewing integration scope, evaluation evidence, security constraints, or production-readiness gaps.

If you already know the production scope and only need commercial terms, choose Starter or Growth directly:
[Evaluation, pricing, and licensing](../pricing/evaluation-and-licensing.md#evaluation-or-production-direct).

## Path

1. Run the local proof kit:
   [Commercial proof kit](../tutorials/commercial-proof-kit.md)
2. Choose one customer-environment pilot surface.
3. Send the guided-evaluation intake below.
4. Run the pilot with your own exposure/outcome logs and success metrics.
5. Produce a report and decision note.
6. Choose stop, extend, Starter, Growth, or an Enterprise discovery path.

## Guided-evaluation intake

Send one concise request with:

- Surface name and user journey, for example `home feed` or `PDP similar-items`
- Tenant and deployment target, including non-prod or production-like environment
- Expected event types, approximate volume, and pilot time window
- Links or paths to proof-kit outputs: recommendation response, manifest, eval report, and decision note
- Primary success metric and 2-5 guardrails
- Rollback or safety guardrails that must hold before serving traffic
- Data constraints, including PII, retention, residency, and pseudonymous identifier rules
- Security/procurement constraints, including review deadlines or required artifacts
- Selected plan, bundle, or service package if already known

## What guided evaluation includes

- Review of the planned surface, tenant/deployment scope, and instrumentation assumptions
- Review of proof-kit output and pilot evidence shape
- Guidance on using published docs, runbooks, and procurement artifacts
- Written recommendations for next steps: stop, continue pilot, buy Starter/Growth, or enter Enterprise discovery

## What it does not include

- Hosted sandbox access
- Managed hosting or production operations
- Unlimited custom development
- KPI guarantees
- Custom legal, security, SLA, or HA commitments unless captured in an Enterprise Order Form

## Exit criteria

By the end of a guided evaluation, you should have:

- One customer-environment surface with joinable exposure/outcome evidence
- A report tied to the agreed success metric and guardrails
- A written ship/hold/rollback decision
- A procurement path: no purchase, Commercial Evaluation extension, Starter, Growth, or Enterprise discovery

## Read next

- Start an evaluation: [Start an evaluation](index.md)
- Buyer guide: [Evaluation, pricing, and licensing](../pricing/evaluation-and-licensing.md)
- Self-serve procurement: [Self-serve procurement](../for-businesses/self-serve-procurement.md)
- Enterprise readiness evidence: [Enterprise readiness evidence](../for-businesses/enterprise-readiness-evidence.md)
