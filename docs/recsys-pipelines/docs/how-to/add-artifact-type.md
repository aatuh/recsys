
# How-to: Add a new artifact type

This repo uses ports/adapters and a workflow pipeline.

## Checklist

1. Define the domain model

   - Add a new `artifacts.Type` value
   - Add a v1 model struct and constructor

1. Implement a compute usecase

   - IO via `datasource.CanonicalStore`
   - Deterministic version hash (exclude build info)

1. Update validation

   - Extend builtin validator: schema checks + version recompute

1. Wire into workflow

   - Add to `workflow.Pipeline.RunDay`
   - Add to job mode (compute job + publish job)

1. Add reference docs

   - Add schema under `schemas/artifacts/`
   - Update `reference/output-layout.md`

## Non-negotiables

- Deterministic output for same canonical inputs
- Bounded resource usage
- Publish pointer updated last

## Read next

- Suite how-to (signal end-to-end): [`how-to/add-signal-end-to-end.md`](../../../how-to/add-signal-end-to-end.md)
- Start here: [`start-here.md`](../start-here.md)
- Artifacts and versioning: [`explanation/artifacts-and-versioning.md`](../explanation/artifacts-and-versioning.md)
- Artifact schemas: [`reference/schemas-artifacts.md`](../reference/schemas-artifacts.md)
- Output layout: [`reference/output-layout.md`](../reference/output-layout.md)
