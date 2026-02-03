
# Learning path: Data Engineering

## Goals

- Understand event schemas and file layouts
- Run backfills safely
- Operate data quality gates
- Define evolution rules for new fields

## Read in this order

1) `reference/schemas-events.md`
1) `explanation/data-lifecycle.md`
1) `how-to/run-backfill.md`
1) `how-to/add-event-field.md`
1) `reference/output-layout.md`

## Key practical advice

- Treat canonical events as the contract boundary.
- Keep schema evolution backwards compatible.
- Always validate before publishing.
