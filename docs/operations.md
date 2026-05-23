# Operations

This page gives the first operational checks for local development, pilot deployments, and production readiness.

## Health and readiness

| Endpoint | Meaning |
| --- | --- |
| `/healthz` | Process liveness. Use this for "is the service up?" checks. |
| `/readyz` | Dependency readiness. Use this before routing traffic. |
| `/metrics` | Prometheus metrics when the service is running with the default toolkit middleware. |

```bash
curl -f http://localhost:8000/healthz
curl -f http://localhost:8000/readyz
```

## Production readiness checklist

- Auth is enforced with JWT or API keys; dev headers are disabled.
- Tenant source is explicit and tested for cross-tenant isolation.
- `EXPOSURE_HASH_SALT`, `EXPERIMENT_ASSIGNMENT_SALT`, and `API_KEY_HASH_SECRET` are set when their production features
  are enabled.
- CORS allows only expected browser origins.
- Artifact mode has a rollback path and a manifest TTL that matches operational needs.
- Exposure/outcome data retention is documented.
- Logs preserve request IDs and avoid raw PII.
- `make docs-check` and the module quality gates are green before release.

## Rollback levers

| Change | Rollback lever |
| --- | --- |
| Tenant config | Reapply the previous config version through admin config routes. |
| Rules | Reapply previous rules or disable rules with `RECSYS_ALGO_RULES_ENABLED=false`. |
| Artifact manifest | Point the manifest to the last known-good artifact version. |
| Algorithm plugin | Disable `RECSYS_ALGO_PLUGIN_ENABLED` or revert `RECSYS_ALGO_PLUGIN_PATH`. |
| Service release | Roll back the container image or binary to the previous release. |

## Empty recommendations

First checks:

1. Confirm the request has the expected tenant and surface.
2. Validate the request with `/v1/recommend/validate`.
3. Check whether candidate include/exclude lists removed all items.
4. Check tenant config, rules, artifact manifest, and artifact load errors.
5. Review service logs using the request ID.

## Stale manifest

First checks:

1. Confirm the configured `RECSYS_ARTIFACT_MANIFEST_TEMPLATE` resolves to the expected tenant and surface.
2. Check object-store reachability and credentials.
3. Confirm the manifest `updated_at` and artifact paths.
4. Invalidate relevant caches through admin cache invalidation when the new manifest is known-good.

## Service not ready

First checks:

1. Inspect Compose or orchestrator logs.
2. Check database connectivity and migration status.
3. Confirm production-only config validation is not failing on missing secrets or unsafe S3 SSL settings.
4. Check `/healthz` separately from `/readyz` to distinguish process liveness from dependency readiness.

```bash
docker compose logs --tail=100 api
cd api && make migrate-status
```

Expected result: logs and migration status identify whether the failure is config, database, migrations, or service
startup.
