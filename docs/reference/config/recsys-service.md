# recsys-service configuration

Canonical env var list: `recsys/api/.env.example`.

- db.dsn: Postgres DSN
- auth.required: enable auth on protected routes
- auth.tenant_claim: claim used for tenant id (JWT mode)
- auth.viewer_role: role that can read admin resources (default: viewer)
- auth.operator_role: role that can mutate admin resources (default: operator)
- auth.admin_role: role with full admin access (default: admin)
- auth.dev_headers: enable dev headers (local)
- auth.dev_tenant_header: tenant header used with dev auth
- limits.rps_per_tenant: per-tenant rate limit
- audit.log_path: file path for optional audit JSONL log
- audit.log_fsync: fsync on each audit write (default: false)
- cache.config_ttl_seconds: config cache TTL
- cache.rules_ttl_seconds: rules cache TTL
- exposure.log_path: file path (JSONL)
- exposure.log_format: service_v1 | eval_v1
- algo.mode: blend | popularity | cooc | implicit | content_sim | session_seq (default: blend)
- algo.plugin_enabled: enable Go plugin loading (dev only)
- algo.plugin_path: filesystem path to .so plugin (dev only)
- artifacts.enabled: enable artifact/manifest mode
- artifacts.manifest_template: manifest URI template (supports {tenant} and {surface})
- artifacts.manifest_ttl_seconds: manifest cache TTL
- artifacts.cache_ttl_seconds: artifact cache TTL
- artifacts.max_bytes: max artifact size in bytes
- artifacts.s3.endpoint: S3/MinIO endpoint (host:port)
- artifacts.s3.access_key: S3 access key
- artifacts.s3.secret_key: S3 secret key
- artifacts.s3.region: S3 region
- artifacts.s3.use_ssl: use TLS for S3 (true/false)

Notes:

- content_sim and session_seq modes require corresponding artifacts in the manifest.
- auth.tenant_claim and auth.role_claims support dotted keys (e.g., realm_access.roles).
