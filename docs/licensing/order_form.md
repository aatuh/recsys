---
diataxis: reference
tags:
  - licensing
  - commercial
  - business
---
# Order Form (Template) — RecSys Commercial License

**Order Form ID:** `OF-YYYY-NNN`  
**Effective Date:** `YYYY-MM-DD`  
**Vendor:** `PakkaSys` ("Vendor")  
**Customer:** `CUSTOMER_LEGAL_NAME` ("Customer")

This Order Form is governed by and incorporates the **Commercial License Agreement (RecSys) v`1.0`**
("Agreement"). Capitalized terms not defined here have the meaning given in the Agreement.

## Document controls

- Owner: RecSys maintainers (`contact@recsys.app`)
- Last legal/doc review: 2026-02-08
- Next review due: 2026-05-08

---

## 1. Products and Fees

### 1.1 Product

- Product: **RecSys Commercial License**
- Plan: ☐ Commercial Evaluation ☐ Starter ☐ Growth ☐ Enterprise (custom)
- First-year bundle (if any): ☐ None ☐ Starter + Pilot Integration Review ☐ Starter + Production Readiness Package
  ☐ Growth + Production Readiness Package ☐ Other: `...`
- Fixed-scope service package(s): ☐ None ☐ Pilot Integration Review ☐ Production Readiness Package ☐ Security /
  Procurement Review Package ☐ Other: `...`

First-year bundles are Order Form packaging examples, not new public tiers. They do not change plan entitlements, list
prices, support defaults, renewal pricing, or service-package scope unless this Order Form states otherwise.

### 1.2 Term

- Start date: `YYYY-MM-DD`
- End date: `YYYY-MM-DD`
- Renewal: ☐ Annual renewal ☐ Non-renewing (evaluation/custom)

### 1.3 Fees (excl. VAT)

- License fee: € `AMOUNT`
- First-year bundle total (if applicable): € `AMOUNT`
- Fixed-scope service package fee (if any): € `AMOUNT`
- Support fee (if any): € `AMOUNT`
- Total: € `AMOUNT`
- Payment terms: `e.g., Net 14 / Net 30`
- Billing method: ☐ Invoice ☐ Payment link/credit card ☐ Other: `...`

For first-year bundles, either itemize the license and service package fees or use the bundle total. Do not double count
the same service package in the Total line.

### 1.4 Taxes

Customer is responsible for applicable VAT/sales taxes unless a valid exemption applies. Customer VAT ID:
`VAT_ID`

---

## 2. Authorized Scope (Entitlements)

### 2.1 Tenants

- Authorized Tenants: `N`

### 2.2 Production Deployments

- Authorized Production Deployments: `N`

### 2.3 Non-Production Environments

Included at no extra charge:

- Up to **2** non-prod environments per Production Deployment (dev/staging/sandbox)

### 2.4 Production Recommendation Surfaces

- Authorized Production Recommendation Surfaces: `N`
- Surface names/descriptions: `e.g., home feed, product detail page`

### 2.5 Fixed-Scope Service Package Scope (if purchased)

- Package(s): `e.g., Pilot Integration Review`
- Review milestone: `pilot readiness / production readiness / security procurement`
- In-scope tenant/deployment/surface: `...`
- Review inputs: `links to docs, configs, reports, runbooks, or artifacts`
- Guided-evaluation proof-kit outputs, if applicable:
  `recommendation response / manifest / eval report / decision note`

Fixed-scope service packages produce written review deliverables. They are advisory/review packages and do not include
managed hosting, production on-call, unlimited custom development, SLA commitments, or guaranteed KPI lift unless
explicitly stated in Section 5.

### 2.6 First-Year Bundle Scope (if purchased)

- Bundle: `e.g., Starter + Pilot Integration Review`
- Bundle components: `plan + fixed-scope service package`
- First-year total: € `AMOUNT`
- Renewal default: `selected plan renewal only / custom`

Unless Section 5 states otherwise, bundle pricing applies only to the first year and renewals follow the selected plan's
standard renewal terms without the one-time service package fee.

Example first-year bundle totals:

| Bundle | First-year total |
| --- | ---: |
| Starter + Pilot Integration Review | €14,900 |
| Starter + Production Readiness Package | €22,400 |
| Growth + Production Readiness Package | €37,400 |

### 2.7 Regions / Affiliates (if applicable)

- Regions allowed: `e.g., EU / global`
- Affiliates allowed: ☐ Yes ☐ No (details if yes): `...`

### 2.8 OEM / Resale / Third-Party Hosting

- OEM/resale: ☐ Not allowed ☐ Allowed (details): `...`
- Third-party hosting: ☐ Not allowed ☐ Allowed (details): `...`

---

## 3. Support (If Purchased)

Support tier: ☐ None ☐ Best effort async ☐ 8x5 SLA ☐ 24x7 SLA  
Support channel(s): `email/ticket portal`  
Response targets (if applicable): `e.g., P1 4h, P2 1bd, ...`  
Exclusions/limits: `optional`

Default schedule for self-serve plans:
[SLA and support schedule](../security/sla-schedule.md)

---

## 4. Delivery and Access

- Delivery method: ☐ Private container registry ☐ Download link ☐ Other
- Registry/URL: `REGISTRY_URL or DOWNLOAD_URL`
- Credential delivery: `how credentials are provided`
- License file delivery: `signed license.json/JWT, delivered via email/portal`

---

## 5. Special Terms (Optional)

`Any negotiated terms, e.g., security addendum, DPA reference, custom liability cap, etc.`

Default self-serve legal/security references:

- DPA/SCC baseline: [DPA and SCC terms](../security/dpa-and-scc.md)
- Subprocessor/distribution disclosure: [Subprocessors and distribution details](../security/subprocessors.md)
- Support/SLA defaults: [SLA and support schedule](../security/sla-schedule.md)

Use this section for Enterprise/custom overrides only.

---

## 6. Signatures

**Vendor:** ________________________  Date: __________  
Name/Title: ________________________

**Customer:** ______________________  Date: __________  
Name/Title: ________________________

---

## 7. Filled example — Starter

Example only. Replace with actual customer and commercial terms before signature.

- Order Form ID: `OF-2026-001`
- Effective Date: `2026-03-01`
- Customer: `Example Commerce Oy`
- Plan: Starter
- First-year bundle: None
- Term: `2026-03-01` to `2027-02-28` (annual renewal)
- License fee: € `9,900`
- First-year bundle total: `N/A`
- Fixed-scope service package fee: € `0`
- Support fee: € `0`
- Total: € `9,900` (excl. VAT)
- Payment terms: `Net 30`
- Authorized Tenants: `1`
- Authorized Production Deployments: `1`
- Authorized Production Recommendation Surfaces: `2`
- Support tier: `Best effort async`
- Response target: `No SLA`
- Delivery method: `Private container registry`

---

## 8. Filled example — Growth

Example only. Replace with actual customer and commercial terms before signature.

- Order Form ID: `OF-2026-002`
- Effective Date: `2026-03-01`
- Customer: `Example Retail Group Ltd`
- Plan: Growth
- First-year bundle: None
- Term: `2026-03-01` to `2027-02-28` (annual renewal)
- License fee: € `24,900`
- First-year bundle total: `N/A`
- Fixed-scope service package fee: € `0`
- Support fee: € `0`
- Total: € `24,900` (excl. VAT)
- Payment terms: `Net 30`
- Authorized Tenants: `3`
- Authorized Production Deployments: `3`
- Authorized Production Recommendation Surfaces: `6`
- Support tier: `Growth default async`
- Response target: `First async response target within 2 business days`
- Delivery method: `Private container registry`

---

## 9. Filled example — Starter + Pilot Integration Review first-year bundle

Example only. Replace with actual customer and commercial terms before signature.

- Order Form ID: `OF-2026-003`
- Effective Date: `2026-03-01`
- Customer: `Example Marketplace GmbH`
- Plan: Starter
- First-year bundle: Starter + Pilot Integration Review
- Term: `2026-03-01` to `2027-02-28` (annual renewal)
- License fee: € `9,900`
- First-year bundle total: € `14,900`
- Fixed-scope service package fee: included in bundle
- Support fee: € `0`
- Total: € `14,900` (excl. VAT)
- Payment terms: `Net 30`
- Authorized Tenants: `1`
- Authorized Production Deployments: `1`
- Authorized Production Recommendation Surfaces: `2`
- Fixed-scope service package: `Pilot Integration Review`
- Review milestone: `pilot readiness`
- Renewal default: `Starter renewal only unless a later Order Form says otherwise`
- Support tier: `Best effort async`
- Response target: `No SLA`
- Delivery method: `Private container registry`

---

## 10. Filled example — Starter + Production Readiness Package first-year bundle

Example only. Replace with actual customer and commercial terms before signature.

- Order Form ID: `OF-2026-004`
- Effective Date: `2026-03-01`
- Customer: `Example Commerce GmbH`
- Plan: Starter
- First-year bundle: Starter + Production Readiness Package
- Term: `2026-03-01` to `2027-02-28` (annual renewal)
- License fee: € `9,900`
- First-year bundle total: € `22,400`
- Fixed-scope service package fee: included in bundle
- Support fee: € `0`
- Total: € `22,400` (excl. VAT)
- Payment terms: `Net 30`
- Authorized Tenants: `1`
- Authorized Production Deployments: `1`
- Authorized Production Recommendation Surfaces: `2`
- Fixed-scope service package: `Production Readiness Package`
- Review milestone: `production readiness`
- Review inputs: `deployment shape, rollback runbooks, observability, hardening checklist`
- Renewal default: `Starter renewal only unless a later Order Form says otherwise`
- Support tier: `Best effort async`
- Response target: `No SLA`
- Delivery method: `Private container registry`

---

## 11. Filled example — Growth + Production Readiness Package first-year bundle

Example only. Replace with actual customer and commercial terms before signature.

- Order Form ID: `OF-2026-005`
- Effective Date: `2026-03-01`
- Customer: `Example Retail Group Ltd`
- Plan: Growth
- First-year bundle: Growth + Production Readiness Package
- Term: `2026-03-01` to `2027-02-28` (annual renewal)
- License fee: € `24,900`
- First-year bundle total: € `37,400`
- Fixed-scope service package fee: included in bundle
- Support fee: € `0`
- Total: € `37,400` (excl. VAT)
- Payment terms: `Net 30`
- Authorized Tenants: `3`
- Authorized Production Deployments: `3`
- Authorized Production Recommendation Surfaces: `6`
- Fixed-scope service package: `Production Readiness Package`
- Review milestone: `production readiness`
- Review inputs: `deployment shape, rollback runbooks, observability, hardening checklist`
- Renewal default: `Growth renewal only unless a later Order Form says otherwise`
- Support tier: `Growth default async`
- Response target: `First async response target within 2 business days`
- Delivery method: `Private container registry`

---

## 12. Fallback text for common legal/procurement edits

These short clauses can be copied to Section 5 (Special Terms) when needed.

- Liability cap fallback:
  `The limitation of liability in Section 12 of the Agreement applies. Any alternate cap must be stated explicitly in this Order Form.`
- Support fallback (no SLA):
  `Support is best-effort async. No service credits apply unless an SLA schedule is attached to this Order Form.`
- DPA/privacy fallback:
  `For self-serve plans, the published DPA/SCC baseline applies by reference. Any Enterprise/custom DPA/SCC rider must be stated explicitly in this Order Form.`
- Security incident process fallback:
  `Security escalation and notification contacts are defined in the support/security addendum attached to this Order Form, if any.`

## Read next

- Pricing: [Pricing](../pricing/index.md)
- Evaluation deliverables pack: [Evaluation deliverables pack](../for-businesses/evaluation-deliverables-pack.md)
- Evaluation and licensing: [Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)
