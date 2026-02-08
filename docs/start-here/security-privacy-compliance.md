---
diataxis: explanation
tags:
  - security
  - business
  - ops
---
# Security, privacy, and compliance (overview)
This page explains Security, privacy, and compliance (overview) and how it fits into the RecSys suite.


## Who this is for

Engineering leads, product owners, SRE/on-call, and security/compliance reviewers evaluating or adopting the **RecSys
suite**.

## What you will get

- A practical “shared responsibility” view for running the suite
- The minimum privacy posture needed for a pilot (and what changes for production)
- A checklist for security review starting points

## Scope: what the suite is (and is not)

- The RecSys suite is typically **self-hosted**: you run the infrastructure and own the data.
- The suite does not require raw PII: use **pseudonymous stable identifiers**.
- The suite does not attempt to be a full privacy/compliance program. You still need org-level policies for:
  - data classification and retention
  - access control and auditing
  - data subject requests (GDPR/CCPA) and deletion workflows

## Data you will handle

At a minimum, adopting the suite introduces these data flows:

- **Serving requests**: `tenant_id`, `surface`, and a pseudonymous `user_id`/`session_id` context
- **Exposure logs**: the ranked list served, with ranks and a `request_id` for attribution
- **Outcome logs**: clicks/conversions with the same `request_id`
- **Artifacts** (optional): aggregated signals (popularity, co-visitation, etc.) stored in object storage

Treat exposure and outcome logs as **sensitive**. Even if identifiers are pseudonymous, they are often still considered
personal data under many policies.

## Identity and PII guidance (baseline)

- Do not send or log raw PII (email, phone, address).
- Prefer **pseudonymous, stable identifiers** (for example: an internal UUID or a one-way hash you control).
- If you enable eval-compatible exposure logs in `recsys-service`, set a secret salt:
  - `EXPOSURE_HASH_SALT=<secret>`

Changing the salt breaks joins over time; rotate intentionally (and treat it like a credential).

## Access control and hardening

### Auth modes

`recsys-service` supports JWT and API keys. For local development it can also accept dev headers.

Production guidance:

- Disable dev headers: `DEV_AUTH_ENABLED=false`
- Require production auth (`JWT_AUTH_ENABLED=true` and/or `API_KEY_ENABLED=true`)
- Ensure tenant scope comes from trusted auth claims (`AUTH_TENANT_CLAIMS`) or a trusted gateway

### Admin endpoints

Admin endpoints can change configuration/rules and invalidate caches. Treat them as control-plane:

- restrict network access (private ingress / allow-list / VPN)
- require admin roles (`AUTH_ADMIN_ROLE`) and strong identity
- enable audit logging for admin actions (`AUDIT_LOG_ENABLED=true`)

### Rate limiting and abuse

Enable per-tenant rate limiting in production and monitor throttling:

- `TENANT_RATE_LIMIT_ENABLED=true`

## Logging and retention

- Configure exposure logging intentionally:
  - `EXPOSURE_LOG_ENABLED=true`
  - set retention (`EXPOSURE_LOG_RETENTION_DAYS`) and storage controls (permissions, encryption, backups)
- Treat evaluation outputs as sensitive artifacts:
  - reports may reveal behavior patterns or business logic
  - store them with appropriate access control and retention

## Compliance notes (high level)

- **GDPR/CCPA**: pseudonymous identifiers can still be personal data. Plan for deletion and retention limits.
- **Data residency**: choose DB/object store regions consistent with your policy.
- **Auditability**: enable audit logs and keep `request_id` propagation end-to-end for investigations.

## Quick checklist (start here)

- [ ] Use pseudonymous IDs; do not log raw PII.
- [ ] Set `EXPOSURE_HASH_SALT` when logging exposures for evaluation.
- [ ] Disable dev auth headers in production (`DEV_AUTH_ENABLED=false`).
- [ ] Restrict admin endpoints (network + roles) and enable audit logging.
- [ ] Define retention for exposure/outcome logs and evaluation reports.

## Read next

- Responsibilities (RACI): [Responsibilities (RACI): who owns what](responsibilities.md)
- Production readiness checklist: [Production readiness checklist (RecSys suite)](../operations/production-readiness-checklist.md)
- Exposure logging and attribution: [Exposure logging and attribution](../explanation/exposure-logging-and-attribution.md)
- Security policy (reporting vulnerabilities): [Security Policy](../project/security.md)
