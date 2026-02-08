---
diataxis: reference
tags:
  - security
  - privacy
  - procurement
  - business
---
# Subprocessors and distribution details

This page publishes the default subprocessor and delivery disclosure model for self-serve plans.

## Document controls

- Owner: RecSys maintainers (`contact@recsys.app`)
- Last reviewed: 2026-02-08
- Next review due: 2026-05-08

## Scope

This page applies to:

- Commercial Evaluation
- Starter
- Growth

Enterprise plans may add provider-specific annexes in the Order Form.

## Runtime data processing model

RecSys is self-hosted by default.

- Customer controls production infrastructure, storage, and runtime recommendation data.
- Vendor does not require access to customer production recommendation payloads by default.

## Default processing disclosure

| Scope | Typical data categories | Processor role | Notes |
| --- | --- | --- | --- |
| License fulfillment and billing | Business contact details, company identifiers, contract metadata | Vendor | Required to provide commercial entitlement and invoicing |
| Support operations | Business contact details and troubleshooting material shared by customer | Vendor | Customer should avoid sending raw PII unless required for incident handling |
| Artifact distribution access logs | Organization/account identifiers, pull/access metadata | Vendor | Used for fulfillment, abuse prevention, and support diagnostics |
| Runtime recommendation traffic and logs | Recommendation request/response data, exposure/outcome logs | Customer | Stored and operated in customer-managed infrastructure by default |

## Data residency and transfer notes

- Customer runtime data residency is determined by customer infrastructure choices.
- Vendor contract/support metadata residency and transfer controls follow the baseline terms in:
  [DPA and SCC terms](dpa-and-scc.md)

## Enterprise customization

Use Enterprise terms if you require:

- provider-name annexes,
- region-locked delivery or support processing,
- additional supply-chain disclosures.

These requirements are captured in a signed Order Form.

## Read next

- DPA and SCC baseline: [DPA and SCC terms](dpa-and-scc.md)
- Procurement artifact index: [Commercial procurement artifacts](commercial-procurement-artifacts.md)
- Security pack: [Security pack](security-pack.md)
