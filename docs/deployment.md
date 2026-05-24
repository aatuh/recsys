# Deployment

Use the Helm chart as a starting point for Kubernetes deployments. The default chart values are local/demo oriented.
Production deployments should use external Postgres and object storage, production secrets, resource requests, readiness
checks, and a manifest rollback process.

## Production values

Start from the hardened example:

```bash
helm template recsys charts/recsys -f charts/recsys/values.production.example.yaml
```

Expected result: Helm renders the API deployment, optional pipeline CronJob, ConfigMaps, and references to the external
secret named by `api.existingSecret`.

The example intentionally does not contain real secrets. Create the referenced Kubernetes secret through your secret
manager or release pipeline. At minimum it should provide:

- `DATABASE_URL`
- `API_KEY_HASH_SECRET` when API-key auth is enabled
- `EXPOSURE_HASH_SALT` when exposure logging is enabled
- `EXPERIMENT_ASSIGNMENT_SALT` when experiment assignment is enabled
- `RECSYS_ARTIFACT_S3_ACCESS_KEY` and `RECSYS_ARTIFACT_S3_SECRET_KEY` when object-store credentials are not supplied by
  workload identity

## Rollback

Use two rollback levers:

- Service release rollback: deploy the previous API or pipeline image tag.
- Artifact rollback: restore the previous known-good current manifest for the affected tenant and surface.

Keep the previous manifest available until readiness, recommendation quality, and guardrail metrics have recovered.
