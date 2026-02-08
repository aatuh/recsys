---
diataxis: reference
tags:
  - overview
  - business
  - security
---
# Procurement pack (Security, Legal, IT, Finance)
Role-based links to the exact artifacts Security, Legal, IT, and Finance typically review.


## Who this is for

- Procurement and finance stakeholders asking “what do we need to review?”
- Security and IT stakeholders reviewing data posture and operational fit
- Engineering leaders preparing a purchase request

## What you will get

- A skimmable checklist of the exact artifacts to review
- Role-based sections (Security/Legal/IT/Finance) with canonical links only

## Security

- Security pack (canonical): [Security pack](../security/security-pack.md)
- Security posture snapshot (dated one-page summary): [Security posture snapshot](../security/posture-snapshot.md)
- Commercial procurement artifacts (what is published vs Enterprise-custom):
  [Commercial procurement artifacts](../security/commercial-procurement-artifacts.md)
- DPA/SCC baseline (self-serve plans): [DPA and SCC terms](../security/dpa-and-scc.md)
- Subprocessor/distribution disclosure: [Subprocessors and distribution details](../security/subprocessors.md)
- Standard support schedule: [SLA and support schedule](../security/sla-schedule.md)
- Security/privacy/compliance overview: [Security, privacy, compliance](../start-here/security-privacy-compliance.md)
- Known limitations (non-goals): [Known limitations](../start-here/known-limitations.md)


## Privacy / data protection

This is not legal advice. Use this as a practical checklist for your privacy review.

- [ ] Confirm no raw PII is required for the pilot (pseudonymous identifiers are sufficient).
- [ ] Confirm which identifiers you will send (user_id / anonymous_id / session_id) and how they are generated.
- [ ] Define retention for exposure/outcome logs and who can access them.
- [ ] Define deletion/erasure handling (if your org requires it).
- [ ] Confirm data residency requirements (where logs and DB data live).
- [ ] Review default contractual terms (DPA/SCC/subprocessor disclosures) and decide if Enterprise customization is needed.

Canonical overview: [Security, privacy, compliance](../start-here/security-privacy-compliance.md)

## Legal

- Licensing decision tree (canonical): [Licensing](../licensing/index.md)
- Commercial use and how to buy: [Commercial use](../licensing/commercial.md)
- Evaluation license text: [Evaluation license](../licensing/eval_license.md)
- Commercial license text: [Commercial license](../licensing/commercial_license.md)
- Self-serve procurement path: [Self-serve procurement](self-serve-procurement.md)
- Pricing definitions (order form terms): [Pricing definitions](../licensing/pricing.md)
- Order form template: [Order form template](../licensing/order_form.md)

## IT / Operations

- Operations hub: [Operations](../operations/index.md)
- Baseline benchmarks (performance anchors): [Baseline benchmarks](../operations/baseline-benchmarks.md)
- Failure modes and diagnostics: [Failure modes](../operations/failure-modes.md)
- Rollback story (ship/hold/rollback levers): [Operational reliability & rollback](../start-here/operational-reliability-and-rollback.md)
- Deployment guide: [Deploy with Helm](../how-to/deploy-helm.md)

## Finance / Procurement

- Buyer guide (evaluation + procurement flow): [Buyer guide](../pricing/evaluation-and-licensing.md)
- Pricing overview (commercial plans): [Pricing](../pricing/index.md)
- Support model (expectations by plan): [Support](../project/support.md)
- Final cross-functional review template: [Decision readiness matrix](decision-readiness.md)

## Procurement checklist (Definition of Done)

Use the canonical checklist when you want a single shareable DoD list:

- [Procurement checklist](procurement-checklist.md)

## Read next

- Buyer journey (5-minute path): [Buyer journey](buyer-journey.md)
- Start an evaluation (technical path): [Start an evaluation](../evaluate/index.md)
- Buyer guide (evaluation + procurement): [Buyer guide](../pricing/evaluation-and-licensing.md)
- Final go/no-go review: [Decision readiness matrix](decision-readiness.md)
