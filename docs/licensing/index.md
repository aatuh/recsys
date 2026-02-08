---
diataxis: reference
tags:
  - licensing
  - business
  - overview
---
# Licensing

This repository is a **multi-license** codebase. Different directories are licensed under different terms.

!!! info "Canonical page"
    This page is canonical for the **AGPLv3 vs commercial** decision tree and obligations. For plan pricing, see
    [pricing](../pricing/index.md). For the recommended evaluation path and procurement checklist, see the
    [buyer guide](../pricing/evaluation-and-licensing.md).

## Document controls

- Owner: RecSys maintainers (`contact@recsys.app`)
- Last legal/doc review: 2026-02-08
- Next review due: 2026-05-08
- Order of precedence: per-file SPDX/header notices and signed commercial agreements > this page

## Quick map

Path/component: `recsys-eval/**`
License: Apache License 2.0
Purpose: Offline evaluation & reporting tooling.

Path/component: `api/**`, `recsys-algo/**`, `recsys-pipelines/**`, and everything else unless stated otherwise
License: GNU AGPL v3
Purpose: Serving API, algorithms, pipelines, ops templates

The authoritative license texts are in:

- [AGPLv3 license text](https://github.com/aatuh/recsys/blob/master/LICENSE)

  Location: `LICENSE`

- [Apache-2.0 license text for recsys-eval](https://github.com/aatuh/recsys/blob/master/recsys-eval/LICENSE)

  Location: `recsys-eval/LICENSE`

## How to determine the license for a file

We recommend (and are moving toward) using **SPDX license identifiers** in file headers and storing license texts in a
`LICENSES/` directory (REUSE specification style).

If there is ever a mismatch between this page and file headers, the **per-file SPDX identifier**
(or the closest directory-level declaration) is the source of truth.

## Using `recsys-eval` (Apache-2.0)

You can use, modify, and redistribute `recsys-eval` under Apache-2.0 terms, including in proprietary systems, provided
you comply with the Apache-2.0 conditions (e.g., preserving notices).

## Using the serving stack (AGPLv3)

The serving stack is licensed under the **GNU Affero General Public License v3 (AGPLv3)**.

If your organization cannot or does not want to comply with AGPL obligations, you can obtain a **commercial license**
(see [Commercial Use & Licensing](commercial.md)).

## Do you need a commercial license? (decision tree)

This is not legal advice. If licensing affects your business, involve counsel.

1. Are you only using `recsys-eval`?
   - Yes → it is Apache-2.0 (commercial license not needed; comply with Apache-2.0 conditions).
   - No → continue.
1. Are you deploying or offering network access to the serving stack (`api/**`, `recsys-algo/**`,
   `recsys-pipelines/**`)?
   - No → your obligations depend on how you use/distribute the code.
   - Yes → AGPLv3 applies.
1. Can you comply with AGPLv3 terms for your deployment (including source-offer obligations for modifications)?
   - Yes → use under AGPLv3.
   - No / we need modifications private → request a commercial license.

Next step:

- Commercial use & how to buy: [Commercial Use & Licensing](commercial.md)
- Self-serve procurement path: [Self-serve procurement](../for-businesses/self-serve-procurement.md)

## Obligations by scenario (quick reference)

This table is an operational summary, not legal advice.

| Scenario | Applicable path | What you must do | Commercial license required |
| --- | --- | --- | --- |
| Use only `recsys-eval/**` in internal or proprietary workflows | Apache-2.0 | Keep Apache notices and comply with Apache-2.0 terms | No |
| Deploy serving stack (`api/**`, `recsys-algo/**`, `recsys-pipelines/**`) with network access and comply with AGPL | AGPLv3 | Comply with AGPL obligations, including source-offer obligations for modifications | No |
| Deploy serving stack with private modifications and do not want AGPL source-offer obligations | Commercial terms | Buy and operate under signed commercial agreement + Order Form scope | Yes |
| Internal evaluation under vendor-provided commercial evaluation artifacts | Evaluation license | Follow evaluation term, non-production limits, and restrictions in evaluation terms | Yes (evaluation license) |
| OEM/resale/third-party hosting use case | Commercial enterprise terms | Negotiate explicit rights in signed Order Form | Yes |

If a use case is ambiguous, involve counsel and contact `contact@recsys.app` with deployment details.

## Commercial licensing

We offer a commercial license as an alternative set of terms for parts of this repository covered by AGPLv3.

See:

- [Commercial Use & Licensing](commercial.md) (overview, what you get, and how to buy)
- Pricing overview (commercial plans): [Pricing](../pricing/index.md)
- Legal pricing definitions (order forms): [Pricing definitions](pricing.md)

## Third-party dependencies

This project depends on third-party open source libraries with their own licenses. Compliance for those dependencies is
separate from this project’s license. If you publish releases, include SBOMs and/or dependency license reports as part
of your compliance workflow.

## Questions

If you have licensing questions, open an issue titled **"Licensing question"** (public) or contact us privately if your
question contains confidential details.

## Read next

- Pricing: [Pricing](../pricing/index.md)
- Self-serve procurement: [Self-serve procurement](../for-businesses/self-serve-procurement.md)
- Evaluation and licensing: [Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)
- Final go/no-go review: [Decision readiness matrix](../for-businesses/decision-readiness.md)
