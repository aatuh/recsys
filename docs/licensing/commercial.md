---
diataxis: explanation
tags:
  - licensing
  - commercial
  - business
---
# Commercial Use & Licensing

This page explains how to purchase and use commercial terms for RecSys when you need private-use rights, commercial
release artifacts, patch access, and procurement certainty.

This page is informational and describes our commercial offering at a high level.

## Document controls

- Owner: RecSys maintainers
- Legal/doc review cadence: twice per year
- Order of precedence: signed Commercial License Agreement + signed Order Form > this page

## Why a commercial license?

A commercial license is for teams that want to run RecSys as business infrastructure under signed commercial terms. It
gives procurement, legal, and platform teams a clear path for:

- Keeping modifications private
- Using the software in proprietary stacks
- Accessing commercial release artifacts and patch/update channels defined by the Order Form
- Recording tenants, deployments, support expectations, and optional service packages in one commercial document

The AGPLv3 remains the open-source license for the serving stack. AGPLv3 is designed for software used over a network;
if you modify AGPL-covered code and provide network access to users, AGPLv3 requires offering those users access to the
Corresponding Source of your modified version (see Section 13). Commercial terms provide an alternative path when your
organization does not want to operate under those obligations.

## What is covered?

Commercial licensing applies to the components that are AGPLv3 in this repository, including:

- `recsys-service` (serving API; `docker compose` service name: `api`)
- `recsys-algo` (algorithms used by the service)
- `recsys-pipelines` (batch pipelines and artifact generation)

`recsys-eval` remains Apache-2.0.

## What you get when you buy

A commercial purchase includes, at minimum:

- A signed commercial license grant (agreement + order form)
- Any entitlement token/file described in the Order Form (bookkeeping only, **not DRM**)
- Access to commercial release artifacts when specified in the Order Form
- Security/patch update access according to the purchased plan
- Support terms as stated in the signed Order Form
- Optional fixed-scope service packages when purchased in the Order Form

See:

- Pricing overview (commercial plans): [Pricing](../pricing/index.md)
- Legal pricing definitions (order forms): [Pricing definitions](pricing.md)

## How to buy

Recommended low-friction flow:

1. Choose a plan in [Pricing](../pricing/index.md)
2. For Commercial Evaluation/Starter/Growth, follow the standard self-serve path:
   [Self-serve procurement](../for-businesses/self-serve-procurement.md)
   Direct checkout links:
   [Commercial Evaluation][stripe_commercial_evaluation], [Starter][stripe_starter_plan], [Growth][stripe_growth_plan]
   Need invoicing or an Enterprise quotation? Use the licensing contact options below.
3. Use direct contact only for Enterprise/custom terms (OEM/resale, custom legal/security/SLA commitments):

   - For public questions, open a GitHub issue titled **"Licensing question"**.
   - For confidential commercial licensing inquiries, message [Aatu Harju on LinkedIn][aatu_linkedin].

After payment: you’ll receive the license and invoice/receipt.

## Evaluation licenses

Evaluation terms for this project are documented in:

- [RecSys Evaluation License Terms](eval_license.md)

## Where are the legal terms?

Commercial terms live in:

- [RecSys Commercial License Agreement](commercial_license.md)
- [Order Form (Template) — RecSys Commercial License](order_form.md)

[aatu_linkedin]: https://www.linkedin.com/in/aatu-harju/
[stripe_commercial_evaluation]: https://buy.stripe.com/9B6fZh6r8dl97OP2sD0Jq00
[stripe_starter_plan]: https://buy.stripe.com/28EeVd7vcdl98ST6IT0Jq01
[stripe_growth_plan]: https://buy.stripe.com/aFafZh16Odl9c550kv0Jq02

## Read next

- Pricing: [Pricing](../pricing/index.md)
- Evaluation and licensing: [Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)
