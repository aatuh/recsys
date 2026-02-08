---
diataxis: explanation
tags:
  - licensing
  - commercial
  - business
---
# Commercial Use & Licensing

This page explains how to purchase and use a **commercial license** for parts of this repository that are otherwise
licensed under **AGPLv3**.

This page is informational and describes our commercial offering at a high level.

## Document controls

- Owner: RecSys maintainers (`contact@recsys.app`)
- Last legal/doc review: 2026-02-08
- Next review due: 2026-05-08
- Order of precedence: signed Commercial License Agreement + signed Order Form > this page

## Why a commercial license?

The AGPLv3 is designed for software used over a network. If you modify AGPL-covered code and provide network access to users,
AGPLv3 requires offering those users access to the Corresponding Source of your modified version (see Section 13).

A commercial license allows you to use the covered components under alternative terms, enabling:

- Internal or external deployment without AGPL source-offer obligations (subject to the commercial agreement)
- Keeping modifications private
- Using the software in proprietary stacks

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

See:

- Pricing overview (commercial plans): [Pricing](../pricing/index.md)
- Legal pricing definitions (order forms): [Pricing definitions](pricing.md)

## How to buy

Recommended low-friction flow:

1. Choose a plan in [Pricing](../pricing/index.md)
2. For Commercial Evaluation/Starter/Growth, follow the standard self-serve path:
   [Self-serve procurement](../for-businesses/self-serve-procurement.md)
3. Use direct contact only for Enterprise/custom terms (OEM/resale, custom legal/security/SLA commitments):

   - Open a GitHub issue titled **"Commercial licensing inquiry"** (public), or
   - Email: [`contact@recsys.app`][pricing_contact]
   - LinkedIn: [`linkedin.com/showcase/recsys-suite`][recsys_linkedin]

## Evaluation licenses

Evaluation terms for this project are documented in:

- [RecSys Evaluation License Terms](eval_license.md)

## Where are the legal terms?

Commercial terms live in:

- [RecSys Commercial License Agreement](commercial_license.md)
- [Order Form (Template) — RecSys Commercial License](order_form.md)

[pricing_contact]: mailto:contact@recsys.app?subject=RecSys%20pricing%20inquiry
[recsys_linkedin]: https://www.linkedin.com/showcase/recsys-suite

## Read next

- Pricing: [Pricing](../pricing/index.md)
- Evaluation and licensing: [Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)
