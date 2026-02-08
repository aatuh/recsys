---
diataxis: how-to
tags:
  - business
  - procurement
  - decision
  - licensing
---
# Decision readiness matrix

Use this page to run a final go/no-go review before requesting a commercial order form.

## Who this is for

- Business owners accountable for the purchase decision
- Security, legal, and finance reviewers who need a single source of status

## What you will get

- A single matrix of procurement-critical areas
- Clear separation between what is documented vs what still needs vendor confirmation
- Suggested owners for each decision area

## Go / no-go matrix

Mark each area with `Ready`, `Needs clarification`, or `Blocked`.

| Area | Decision question | Documented in docs | Requires vendor confirmation | Suggested owner | Status |
| --- | --- | --- | --- | --- | --- |
| Product fit | Does one scoped surface show measurable pilot value? | [Evidence](evidence.md), [Success metrics](success-metrics.md), [Pilot plan](../start-here/pilot-plan.md) | No | Product + Eng lead | ☐ Ready ☐ Needs clarification ☐ Blocked |
| Licensing path | Is AGPL vs commercial path selected with rationale? | [Licensing](../licensing/index.md), [Buyer guide](../pricing/evaluation-and-licensing.md) | Sometimes (edge legal interpretation) | Legal | ☐ Ready ☐ Needs clarification ☐ Blocked |
| Commercial scope | Are tenants/deployments/support scope selected? | [Pricing](../pricing/index.md), [Pricing definitions](../licensing/pricing.md) | No for self-serve plans; yes for Enterprise/custom scope | Buyer + Finance | ☐ Ready ☐ Needs clarification ☐ Blocked |
| Security posture | Does baseline posture meet internal policy for pilot/production? | [Security pack](../security/security-pack.md), [Security posture snapshot](../security/posture-snapshot.md) | Sometimes (org-specific controls) | Security | ☐ Ready ☐ Needs clarification ☐ Blocked |
| Privacy/data protection | Are PII posture, retention, and deletion expectations agreed? | [Security, privacy, compliance](../start-here/security-privacy-compliance.md), [DPA and SCC terms](../security/dpa-and-scc.md), [Subprocessors and distribution details](../security/subprocessors.md) | No for self-serve plans; yes for Enterprise/custom annexes | Privacy + Legal | ☐ Ready ☐ Needs clarification ☐ Blocked |
| Support expectations | Do response expectations match on-call risk tolerance? | [Support](../project/support.md), [SLA and support schedule](../security/sla-schedule.md), [Pricing](../pricing/index.md) | No for self-serve plans; yes for custom SLA/escalation terms | Eng manager + SRE | ☐ Ready ☐ Needs clarification ☐ Blocked |
| Legal terms | Are liability cap, venue, term, and special terms accepted? | [Commercial license](../licensing/commercial_license.md), [Order form template](../licensing/order_form.md), [Self-serve procurement](self-serve-procurement.md) | No for self-serve plans; yes for Enterprise/custom redlines | Legal + Procurement | ☐ Ready ☐ Needs clarification ☐ Blocked |
| Procurement artifacts | Is the internal procurement package complete and shareable? | [Procurement checklist](procurement-checklist.md), [Commercial procurement artifacts](../security/commercial-procurement-artifacts.md) | No (if all required artifacts are present) | Procurement | ☐ Ready ☐ Needs clarification ☐ Blocked |

## Exit criteria

You are purchase-ready when all rows are `Ready`, and any vendor-confirmed points are explicitly recorded in the
procurement thread.

## Read next

- Procurement checklist: [Procurement checklist](procurement-checklist.md)
- Self-serve path: [Self-serve procurement](self-serve-procurement.md)
- Buyer guide: [Evaluation, pricing, and licensing](../pricing/evaluation-and-licensing.md)
- Order form template: [Order form template](../licensing/order_form.md)
