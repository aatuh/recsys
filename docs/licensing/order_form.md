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
- Plan: ☐ Starter ☐ Growth ☐ Enterprise (custom)

### 1.2 Term

- Start date: `YYYY-MM-DD`
- End date: `YYYY-MM-DD`
- Renewal: ☐ Annual renewal ☐ Non-renewing (evaluation/custom)

### 1.3 Fees (excl. VAT)

- License fee: € `AMOUNT`
- Support fee (if any): € `AMOUNT`
- Total: € `AMOUNT`
- Payment terms: `e.g., Net 14 / Net 30`
- Billing method: ☐ Invoice ☐ Payment link/credit card ☐ Other: `...`

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

### 2.4 Regions / Affiliates (if applicable)

- Regions allowed: `e.g., EU / global`
- Affiliates allowed: ☐ Yes ☐ No (details if yes): `...`

### 2.5 OEM / Resale / Third-Party Hosting

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
- Term: `2026-03-01` to `2027-02-28` (annual renewal)
- License fee: € `9,900`
- Support fee: € `0`
- Total: € `9,900` (excl. VAT)
- Payment terms: `Net 30`
- Authorized Tenants: `1`
- Authorized Production Deployments: `1`
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
- Term: `2026-03-01` to `2027-02-28` (annual renewal)
- License fee: € `24,900`
- Support fee: € `0`
- Total: € `24,900` (excl. VAT)
- Payment terms: `Net 30`
- Authorized Tenants: `3`
- Authorized Production Deployments: `3`
- Support tier: `Growth default async`
- Response target: `First async response target within 2 business days`
- Delivery method: `Private container registry`

---

## 9. Fallback text for common legal/procurement edits

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
- Evaluation and licensing: [Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)
