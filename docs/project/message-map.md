---
diataxis: reference
tags:
  - project
---
# Message map (maintainers)

This page helps keep **positioning and wording consistent** across the docs, README, and commercial pages.

Wedge statement (use verbatim):

- RecSys is an auditable recommendation system suite with deterministic ranking and versioned ship/rollback.

## Core promises → proof points

1. **Auditable decisions**
   - Proof: exposure/outcome logs join by `request_id`; evaluation reports produce ship/hold/rollback artifacts.
   - Docs: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)
1. **Deterministic ranking (predictable behavior)**
   - Proof: deterministic ranking core; explicit constraints/rules; versioned artifacts.
   - Docs: [Candidate generation vs ranking](../explanation/candidate-vs-ranking.md)
1. **Safe shipping and rollback**
   - Proof: versioned config/rules/manifests; runbooks for common failures.
   - Docs: [Operational reliability and rollback](../start-here/operational-reliability-and-rollback.md)

## Common objections → short answers

- **“No case studies yet”** → Start with a measurable pilot surface and ship a report you can defend.
- **“Operational burden”** → We document day-2 runbooks and rollback levers; start DB-only, graduate to artifacts.
- **“Security review”** → Self-hosted by default; no raw PII required; publish a trust center and hardening checklist.

## Where to use this

- Homepage (`docs/index.md`), hub pages (`docs/developers/index.md`, `docs/for-businesses/index.md`)
- Pricing and evaluation pages (`docs/pricing/index.md`, `docs/evaluate/index.md`)
- README (`README.md`)

## Read next

- Project hub: [Project](index.md)
- Contributing: [Contributing](contributing.md)
