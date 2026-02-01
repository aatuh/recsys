# recsys-service configuration (lean)

Canonical env var list: `recsys/api/.env.example`.

- db.dsn: Postgres DSN
- auth.required: enable auth on protected routes
- auth.tenant_claim: claim used for tenant id (JWT mode)
- auth.admin_role: role required for /v1/admin
- auth.dev_headers: enable dev headers (local)
- auth.dev_tenant_header: tenant header used with dev auth
- limits.rps_per_tenant: per-tenant rate limit
- cache.config_ttl_seconds: config cache TTL
- cache.rules_ttl_seconds: rules cache TTL
- exposure.log_path: file path (JSONL)
- exposure.log_format: service_v1 | eval_v1
- algo.mode: blend | popularity | cooc | implicit (default: blend)
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
