---
diataxis: explanation
tags:
  - recsys-eval
  - security
  - privacy
---
# Security and privacy notes
This page explains Security and privacy notes and how it fits into the RecSys suite.


## Who this is for

Anyone handling real user data.

## What you will get

- A safe baseline for logging and storage
- Common pitfalls

## Baseline

- Do not log raw PII (email, phone, exact address).
- Prefer pseudonymous user IDs.
- Keep per-tenant boundaries strict.
- Limit report retention based on policy.

## Data minimization

Only log what you can justify measuring.
If you do not need a field, do not collect it.

## OPE and privacy

Propensity logging can include per-item scores. Treat them as sensitive.
They can leak model behavior and business logic if exposed carelessly.

## Read next

- Suite security overview: [Security, privacy, and compliance (overview)](../../start-here/security-privacy-compliance.md)
- Integration logging plan: [Integration: how to produce the inputs](integration.md)
- Data contracts: [Data contracts: what inputs look like](data_contracts.md)
