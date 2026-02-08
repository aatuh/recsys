---
diataxis: how-to
tags:
  - recsys-pipelines
---
# How-to: Add a new field to exposure events
This guide shows how to how-to: Add a new field to exposure events in a reliable, repeatable way.


## Rules

- Keep old readers working (backwards compatible)
- Do not reuse field meanings
- Update schema and examples

## Steps

1) Update JSON schema: `schemas/events/exposure.v1.json`
1) Update domain event struct: `internal/domain/events/exposure.go`
1) Update raw event decoder if needed
1) Update canonical writer/reader tests
1) Update docs: `reference/schemas-events.md`

## Read next

- Start here: [Start here](../start-here.md)
- Event schemas: [Event schemas](../reference/schemas-events.md)
- Data lifecycle: [Data lifecycle](../explanation/data-lifecycle.md)
- Validation and guardrails: [Validation and guardrails](../explanation/validation-and-guardrails.md)
