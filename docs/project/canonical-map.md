---
diataxis: explanation
tags:
  - project
  - docs
  - ia
---
# Canonical content map

This page defines **where the single source of truth lives** for topics that tend to get duplicated.

Use it when writing new pages or reviewing PRs.

## Rules

- Each topic has **one canonical page**.
- Other pages may link to the canonical page, but must not re-explain it.
- Duplicate content is allowed only for:
  - very short summaries (2–3 bullets)
  - “role-based” re-framing (business vs dev) that points to the canonical page
  - checklists that link to the canonical explanation/reference

## Canonical pages

### Product positioning / buyer flow

- “What is RecSys?”: [What the RecSys suite is (stakeholder overview)](../start-here/what-is-recsys.md)
- Buyer journey: [Buyer journey: evaluate RecSys in 5 minutes](../for-businesses/buyer-journey.md)
- Evaluation + licensing: [Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)

### Integration contracts

- Integration contract (headers, tenancy, request_id, invariants): [Integration spec (one surface)](../reference/integration-spec.md)
- Data contracts hub: [Data contracts](../reference/data-contracts/index.md)
- Minimum instrumentation spec: [Minimum instrumentation spec (for credible evaluation)](../reference/minimum-instrumentation.md)

### Evaluation and decisions

- Suite workflow (ship/hold/rollback): [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)
- Evaluation validity: [Evaluation validity](../explanation/eval-validity.md)
- Evaluation modes: [Evaluation modes](../explanation/evaluation-modes.md)

### Operations

- Operational reliability and rollback: [Operational reliability and rollback](../start-here/operational-reliability-and-rollback.md)
- Runbooks hub: [Operations](../operations/index.md)
- Baseline benchmarks (ops): [Baseline benchmarks (anchor numbers)](../operations/baseline-benchmarks.md)

### Suite architecture

- Architecture and data flow (canonical): [How it works: architecture and data flow](../explanation/how-it-works.md)
- Component boundaries and responsibilities: [Suite architecture](../explanation/suite-architecture.md)
- Data modes: [Data modes: DB-only vs artifact/manifest](../explanation/data-modes.md)
- Artifacts + manifest lifecycle: [Artifacts and manifest lifecycle (pipelines → service)](../explanation/artifacts-and-manifest-lifecycle.md)

## Read next

- Docs style: [Documentation style guide](docs-style.md)
- How to contribute: [Contributing](contributing.md)
