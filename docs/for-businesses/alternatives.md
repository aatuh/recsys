---
diataxis: explanation
tags:
  - business
  - build-vs-buy
  - alternatives
---
# Alternatives and build vs buy
This page explains Alternatives and build vs buy and how it fits into the RecSys suite.


## Who this is for

- Buyers comparing approaches
- Developers evaluating long-term ownership and risk

## What you will get

- A decision framework for common alternatives
- What RecSys is optimized for (and what it is not)

## Common alternatives

### 1) Build in-house

Pros:

- Full control over roadmap and implementation details

Costs / risks:

- Operational complexity (freshness, rollback, on-call)
- Hard-to-audit behavior if ranking is not deterministic
- Evaluation workload (offline metrics + governance)

RecSys reduces these costs by providing:

- Deterministic serving + explicit rules/constraints
- Versioned artifacts + rollback patterns
- Evaluation modules and decision playbooks

### 2) Use a managed black-box recommender

Pros:

- Fast initial “something works”

Costs / risks:

- Lower auditability (“why did we show this?”)
- Harder to combine product constraints, merchandising, and explainability
- Vendor coupling in data formats and model behavior

RecSys is optimized for teams that need:

- Explicit control knobs (rules + weights)
- Deterministic behavior and operational predictability

### 3) Use a simple heuristic (popularity only)

Pros:

- Extremely simple and reliable

Costs / risks:

- Limited personalization and limited long-term uplift
- Hard to evolve into a full evaluation + experimentation program

RecSys supports incremental adoption:

- Start with DB-only popularity and rules
- Move to production-like pipelines and richer signals later

## When RecSys is the wrong fit

RecSys is not optimized for:

- Teams that want a fully managed “hands-off” black box
- Use cases requiring deep real-time model training inside the serving path
- Organizations unwilling to integrate exposure logging (evaluation readiness)

## Read next

- Value model (ROI): [Value model (ROI template)](value-model.md)
- Evidence and trust signals: [Evidence (what “good outputs” look like)](evidence.md)
- Procurement pack: [Procurement pack (Security, Legal, IT, Finance)](procurement-pack.md)
