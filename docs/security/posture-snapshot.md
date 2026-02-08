---
diataxis: reference
tags:
  - security
  - procurement
  - business
---
# Security posture snapshot

This page is a dated, procurement-friendly summary of the current RecSys security posture.

## Document controls

- Owner: RecSys maintainers (`contact@recsys.app`)
- Last reviewed: 2026-02-08
- Next review due: 2026-05-08
- Source of truth for broader guidance: [Security, privacy, compliance](../start-here/security-privacy-compliance.md)

## Deployment and data model (current)

- Primary deployment model: self-hosted (customer-operated infrastructure)
- Data stance: pseudonymous identifiers supported; raw PII is not required
- Core sensitive records:
  - recommendation exposures
  - outcomes/events
  - admin audit records

## Authentication and access control baseline

- Supported auth modes: JWT and API keys
- Development-only mode: dev headers (must be disabled in production)
- Admin endpoints:
  - should be private-network only
  - should require admin role claims
  - should have audit logging enabled

Reference: [Security, privacy, compliance](../start-here/security-privacy-compliance.md)

## Logging and auditability baseline

- Exposure logging can be enabled for evaluation and attribution
- `request_id` propagation is expected end-to-end for reliable joins and investigations
- Admin actions can be audited for config/rules/cache control-plane changes

References:

- [Exposure logging & attribution](../explanation/exposure-logging-and-attribution.md)
- [Data contracts](../reference/data-contracts/index.md)
- [Admin API](../reference/api/admin.md)

## Known constraints relevant to security review

- Tenant creation is currently DB bootstrap based (no tenant-create admin endpoint)
- Kafka source is scaffolded but not implemented as a streaming consumer

Reference: [Known limitations](../start-here/known-limitations.md)

## What is not asserted on this page

This page does not claim external certifications, managed hosting controls, or contractual terms not explicitly
published in docs.

For legal/contractual terms, use:

- [Commercial license](../licensing/commercial_license.md)
- [Order form template](../licensing/order_form.md)
- [Commercial procurement artifacts](commercial-procurement-artifacts.md)

## Read next

- Security pack: [Security pack](security-pack.md)
- Procurement pack: [Procurement pack](../for-businesses/procurement-pack.md)
