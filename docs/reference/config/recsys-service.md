# recsys-service configuration (lean)

- db.dsn: Postgres DSN
- auth.required: enable auth on protected routes
- auth.tenant_claim: claim used for tenant id (JWT mode)
- auth.admin_role: role required for /v1/admin
- auth.dev_headers: enable dev headers (local)
- auth.dev_tenant_header: tenant header used with dev auth
- limits.rps_per_tenant: per-tenant rate limit
- cache.manifest_ttl_seconds: manifest cache TTL
- exposure.log_path: file path (JSONL)
- exposure.log_format: service_v1 | eval_v1
