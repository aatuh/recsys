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

- Owner: RecSys maintainers (`contact@recsys.app`)
- Last pricing review: 2026-02-08
- Next review due: 2026-05-08
- Order of precedence: signed Order Form > this page > supporting reference pages

[Start evaluation](../evaluate/index.md){ .md-button .md-button--primary }
[Read licensing](../licensing/index.md){ .md-button }
[Self-serve procurement](../for-businesses/self-serve-procurement.md){ .md-button }
[Buyer guide](evaluation-and-licensing.md){ .md-button }
[Contact](#contact){ .md-button }

All prices below are **EUR, excl. VAT**.

Pilot plans commonly take **2–6 weeks** ([Pilot plan](../start-here/pilot-plan.md)), while the commercial evaluation
license term is **30 days** (unless extended in writing): [Evaluation license](../licensing/eval_license.md).

## Plans (commercial)

<div class="grid cards" markdown>

- **Commercial Evaluation (30 days)**  
  €490 one-time  
  **Who it’s for:** teams validating lift and operational fit under commercial terms.  
  **Scope:** 1 tenant · 1 deployment (non-prod allowed).  
  **Includes:** evaluation license, best-effort async support (no SLA), and access to signed commercial artifacts (if
  applicable).

- **Starter**  
  €9,900 / year  
  **Who it’s for:** one production deployment with a small number of surfaces.  
  **Scope:** 1 tenant · 1 production deployment (+ up to 2 non-prod).  
  **Includes:** commercial license grant, signed artifacts, updates, and best-effort async support (no SLA).

- **Growth**  
  €24,900 / year  
  **Who it’s for:** a few tenants and/or deployments, with faster response expectations.  
  **Scope:** up to 3 tenants and/or up to 3 production deployments.  
  **Includes:** everything in Starter + default async first response target within 2 business days (no SLA unless
  purchased as add-on).

- **Enterprise**  
  From €60,000 / year  
  **Who it’s for:** multi-region HA, OEM/resale, regulated environments, or custom terms.  
  **Scope:** custom.  
  **Includes:** custom scope/terms. Typical negotiation scope: SLA/service credits, DPA/SCC riders, liability/legal
  riders, OEM/resale rights, and custom deployment/support commitments.

</div>

## Add-ons (not included by default)

- Premium support / SLA (8×5 or 24×7)
- Security review package (SBOM/provenance guidance, hardening checklist)
- Fixed-scope onboarding (time-boxed)

## Plan chooser (objective thresholds)

Use the smallest plan that meets your current scope.

| If your current need is... | Recommended plan |
| --- | --- |
| Time-boxed pilot in non-production only (1 tenant, 1 deployment) | Commercial Evaluation |
| One tenant, one production deployment, no contractual SLA required | Starter |
| Up to 3 tenants and/or up to 3 production deployments, faster async response expectations | Growth |
| Multi-region HA, OEM/resale, regulated requirements, or custom legal/security terms | Enterprise |

If you exceed plan scope at any time, upgrade via a new Order Form.

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

## Ways to buy

Choose the smallest purchase that matches your current stage.

1. **Run a time-boxed evaluation under commercial terms**

   - Pick **Commercial Evaluation (30 days)** if you want to avoid AGPL uncertainty during the pilot.
   - Follow the buyer flow: [Evaluation, pricing, and licensing (buyer guide)](evaluation-and-licensing.md)

2. **Use the self-serve path (Commercial Evaluation/Starter/Growth)**

   - Use published legal/security artifacts and standard defaults.
   - Follow: [Self-serve procurement](../for-businesses/self-serve-procurement.md)

3. **Use Enterprise procurement (custom terms)**

   - Required for OEM/resale, regulated environments, custom legal/security terms, or custom SLA commitments.
   - Submit scope and requested terms via order form + contact channel.

4. **Use only `recsys-eval` (Apache-2.0)**

   - If you only need offline evaluation tooling, you can use `recsys-eval/**` under Apache-2.0.
   - See the canonical decision tree: [Licensing](../licensing/index.md)

Next step:

- Self-serve plans: follow [Self-serve procurement](../for-businesses/self-serve-procurement.md)
- Enterprise/custom: use [Order form template](../licensing/order_form.md) and [`contact@recsys.app`][pricing_contact]

## Details and legal terms

- Legal pricing definitions (order forms): [Pricing definitions](../licensing/pricing.md)
- Commercial use & how to buy: [Commercial use](../licensing/commercial.md)
- Order form template: [Order form template](../licensing/order_form.md)
- DPA/SCC baseline: [DPA and SCC terms](../security/dpa-and-scc.md)
- Subprocessor disclosure: [Subprocessors and distribution details](../security/subprocessors.md)
- Evaluation terms: [Evaluation terms](../licensing/eval_license.md)

## Contact

Fastest fulfillment is async:

- Email: [`contact@recsys.app`][pricing_contact]
- LinkedIn: [`linkedin.com/showcase/recsys-suite`][recsys_linkedin]
- Suggested subject: “RecSys Commercial Evaluation” / “Starter” / “Growth” / “Enterprise”

[pricing_contact]: mailto:contact@recsys.app?subject=RecSys%20pricing%20inquiry
[recsys_linkedin]: https://www.linkedin.com/showcase/recsys-suite

## Read next

- Buyer guide: [Evaluation, pricing, and licensing (buyer guide)](evaluation-and-licensing.md)
- Self-serve path: [Self-serve procurement](../for-businesses/self-serve-procurement.md)
- Procurement pack: [Procurement pack (Security, Legal, IT, Finance)](../for-businesses/procurement-pack.md)
- Final go/no-go review: [Decision readiness matrix](../for-businesses/decision-readiness.md)
