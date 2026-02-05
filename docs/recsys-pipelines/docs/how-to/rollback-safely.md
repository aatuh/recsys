---
tags:
  - how-to
  - ops
  - rollback
  - artifacts
  - recsys-pipelines
---

# How-to: Roll back artifacts safely

## Who this is for

- SRE / on-call responding to a bad publish or stale/incorrect artifacts
- Data engineers who need a reversible release process

## Goal

Roll back the “current” manifest pointer to a known-good version with a clear verification step.

## Quick paths

- Roll back the manifest: [`how-to/rollback-manifest.md`](rollback-manifest.md)
- Artifacts and versioning (concepts): [`explanation/artifacts-and-versioning.md`](../explanation/artifacts-and-versioning.md)
- Output layout (where pointers live): [`reference/output-layout.md`](../reference/output-layout.md)
- Stale artifacts runbook: [`operations/runbooks/stale-artifacts.md`](../operations/runbooks/stale-artifacts.md)

## Safety checklist

1. Confirm you are rolling back the right (tenant, surface, segment)

- Make the scope explicit in your command and verification steps.

1. Roll back the manifest pointer

- Use the canonical flow: [`how-to/rollback-manifest.md`](rollback-manifest.md)

1. Verify the rollback

- Confirm “current” points to the expected version:
  [`reference/output-layout.md`](../reference/output-layout.md)

## Read next

- Operate daily: [`how-to/operate-daily.md`](operate-daily.md)
- Pipeline failed runbook: [`operations/runbooks/pipeline-failed.md`](../operations/runbooks/pipeline-failed.md)
