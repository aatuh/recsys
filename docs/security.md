# Security

RecSys is usually self-hosted: the operator runs the infrastructure and controls the data. This page summarizes the
repository posture and the minimum security expectations for a pilot or production deployment.

## Vulnerability reporting

- Public non-sensitive issues: open a GitHub issue.
- Confidential vulnerability reports: use GitHub private vulnerability reporting for this repository when the
  **Security** tab offers it. If that flow is unavailable, open a minimal public issue titled `Security contact
  requested` with no sensitive details and ask for a private channel.
- Do not paste secrets, customer data, exploit payloads, or private logs into public issues.

## Baseline posture

| Area | Current posture |
| --- | --- |
| Auth | JWT, API key, and local dev-header modes are represented in config and middleware. |
| Tenancy | Tenant claims or tenant headers scope serving/admin routes. Production should enforce a single tenant source. |
| Admin access | Admin routes require configured admin roles when auth is enabled. |
| Rate limits | Global and per-tenant rate limit controls exist in service config. |
| Audit | Admin audit logging can be enabled with `AUDIT_LOG_ENABLED`. |
| Exposure data | Exposure logging can hash sensitive values with a production salt. |
| Pprof | Config validation restricts pprof to loopback bindings. |
| Artifacts | Production config rejects insecure S3 artifact mode when S3 endpoint use is configured. |

## EU-baseline privacy guidance

- Use pseudonymous user and session IDs.
- Avoid direct PII in request payloads, context fields, logs, artifacts, and evaluation datasets.
- Document retention for exposure, outcome, and audit logs before launch.
- Treat exported reports and datasets as sensitive operational data.
- Record subprocessors and hosting responsibilities in customer-specific deployment documentation.

## Pre-production security checks

```bash
make security
make docs-check
```

Expected result: Go vulnerability/static security scans pass for modules, docs build and link checks pass, and any
remaining production exceptions are documented before release.

## Limits

This page is not a compliance certification. It describes repository controls and operator responsibilities. A specific
deployment still needs environment-specific review for network policy, identity provider configuration, secret storage,
backup/retention, and incident response.
