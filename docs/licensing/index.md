# Licensing

This repository is a **multi-license** codebase. Different directories are licensed under different terms.

## Quick map

Path/component: `recsys-eval/**`
License: Apache License 2.0
Purpose: Offline evaluation & reporting tooling.

Path/component: `api/**`, `recsys-algo/**`, `recsys-pipelines/**`, and everything else unless stated otherwise
License: GNU AGPL v3
Purpose: Serving API, algorithms, pipelines, ops templates

The authoritative license texts are in:

- [`LICENSE`](https://github.com/aatuh/recsys/blob/master/LICENSE) (AGPLv3)
- [`recsys-eval/LICENSE`](https://github.com/aatuh/recsys/blob/master/recsys-eval/LICENSE) (Apache-2.0)

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
(see [`commercial.md`](commercial.md)).

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

- Commercial use & how to buy: [`licensing/commercial.md`](commercial.md)

## Commercial licensing

We offer a commercial license as an alternative set of terms for parts of this repository covered by AGPLv3.

See:

- [`commercial.md`](commercial.md) (overview, what you get, and how to buy)
- [`pricing.md`](pricing.md) (tiers)

## Third-party dependencies

This project depends on third-party open source libraries with their own licenses. Compliance for those dependencies is
separate from this project’s license. If you publish releases, include SBOMs and/or dependency license reports as part
of your compliance workflow.

## Questions

If you have licensing questions, open an issue titled **"Licensing question"** (public) or contact us privately if your
question contains confidential details.
