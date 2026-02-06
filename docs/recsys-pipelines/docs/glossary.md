# Glossary

This page defines `recsys-pipelines`-specific terms.

For shared suite terminology, use the suite glossary:

- Artifact: [`project/glossary.md#artifact`](../../project/glossary.md#artifact)
- Manifest: [`project/glossary.md#manifest`](../../project/glossary.md#manifest)
- Tenant: [`project/glossary.md#tenant`](../../project/glossary.md#tenant)
- Surface: [`project/glossary.md#surface`](../../project/glossary.md#surface)
- Segment: [`project/glossary.md#segment`](../../project/glossary.md#segment)

## Pipeline-specific terms

**Canonical events**
: Events stored in a normalized format that the rest of the pipeline relies on.

**Window**
: A time range that a job processes. In v1, windows are daily UTC buckets.

**Version**
: A deterministic identifier (SHA-256 hex) of an artifact payload excluding
  volatile build metadata.

**Checkpoint**
: A small state file that records the latest successfully processed window so
  incremental runs can skip work already done.

**Incremental run**
: A run mode that processes only new windows since the last checkpoint
  (see `how-to/run-incremental.md`).

**Backfill**
: Re-processing a historical range of windows (see `how-to/run-backfill.md`).

**Current manifest pointer**
: The mutable "what is live right now" location for a `(tenant, surface)`
  manifest (for example: `.out/registry/current/<tenant>/<surface>/manifest.json`).

**Registry**
: Storage for artifact records and current manifests.

**Object store**
: Storage for artifact blobs. In local mode, this is the filesystem.

**Idempotent**
: Safe to run multiple times without changing the result.

## Read next

- Start here: [`start-here.md`](start-here.md)
- Artifacts and versioning: [`explanation/artifacts-and-versioning.md`](explanation/artifacts-and-versioning.md)
- Output layout (registry and object store): [`reference/output-layout.md`](reference/output-layout.md)
- Roll back artifacts safely: [`how-to/rollback-safely.md`](how-to/rollback-safely.md)
