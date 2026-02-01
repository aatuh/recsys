
# How-to: Add a new field to exposure events

## Rules

- Keep old readers working (backwards compatible)
- Do not reuse field meanings
- Update schema and examples

## Steps

1) Update JSON schema: `schemas/events/exposure.v1.json`
2) Update domain event struct: `internal/domain/events/exposure.go`
3) Update raw event decoder if needed
4) Update canonical writer/reader tests
5) Update docs: `reference/schemas-events.md`
