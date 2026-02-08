---
diataxis: how-to
tags:
  - how-to
  - ops
  - rollback
  - artifacts
  - recsys-pipelines
---
# How-to: Roll back artifacts safely
This guide shows how to how-to: Roll back artifacts safely in a reliable, repeatable way.


## Who this is for

- SRE / on-call responding to a bad publish or stale/incorrect artifacts
- Data engineers who need a reversible release process

## Goal

Roll back the “current” manifest pointer to a known-good version with a clear verification step.

## Quick paths

- Roll back the manifest: [How-to: Roll back to a previous artifact version](rollback-manifest.md)
- Artifacts and versioning (concepts): [Artifacts and versioning](../explanation/artifacts-and-versioning.md)
- Output layout (where pointers live): [Output layout (local filesystem)](../reference/output-layout.md)
- Stale artifacts runbook: [Runbook: Stale artifacts](../operations/runbooks/stale-artifacts.md)

## Safety checklist

1. Confirm you are rolling back the right (tenant, surface, segment)

- Make the scope explicit in your command and verification steps.

1. Roll back the manifest pointer

- Use the canonical flow: [How-to: Roll back to a previous artifact version](rollback-manifest.md)

1. Verify the rollback

- Confirm “current” points to the expected version:
  [Output layout (local filesystem)](../reference/output-layout.md)

## Read next

- Artifacts and manifest lifecycle: [Artifacts and manifest lifecycle](../../../explanation/artifacts-and-manifest-lifecycle.md)
- Runbook: stale manifest: [Runbook: Stale manifest](../../../operations/runbooks/stale-manifest.md)
- Pipelines operations index: [Operations (recsys-pipelines)](../operations/slos-and-freshness.md)
