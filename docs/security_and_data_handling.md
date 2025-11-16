# Security & Data Handling

This document summarizes the basic security expectations when integrating with RecSys. Use it to answer high-level questions before looping in ops/security teams.

## Transport security

- Hosted environments expose the API over HTTPS/TLS. Always use `https://` endpoints in production.
- Local developer stacks (`http://localhost:8000`) are for sandbox use only.
- Load balancers terminate TLS and forward requests to the API container. Client certificates are not required by default; API key auth (see below) protects the application layer.

## Authentication & multi-tenancy

- **Org header:** Every request must include `X-Org-ID: <uuid>`. This enforces tenant isolation. Missing or incorrect headers are rejected with `400` or `401` responses.
- **API keys:** When `API_AUTH_ENABLED=true`, include `X-API-Key: <token>` or the configured `Authorization` header. Keys are managed per org/tenant.
- **Namespacing:** Payloads include a `namespace` field. Namespaces isolate catalogs, rules, guardrails, and audit logs. Use separate namespaces for different surfaces/customers to avoid cross-tenant leakage.
- **Permissions:** Admin endpoints (`/v1/admin/*`) require org + namespace scoping. We recommend issuing API keys with least privilege (e.g., read-only vs admin).

## PII guidance

- Only send the information needed to power recommendations. Fields like `user_id`, `traits`, and `events.meta` should avoid raw PII whenever possible (use hashed IDs or pseudonyms).
- Do not send payment data, addresses, or sensitive attributes via RecSys; the platform is not a general-purpose PII store.
- If you need to include attributes that could be considered personal data (e.g., loyalty tier, location bucket), ensure they align with your company’s privacy policy.

## Data retention & deletion

- **Items/users/events:** Use the corresponding `*:delete` endpoints to remove records when they should no longer participate in recommendations. Deletions cascade through guardrails and are honored by `/v1/recommendations` once caches refresh.
- **Decision traces:** Guardrail and audit traces are stored in `rec_decisions`. Retention defaults to 30 days; adjust via database/ops tooling if stricter policies apply.
- **Manual overrides/rules:** Use `/v1/admin/rules/{id}` and `/v1/admin/manual_overrides/{id}/cancel` to remove or expire rules.
- RecSys does not automatically purge tenant data; work with your ops team to schedule periodic cleanups if needed.

## Incident response & auditability

- Every recommendation response includes a `trace_id`; log it to correlate client-side issues with server-side audits.
- `/v1/audit/decisions` exposes the full recorded request + config + response for governance teams. Restrict access to this endpoint if it contains sensitive metadata.
- Guardrails and Prometheus dashboards (`policy_guardrail_failures_total`, `policy_rule_zero_effect_total`) are the first line of defense for catching anomalies.

## Compliance posture

- RecSys is designed to run inside your infrastructure (Docker/Kubernetes). It inherits your organization’s compliance posture (SOC2, GDPR, etc.).
- Security-sensitive features (TLS termination, key management, database encryption) are managed at the deployment level; consult your DevOps/SRE team for environment-specific controls.
