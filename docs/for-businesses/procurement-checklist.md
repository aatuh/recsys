---
diataxis: how-to
tags:
  - business
  - procurement
  - security
  - licensing
---
# Procurement checklist

Use this checklist to move from **completed pilot** → **approved procurement** with minimal back-and-forth.

!!! info "Intent"
    This page is a Definition of Done for Security / Legal / IT / Finance. It links only to canonical pages in this docs site.

## When to use this

- You have run a pilot (or a time-boxed evaluation) and now want to procure a commercial evaluation or production license.
- You want a shareable list of required artifacts and review steps.

## Checklist (Definition of Done)

### 1) Evaluation evidence (product + analytics)

- [ ] We can show a non-empty recommendation response (one surface).
- [ ] We have exposures and outcomes joined by stable `request_id`.
- [ ] We produced at least one evaluation report comparing baseline vs candidate.
- [ ] We recorded a written ship/hold/rollback decision with links to the artifacts.

Links:

- Evidence (what “good outputs” look like): [Evidence](evidence.md)
- Suite workflow (report → decision): [Run eval and ship](../how-to/run-eval-and-ship.md)
- Data contracts (schemas + join logic): [Data contracts](../reference/data-contracts/index.md)

### 2) Security and privacy review

- [ ] We reviewed what data is logged/stored, retention expectations, and access control boundaries.
- [ ] We reviewed operational hardening expectations (auth, tenancy, auditability).

Links:

- Security pack (canonical): [Security pack](../security/security-pack.md)
- Security, privacy, compliance (overview): [Security, privacy, compliance](../start-here/security-privacy-compliance.md)

### 3) Operational fit (SRE / on-call)

- [ ] We reviewed known limitations and confirmed they fit our current stage.
- [ ] We reviewed rollback and failure-mode runbooks.
- [ ] We have a minimal production readiness plan (even if we start with a pilot deployment).

Links:

- Known limitations: [Known limitations](../start-here/known-limitations.md)
- Operational reliability & rollback: [Operational reliability & rollback](../start-here/operational-reliability-and-rollback.md)
- Production readiness checklist: [Production readiness checklist](../operations/production-readiness-checklist.md)

### 4) Licensing and purchasing decision

- [ ] We chose **AGPL vs commercial** path and documented why.
- [ ] We chose plan scope (tenants, deployments) and support expectations.
- [ ] For self-serve plans, we use published legal/security defaults; for Enterprise, negotiated terms are captured in the Order Form.

Links:

- Pricing (canonical): [Pricing](../pricing/index.md)
- Buyer guide (recommended flow): [Evaluation, pricing, and licensing](../pricing/evaluation-and-licensing.md)
- Self-serve procurement flow: [Self-serve procurement](self-serve-procurement.md)
- Licensing decision tree: [Licensing](../licensing/index.md)
- Commercial use (what you get + how to buy): [Commercial Use & Licensing](../licensing/commercial.md)
- DPA/SCC baseline: [DPA and SCC terms](../security/dpa-and-scc.md)
- Subprocessor/distribution disclosure: [Subprocessors and distribution details](../security/subprocessors.md)
- Standard support schedule: [SLA and support schedule](../security/sla-schedule.md)
- Order form template: [Order form template](../licensing/order_form.md)

## What to send in one email (suggested bundle)

Copy/paste this into your procurement thread:

- Pilot summary (surface, KPI, window)
- Links to:
  - evaluation report
  - evidence kit / logs sample
  - security pack
  - known limitations
  - selected plan and scope
  - order form draft

If you want a ready-made internal bundle format, start from the template in:

- Evidence kit template: [Evidence](evidence.md)

## Read next

- Buyer guide: [Evaluation, pricing, and licensing](../pricing/evaluation-and-licensing.md)
- Procurement pack (role-based links): [Procurement pack](procurement-pack.md)
- Pricing (canonical): [Pricing](../pricing/index.md)
- Security pack: [Security pack](../security/security-pack.md)
- Final go/no-go review: [Decision readiness matrix](decision-readiness.md)
