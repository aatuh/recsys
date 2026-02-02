
# How-to: Add a new field to exposure events

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
