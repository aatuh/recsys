---
diataxis: reference
tags:
  - overview
  - business
  - licensing
---
# Pricing

RecSys is an auditable recommendation system suite with deterministic ranking and versioned ship/rollback.

!!! info "Canonical page"
    This page is canonical for **commercial plan definitions and prices**. For evaluation flow and licensing decision
    rules, see the [buyer guide](evaluation-and-licensing.md) and [licensing](../licensing/index.md).

## Document controls

- Owner: RecSys maintainers
- Last pricing review: 2026-02-08
- Next review due: 2026-05-08
- Order of precedence: signed Order Form > this page > supporting reference pages

[Start evaluation](../evaluate/index.md){ .md-button .md-button--primary }
[Read licensing](../licensing/index.md){ .md-button }
[Self-serve procurement](../for-businesses/self-serve-procurement.md){ .md-button }
[Buyer guide](evaluation-and-licensing.md){ .md-button }
[Guided evaluation](../evaluate/guided-evaluation.md){ .md-button }
[Contact](#contact){ .md-button }

All prices below are **EUR, excl. VAT**.

## Pricing posture

RecSys is priced as **auditable, self-hosted recommendation infrastructure**: deterministic ranking, evaluation,
versioned ship/rollback, and commercial procurement certainty. The production plans are intentionally not priced like a
low-cost library or a usage-only API wrapper.

We do not publish a radically cheaper production tier because the main adoption cost is not the license line item. The
real work is integration, exposure/outcome logging, operational readiness, and trust in the evaluation loop. Lowering
production list prices would create a weaker business-grade signal without reducing that implementation effort.

Pilot plans commonly take **2–6 weeks** ([Pilot plan](../start-here/pilot-plan.md)), while the commercial evaluation
license term is **30 days** (unless extended in writing): [Evaluation license](../licensing/eval_license.md).

## Plans (commercial)

<div class="grid cards" markdown>

- **Commercial Evaluation (30 days)**  
  €490 one-time  
  **Who it’s for:** teams validating lift and operational fit under commercial terms.  
  **Scope:** 1 tenant · 1 deployment (non-prod allowed).  
  **Includes:** evaluation license, best-effort async support (no SLA), and access to signed commercial artifacts (if
  applicable). The evaluation fee is credited toward Starter or Growth if purchased within 60 days after evaluation
  completion. One written 30-day extension may be granted for active pilots with clear scope.

- **Starter**  
  €9,900 / year  
  **Who it’s for:** one production deployment with one or two recommendation surfaces.
  **Scope:** 1 tenant · 1 production deployment (+ up to 2 non-prod) · 1-2 production recommendation surfaces.
  **Includes:** commercial license grant, signed artifacts, updates, and best-effort async support (no SLA).

- **Growth**  
  €24,900 / year  
  **Who it’s for:** a few tenants and/or deployments, with faster response expectations.  
  **Scope:** up to 3 tenants and/or up to 3 production deployments · up to 6 production recommendation surfaces.
  **Includes:** everything in Starter + default async first response target within 2 business days (no SLA unless
  purchased as add-on).

- **Enterprise**  
  From €60,000 / year  
  **Who it’s for:** custom scope only: OEM/resale, regulated deployments, multi-region HA, custom SLA, or custom
  legal/security terms.
  **Scope:** custom.  
  **Includes:** custom scope/terms. Typical negotiation scope: SLA/service credits, DPA/SCC riders, liability/legal
  riders, OEM/resale rights, and custom deployment/support commitments. Enterprise buyers should review the
  [Enterprise readiness evidence](../for-businesses/enterprise-readiness-evidence.md) before relying on custom terms.

</div>

## Fixed-scope services and add-ons

These packages reduce adoption risk. They are advisory/review packages, not unlimited custom development, managed
operations, or production on-call.

| Package | Price | What it covers |
| --- | ---: | --- |
| Pilot Integration Review | €5,000 one-time | Review one scoped pilot integration, instrumentation, tenant/surface setup, and evaluation readiness. |
| Production Readiness Package | €12,500 one-time | Review deployment shape, rollback runbooks, hardening checklist, observability, and production cutover risks. |
| Security / Procurement Review Package | €5,000 one-time | Support security/procurement review with SBOM/provenance guidance, hardening checklist review, and artifact navigation. |
| Premium support / SLA | Custom | 8×5 or 24×7 response commitments for Growth/Enterprise, captured in the Order Form. |

### Service package deliverables

- **Pilot Integration Review:** one written review memo covering request/response integration, tenant/surface setup,
  exposure/outcome instrumentation, and evaluation-readiness gaps.
- **Production Readiness Package:** one readiness memo covering deployment shape, rollback procedure, observability,
  hardening checklist status, and launch blockers.
- **Security / Procurement Review Package:** one review packet covering published security artifacts, SBOM/provenance
  navigation, procurement checklist gaps, and recommended follow-up questions.

## Private first-year bundles (Order Form examples)

These examples are private Order Form bundles, not new public tiers, Stripe products, or public discounts. They package
a published production plan with a fixed-scope review package for the first year.

| Example bundle | First-year total | Included components |
| --- | ---: | --- |
| Starter + Pilot Integration Review | €14,900 | Starter (€9,900/year) + Pilot Integration Review (€5,000 one-time) |
| Starter + Production Readiness Package | €22,400 | Starter (€9,900/year) + Production Readiness Package (€12,500 one-time) |
| Growth + Production Readiness Package | €37,400 | Growth (€24,900/year) + Production Readiness Package (€12,500 one-time) |

Bundles do not change plan entitlements, list prices, support defaults, renewal pricing, or service-package scope unless
a signed Order Form states otherwise. Renewal defaults to the selected plan's then-current commercial terms without the
one-time service package fee unless the Order Form says otherwise.

## Plan chooser (objective thresholds)

Use the smallest plan that meets your current scope.

| If your current need is... | Recommended plan |
| --- | --- |
| Time-boxed pilot in non-production only (1 tenant, 1 deployment) | Commercial Evaluation |
| One tenant, one production deployment, 1-2 production surfaces, no contractual SLA required | Starter |
| Up to 3 tenants/deployments and up to 6 production surfaces, faster async response expectations | Growth |
| OEM/resale, regulated deployments, multi-region HA, custom SLA, or custom legal/security terms | Enterprise |

If you exceed plan scope at any time, upgrade via a new Order Form.

## Discount policy

There is no public low-cost production tier. Private design-partner discounts may be offered through a custom Order
Form, but discounts do not change published list prices, plan scope, or support defaults.

## Default support terms by plan

Unless a signed Order Form says otherwise:

| Plan | Response commitment | SLA included by default |
| --- | --- | --- |
| Commercial Evaluation | Best-effort async | No |
| Starter | Best-effort async | No |
| Growth | Async first response target within 2 business days | No |
| Enterprise | Defined in Order Form | Depends on Order Form |

Standard schedule reference: [SLA and support schedule](../security/sla-schedule.md).

## What you receive

- A signed license grant (or equivalent commercial entitlement artifact)
- Credentials/instructions to pull signed commercial images from the private registry (if applicable)
- A recommended evaluation workflow and ship/rollback runbooks
- A buyer deliverables example pack for evaluation and fixed-scope review outputs:
  [Evaluation deliverables pack](../for-businesses/evaluation-deliverables-pack.md)

## Ways to buy

Choose the smallest purchase that matches your current stage.

1. **Run a time-boxed evaluation under commercial terms**

   - Pick **Commercial Evaluation (30 days)** if you want to avoid AGPL uncertainty during the pilot.
   - Apply the €490 evaluation fee toward Starter or Growth if you buy within 60 days after evaluation completion.
   - Request one written 30-day extension before expiry if the pilot is active and has clear scope.
   - Follow the buyer flow: [Evaluation, pricing, and licensing (buyer guide)](evaluation-and-licensing.md)

2. **Use the self-serve path (Commercial Evaluation/Starter/Growth)**

   - Use published legal/security artifacts and standard defaults.
   - Choose a private first-year bundle only through an Order Form if packaged first-year help reduces adoption risk.
   - Follow: [Self-serve procurement](../for-businesses/self-serve-procurement.md)
   - Checkout links:
     [Commercial Evaluation][stripe_commercial_evaluation] ·
     [Starter][stripe_starter_plan] ·
     [Growth][stripe_growth_plan]
   - Need invoicing or an Enterprise quotation? Use the contact options below.
   - Include any fixed-scope service package in the same procurement request or Order Form.

3. **Use Enterprise procurement (custom terms)**

   - Required for OEM/resale, regulated environments, custom legal/security terms, or custom SLA commitments.
   - Review the Enterprise evidence map before depending on HA, support, security, or operational commitments:
     [Enterprise readiness evidence](../for-businesses/enterprise-readiness-evidence.md)
   - Submit scope and requested terms via order form + contact channel.

4. **Use only `recsys-eval` (Apache-2.0)**

   - If you only need offline evaluation tooling, you can use `recsys-eval/**` under Apache-2.0.
   - See the canonical decision tree: [Licensing](../licensing/index.md)

Next step:

- Self-serve plans: follow [Self-serve procurement](../for-businesses/self-serve-procurement.md)
- Enterprise/custom: use [Order form template](../licensing/order_form.md) and the contact options below.

## Details and legal terms

- Legal pricing definitions (order forms): [Pricing definitions](../licensing/pricing.md)
- Enterprise readiness evidence: [Enterprise readiness evidence](../for-businesses/enterprise-readiness-evidence.md)
- Commercial use & how to buy: [Commercial use](../licensing/commercial.md)
- Order form template: [Order form template](../licensing/order_form.md)
- DPA/SCC baseline: [DPA and SCC terms](../security/dpa-and-scc.md)
- Subprocessor disclosure: [Subprocessors and distribution details](../security/subprocessors.md)
- Evaluation terms: [Evaluation terms](../licensing/eval_license.md)

## Contact

Fastest fulfillment is async:

- Public licensing questions: open a GitHub issue titled **"Licensing question"**
- Confidential commercial licensing inquiries: message Aatu Harju on LinkedIn: `linkedin.com/in/aatu-harju/`
- Suggested subject: “RecSys Commercial Evaluation” / “Starter” / “Growth” / “Enterprise”

[stripe_commercial_evaluation]: https://buy.stripe.com/9B6fZh6r8dl97OP2sD0Jq00
[stripe_starter_plan]: https://buy.stripe.com/28EeVd7vcdl98ST6IT0Jq01
[stripe_growth_plan]: https://buy.stripe.com/aFafZh16Odl9c550kv0Jq02

## Read next

- Buyer guide: [Evaluation, pricing, and licensing (buyer guide)](evaluation-and-licensing.md)
- Evaluation deliverables pack: [Evaluation deliverables pack](../for-businesses/evaluation-deliverables-pack.md)
- Guided evaluation: [Guided evaluation](../evaluate/guided-evaluation.md)
- Self-serve path: [Self-serve procurement](../for-businesses/self-serve-procurement.md)
- Enterprise evidence map: [Enterprise readiness evidence](../for-businesses/enterprise-readiness-evidence.md)
- Procurement pack: [Procurement pack (Security, Legal, IT, Finance)](../for-businesses/procurement-pack.md)
- Final go/no-go review: [Decision readiness matrix](../for-businesses/decision-readiness.md)
