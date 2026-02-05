---
tags:
  - explanation
  - ops
  - pipelines
  - recsys-pipelines
---

# Pipelines operational invariants (safety model)

## Who this is for

- SRE/on-call and platform engineers operating `recsys-pipelines`
- Lead developers designing a safe ship/rollback workflow
- Data engineers who need to reason about retries and partial failures

## What you will get

- The invariants the pipelines try to maintain (and what is not guaranteed)
- What “idempotent publish” and “atomic ship” mean in practice
- The failure modes that matter operationally, with safe recovery actions

## Invariant 1: artifacts are immutable and version-addressed

An **artifact** is published under a version (hash) and treated as immutable:

- the same canonical payload should produce the same version
- older artifact versions remain readable after new publishes

Operational consequence: you can cache aggressively and roll back safely because “old” data still exists.

## Invariant 2: publish is two-phase and swaps the manifest last

Publishing is structured so serving never reads a half-updated set:

1. Write the versioned artifact blob
2. Validate the artifact (including recomputing the version)
3. Write a registry record for audit/rollback
4. **Swap the current manifest pointer last**

Operational consequence: if a publish fails before the final swap, “current” serving stays on the previous manifest.

## Invariant 3: re-running publish is intended to be safe (idempotent)

Rerunning a pipeline for the same `(tenant, surface, window)` should not corrupt “current”:

- versioned objects are written under stable keys (`…/<type>/<version>.json`)
- registry records are append-only (re-recording an existing version is a no-op)
- the manifest swap is a whole-document replace, not an in-place patch

Operational consequence: retrying is the default remediation for transient failures.

## Invariant 4: rollback is a manifest pointer change

Rollback is switching the current manifest back to a previous version.

Operational consequence: rollback is fast and does not require re-computing artifacts.

## What is not guaranteed (plan for it)

- **No built-in concurrency control**: two publishes to the same `(tenant, surface)` can race; last manifest swap wins.
- **No automatic garbage collection**: failed publishes can leave unreferenced artifacts (“orphans”) in storage.
- **Eventual visibility in serving**: `recsys-service` may cache manifests/artifacts; changes apply after TTL or explicit
  cache invalidation.

## Safe recovery patterns

- If publish failed before manifest swap: fix the cause and retry (serving stayed on the previous manifest).
- If you shipped a bad manifest: roll back the manifest pointer (do not delete artifacts).
- If you see stale serving: invalidate caches or wait for TTL; then verify the service is reading the expected manifest.

## Read next

- Artifacts + manifest lifecycle: [`explanation/artifacts-and-manifest-lifecycle.md`](artifacts-and-manifest-lifecycle.md)
- Pipelines output layout (registry paths): [`recsys-pipelines/docs/reference/output-layout.md`](../recsys-pipelines/docs/reference/output-layout.md)
- Roll back the manifest: [`recsys-pipelines/docs/how-to/rollback-manifest.md`](../recsys-pipelines/docs/how-to/rollback-manifest.md)
