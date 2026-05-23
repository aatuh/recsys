# Procurement and Trust Review

Use this page when a commercial reviewer needs a compact packet for security, legal, IT, finance, or engineering
approval. It is a navigation page, not a compliance certification.

## Review packet

| Reviewer | Send these pages | Decision they support |
| --- | --- | --- |
| Security | [Security](../security.md), [Operations](../operations.md), [Data Contracts](../reference/data-contracts.md) | Whether the deployment model, auth posture, logging, and data handling are reviewable. |
| Legal | [Licensing](licensing.md), [Pricing](pricing.md), [Support](../support.md) | Whether AGPL, commercial terms, and support expectations are acceptable. |
| IT or platform | [Architecture](../architecture.md), [Artifacts and Pipelines](../artifacts-and-pipelines.md), [Configuration](../reference/config.md) | Whether the service can be deployed and operated safely. |
| Evaluators | [Integration and Evaluation](../integration.md), [Evaluation Decisions](../evaluation-decisions.md), [Local End-to-End](../local-end-to-end.md) | Whether the pilot can produce credible evidence. |

## Baseline statements

- RecSys is usually self-hosted; the operator controls infrastructure, network policy, secrets, backups, and retention.
- Raw PII is not required for recommendation requests or evaluation datasets. Use pseudonymous user and session IDs.
- The repository documents JWT, API key, and local dev-header modes. Dev headers are not a production auth model.
- Admin config/rules/cache routes are control-plane operations and should be restricted to trusted operators.
- Commercial terms are captured in signed order forms. Signed terms override public documentation.

## Procurement checklist

- [ ] Confirm which components are in scope: service, ranking library, pipelines, evaluation CLI, or all of them.
- [ ] Confirm license path: Apache-2.0-only use of `recsys-eval/**`, AGPL-3.0-only use, or commercial terms.
- [ ] Confirm tenant count, production deployments, non-production environments, and recommendation surfaces.
- [ ] Confirm auth mode, tenant source, admin access model, and audit logging expectations.
- [ ] Confirm exposure/outcome retention, report handling, and who can access exported datasets.
- [ ] Confirm rollback procedures for config/rules and artifact manifests.
- [ ] Confirm support tier and response expectations from [Pricing](pricing.md) and [Support](../support.md).

## What is intentionally not asserted

This documentation does not claim external certifications, managed hosting controls, custom SLA terms, DPA terms,
subprocessor lists, data residency guarantees, or incident response commitments beyond what appears in current public
docs or a signed order form.

## Contact path

- Public licensing questions: open a GitHub issue titled `Licensing question`.
- Confidential commercial inquiries: use the contact path listed in [Pricing](pricing.md).
- Do not include secrets, customer data, vulnerability details, exploit payloads, or confidential logs in public issues.

## Read next

- [Licensing](licensing.md)
- [Pricing](pricing.md)
- [Security](../security.md)
- [Support](../support.md)
