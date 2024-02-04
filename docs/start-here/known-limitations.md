# Known limitations and non-goals (current)

This page is intentionally blunt. It exists so evaluators and operators can set expectations early.

## Who this is for

- Evaluators deciding whether this suite fits their current stage
- Lead developers planning adoption and integration scope
- Operators who need to avoid “surprise missing feature” incidents

## What you will get

- The current product boundaries: what is supported vs intentionally not included
- Concrete limitations that affect pilots and production rollouts
- Pointers to the relevant docs for workarounds and operational safety

## Limitations

- **Tenant creation is DB-only today.** There is no admin API to create tenants yet; bootstrap requires a SQL insert.
  See: [`reference/api/admin.md`](../reference/api/admin.md)

- **Pipelines manifest registry is filesystem-based.** `recsys-pipelines` writes the “current manifest pointer” to its
  configured `registry_dir` on the filesystem. Artifacts can be published to S3/MinIO, but publishing the manifest to
  object storage requires an explicit upload step (or serving from a file-based manifest template).

- **Kafka ingestion is scaffolded.** The `raw_source.type = kafka` connector is present as a config option but is not
  implemented as a streaming consumer yet.

- **Dev headers don’t carry roles.** If you enable admin RBAC roles, admin endpoints require roles from JWT/API keys.
  For local dev with dev headers, either disable RBAC roles or use JWT.

## Non-goals (by default)

- Running your infrastructure for you (managed hosting is not implied by this repo)
- “Magic” model training inside the serving stack (the suite is designed for deterministic, auditable behavior first)

## Read next

- Pilot plan (what to do first): [`start-here/pilot-plan.md`](pilot-plan.md)
- Data modes (DB-only vs artifact/manifest): [`explanation/data-modes.md`](../explanation/data-modes.md)
- Local end-to-end tutorial (known-good baseline): [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)
