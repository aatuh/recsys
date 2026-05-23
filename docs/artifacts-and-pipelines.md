# Artifacts and Pipelines

This page explains how `recsys-pipelines` publishes versioned recommendation artifacts and how the service consumes the
current manifest in artifact mode.

## Mental model

Artifact mode separates offline computation from online serving:

1. `recsys-pipelines` reads exposure events for a tenant, surface, segment, and day window.
2. It validates and canonicalizes the input.
3. It computes artifacts such as popularity, co-occurrence, implicit, content similarity, and session sequence signals.
4. It writes immutable artifact blobs to object storage.
5. It updates the current manifest for the tenant and surface.
6. `recsys-service` reads the manifest and referenced artifacts after cache expiry or cache invalidation.

In this mode, the thing you ship and roll back is the current manifest, not a service binary.

## State locations

| State | Owner | Notes |
| --- | --- | --- |
| Tenant config and rules | `recsys-service` database tables | Versioned control-plane documents with ETags. |
| Canonical pipeline data | `recsys-pipelines` configured store | Used to build deterministic daily artifacts. |
| Artifact blobs | Pipeline object store | Immutable payloads keyed by artifact type and version. |
| Current manifest | Pipeline registry | Mutable pointer under `current/<tenant>/<surface>/manifest.json` in local registry mode. |
| Exposure/outcome logs | Operator logging pipeline | Inputs for evaluation and incident reconstruction. |

## Local pipeline command

The repository proof-kit path runs the pipeline against a checked-in ecommerce fixture:

```bash
make proof-kit-test
```

Expected result: the pipeline writes a manifest under
`tmp/commercial-proof-kit/pipelines/registry/current/demo/home/manifest.json` and publishes artifact blobs under
`tmp/commercial-proof-kit/pipelines/objectstore/`.

## Freshness and rollback

Monitor freshness per tenant and surface:

- manifest age
- artifact publish time
- empty recommendation rate
- signal warning rate
- pipeline job success/failure

Rollback in artifact mode means restoring the previous known-good manifest content at the current manifest path. Because
artifact blobs are immutable, rollback should not require recomputing artifacts if the previous blobs still exist.

## Backfills

Backfills reprocess historical windows. Treat them as controlled operations:

- choose the date window explicitly
- check configured backfill limits
- publish only after validation passes
- compare output size and key metrics to the previous run
- keep the previous manifest available until verification passes

## Service configuration

The service consumes artifacts when artifact mode is enabled:

```bash
RECSYS_ARTIFACT_MODE_ENABLED=true
RECSYS_ARTIFACT_MANIFEST_TEMPLATE=s3://recsys-artifacts/registry/current/{tenant}/{surface}/manifest.json
RECSYS_ARTIFACT_MANIFEST_TTL=1m
RECSYS_ARTIFACT_CACHE_TTL=1m
```

Use production-safe object-store credentials, TLS settings, and cache TTLs before routing real traffic.

## Read next

- [Configuration Reference](reference/config.md)
- [Stale Artifact Manifest](operations/runbooks/stale-artifact-manifest.md)
- [Evaluation Decisions](evaluation-decisions.md)
- [Data Contracts](reference/data-contracts.md)
