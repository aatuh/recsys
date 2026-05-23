# Licensing

This repository is multi-license. The file-scope license model is defined by current repository license files and
`.reuse/dep5`.

## License map

| Path | License | Notes |
| --- | --- | --- |
| `recsys-eval/**` | Apache-2.0 | Offline evaluation tooling. See `recsys-eval/LICENSE` and `.reuse/dep5`. |
| All other paths unless stated otherwise | AGPL-3.0-only | Serving API, ranking library, pipelines, charts, scripts, and docs. See `LICENSES/AGPL-3.0-only.txt` and `.reuse/dep5`. |
| `LICENSES/Apache-2.0.txt` | Apache-2.0 | License text. |
| `LICENSES/AGPL-3.0-only.txt` | AGPL-3.0-only | License text. |

If a future per-file SPDX header conflicts with this page, use the per-file or closest directory-level license notice
as the source of truth.

## Decision tree

1. Are you only using `recsys-eval/**`?
   - Yes: use it under Apache-2.0 terms.
   - No: continue.
2. Are you deploying or offering network access to `api/**`, `recsys-algo/**`, or `recsys-pipelines/**`?
   - Yes: AGPL-3.0-only applies unless you buy commercial terms.
   - No: your obligations depend on how you use, modify, and distribute the code.
3. Do you need private modifications or proprietary deployment terms for AGPL-covered components?
   - Yes: use a commercial license.
   - No: comply with AGPL-3.0-only.

This page is not legal advice. Involve counsel when licensing affects business decisions.

## Commercial terms

Commercial licensing is the alternative path for organizations that need private-use rights, commercial release
artifacts, patch/update access, or procurement certainty for AGPL-covered components. Typical commercial scope is
recorded in an order form: tenants, production deployments, production recommendation surfaces, term, support, and any
fixed-scope service package.

## Contact

- Public licensing questions: open a GitHub issue titled `Licensing question`.
- Confidential commercial licensing inquiries: message [Aatu Harju on LinkedIn](https://www.linkedin.com/in/aatu-harju/).
- Do not include secrets, customer data, or vulnerability details in licensing inquiries.

## Read next

- [Pricing](pricing.md)
- [Procurement and Trust Review](procurement.md)
- [Security](../security.md)
- [Support](../support.md)
