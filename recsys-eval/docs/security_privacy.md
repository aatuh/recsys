# Security and privacy notes

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
