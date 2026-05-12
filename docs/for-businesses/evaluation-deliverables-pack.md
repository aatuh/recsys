---
diataxis: reference
tags:
  - business
  - evaluation
  - procurement
  - pricing
---
# Evaluation deliverables pack

This page shows what a buyer should expect after a commercial evaluation or a fixed-scope review package.

!!! info "Synthetic examples"
    These examples use the synthetic `examples/data/ecommerce-mini` fixture. They are not customer data, not a
    production benchmark, and not a promise of KPI lift. Use them as the target shape for a real pilot deliverables
    bundle.

## What you receive

| Deliverable | When it appears | Purpose |
| --- | --- | --- |
| Pilot evaluation report | Commercial Evaluation or guided pilot | Show served recommendations, joinable logs, offline metrics, and a decision trail. |
| Pilot Integration Review memo | Pilot Integration Review | Review one surface's integration, instrumentation, and evaluation readiness. |
| Production Readiness Package memo | Production Readiness Package | Review deployment, rollback, observability, hardening, and launch blockers. |
| Security / Procurement Review packet | Security / Procurement Review Package | Help Security, Legal, IT, and Finance find the right published artifacts and open questions. |
| Decision note | Any evaluation or review | Record ship/hold/rollback rationale and next commercial step. |

Service-package deliverables are advisory review outputs. They are not managed operations, unlimited custom development,
SLA commitments, or guaranteed KPI lift.

## Completed example: pilot evaluation report

### Pilot report context

- Customer: `Example Commerce Oy`
- Tenant: `demo`
- Surface: `home`
- Dataset: `examples/data/ecommerce-mini`
- Data handling: synthetic, non-PII fixture
- Serving mode: artifact-manifest mode with popularity fallback
- Report ID: `<offline-run-id>`
- Created at: `<report-created-at>`

### Pilot report artifacts

- Served recommendation response: `recommendation-response.json`
- Served exposure log: `served-exposures.eval.jsonl`
- Published manifest: `manifest.json`
- Offline evaluation report: `eval/offline-report.json` and `eval/offline-report.md`
- Decision note: `decision-note.md`

### Pilot report excerpt

Stable excerpt from the commercial proof kit:

```text
Cases Evaluated: 6

Executive Summary
- Decision: not_configured
- Highlights:
  - ndcg@3: 0.938488
  - precision@3: 0.388889
  - recall@3: 1.000000

Offline Metrics
precision@3: 0.388889
recall@3: 1.000000
ndcg@3: 0.938488
hitrate@3: 1.000000
coverage@3: 1.000000
```

### Pilot report interpretation

- The serving API returned non-empty recommendations for the `home` surface.
- Exposure and outcome fixtures are joinable by `request_id`.
- The offline report proves the evaluation loop runs end to end on synthetic data.
- The synthetic report does not prove real product lift. Replace the fixture with one real pilot surface before making a
  production decision.

## Completed example: Pilot Integration Review memo

### Integration review summary

- Package: Pilot Integration Review
- Scope: one tenant, one deployment target, one `home` recommendation surface
- Reviewed inputs: request/response sample, exposure/outcome fixture, tenant/surface naming, proof-kit report
- Review outcome: viable for a scoped customer-environment pilot after instrumentation checks pass

### Integration review findings

| Area | Finding | Status |
| --- | --- | --- |
| Tenant and surface scope | `demo` / `home` maps cleanly to one pilot surface. | Ready |
| Serving request | `POST /v1/recommend` returns ranked items with `request_id` metadata. | Ready |
| Exposure logging | Proof-kit exposure output exists and is eval-compatible. | Ready |
| Outcome logging | Synthetic outcomes are joinable; customer event mapping still needs confirmation. | Needs customer input |
| Evaluation report | Offline report is generated and shareable. | Ready |
| Rollback ownership | Owner and rollback drill need to be named for the real pilot. | Open |

### Integration review recommendation

Proceed to a guided customer-environment pilot if the buyer can provide:

- one production-like event source for the selected surface
- stable `request_id` on exposures and outcomes
- a primary KPI and 2-5 guardrails
- a named owner for rollout, hold, and rollback decisions

## Completed example: Production Readiness Package memo

### Readiness review summary

- Package: Production Readiness Package
- Scope: one Starter deployment path for `Example Commerce Oy`
- Reviewed inputs: Helm deployment guide, production readiness checklist, rollback runbooks, proof-kit artifacts
- Review outcome: not yet production-ready; ready for a production-readiness workback plan

### Readiness findings

| Area | Finding | Required before launch |
| --- | --- | --- |
| Deployment | Helm path is documented, but target cluster values are not attached. | Provide target values and owner. |
| Rollback | Rollback runbooks exist, but no buyer drill is recorded. | Complete one rollback drill. |
| Observability | Health checks exist; buyer alert routing is not documented. | Define alert owner and thresholds. |
| Data handling | Synthetic fixture is non-PII; customer retention policy is not attached. | Confirm retention and access rules. |
| Support | Starter has best-effort async support and no SLA by default. | Confirm this matches launch risk. |

### Readiness recommendation

Use Starter + Production Readiness Package when the first production surface is known and the main adoption risk is
operational trust before launch. Use Growth + Production Readiness Package when the first-year plan already includes
multiple tenants, deployments, or up to six production recommendation surfaces.

## Completed example: Security / Procurement Review packet

### Procurement packet summary

- Package: Security / Procurement Review Package
- Scope: self-serve review for Commercial Evaluation, Starter, or Growth
- Review outcome: published artifacts are sufficient for standard procurement; Enterprise discovery is required for
  custom SLA, regulated deployment, OEM/resale, or custom legal/security terms

### Procurement artifact map

| Buyer question | Artifact |
| --- | --- |
| What data is required for a pilot? | [Security, privacy, compliance](../start-here/security-privacy-compliance.md) |
| What security docs can we forward? | [Security pack](../security/security-pack.md) |
| What are the standard commercial terms? | [Commercial license](../licensing/commercial_license.md) |
| What support is included? | [SLA and support schedule](../security/sla-schedule.md) |
| What do we send to buy? | [Self-serve procurement](self-serve-procurement.md) |
| What if we need Enterprise terms? | [Enterprise readiness evidence](enterprise-readiness-evidence.md) |

### Procurement open questions

- Does the buyer require custom DPA/SCC language beyond the published baseline?
- Does the buyer require an SLA or service credits?
- Does the buyer require multi-region HA or regulated deployment commitments?
- Does the buyer need OEM, resale, or third-party hosting rights?

If any answer is yes, use Enterprise discovery before relying on custom terms.

## Completed example: decision note

```text
Decision: hold for customer data

Reasoning:
- The local proof kit returned non-empty recommendations for tenant demo and surface home.
- The manifest and eval reports were generated from synthetic, non-PII ecommerce data.
- The evaluation loop is reproducible, but product lift has not been measured on customer data.

Next step:
- Run a guided customer-environment pilot on one real surface.
- Choose Starter + Pilot Integration Review if integration readiness is the main risk.
- Choose Starter + Production Readiness Package if launch readiness is the main risk.
```

## Which first-year package should I buy first?

| Buyer state | Recommended purchase |
| --- | --- |
| You only need commercial terms for a time-boxed non-production pilot. | Commercial Evaluation |
| You know the first production surface but want review of integration and instrumentation. | Starter + Pilot Integration Review |
| You know the first production surface and need launch-readiness review. | Starter + Production Readiness Package |
| You already expect up to 3 tenants/deployments or up to 6 production surfaces in year one. | Growth + Production Readiness Package |
| You need OEM/resale, regulated deployment, multi-region HA, custom SLA, or custom legal/security terms. | Enterprise discovery |

First-year bundles are private Order Form packaging examples. They do not change public plan prices, plan entitlements,
support defaults, renewal pricing, or service-package scope unless the signed Order Form says otherwise.

## Read next

- Commercial proof kit: [Commercial proof kit](../tutorials/commercial-proof-kit.md)
- Guided evaluation: [Guided evaluation](../evaluate/guided-evaluation.md)
- Pricing: [Pricing](../pricing/index.md)
- Self-serve procurement: [Self-serve procurement](self-serve-procurement.md)
- Order form template: [Order form template](../licensing/order_form.md)
