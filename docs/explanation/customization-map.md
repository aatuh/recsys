---
diataxis: explanation
tags:
  - explanation
  - customization
  - integration
  - operations
---
# Customization map
This page explains Customization map and how it fits into the RecSys suite.


## Who this is for

- Developers, recsys engineers, and stakeholders who need to understand **what can be changed where**.

## What you will get

- A practical “lever map” (config/rules/data/code) that reduces page hunting
- Links to the canonical pages for each lever

## The four customization layers

### Layer 1: Tenant config (fast, safe)

What you can change:

- Signal weights and blending knobs
- Limits (max K, exclude sizes)
- Feature flags that gate behaviors

Where:

- Admin config endpoints: [Admin API + local bootstrap (recsys-service)](../reference/api/admin.md)
- Config schema: [recsys-service configuration](../reference/config/recsys-service.md)

When to use:

- Tuning and controlled rollout of a known behavior

### Layer 2: Merchandising rules (fast, explicit)

What you can change:

- Pin / boost / block and other constraints
- Surface-specific policies

Where:

- Integration guide (rules): [How-to: integrate recsys-service into an application](../how-to/integrate-recsys-service.md)

When to use:

- Business constraints, promotions, and safety guardrails

### Layer 3: Data / signals (medium, production-like)

What you can change:

- Candidate signals such as popularity/co-occurrence, plus new signals
- Artifact freshness and validation rules

Where:

- Add a signal end-to-end: [How-to: add a new signal end-to-end](../how-to/add-signal-end-to-end.md)
- Pipelines operation: [How-to: operate recsys-pipelines](../how-to/operate-pipelines.md)
- Artifacts lifecycle: [Artifacts and manifest lifecycle (pipelines → service)](artifacts-and-manifest-lifecycle.md)

When to use:

- Improving ranking by improving candidate sources or signal quality

### Layer 4: Ranking logic (high leverage, requires evaluation)

What you can change:

- Candidate merge, scoring, tie-break policies
- Explain payload fields and deterministic constraints

Where:

- Ranking reference: [Ranking & constraints reference](../recsys-algo/ranking-reference.md)
- Scoring model spec: [Scoring model specification (recsys-algo)](../recsys-algo/scoring-model.md)

When to use:

- Structural ranking improvements after you have evaluation gates in place

## “If I want X, which layer is it?”

- “I want to pin items for a campaign” → Layer 2
- “I want to increase diversity” → Layer 1 or 4 (depending on the lever)
- “I want better candidates for cold start” → Layer 3
- “I want a new scoring feature” → Layer 4

## Read next

- RecSys engineering hub: [RecSys engineering hub](../recsys-engineering/index.md)
- Evaluation modes: [Evaluation modes](evaluation-modes.md)
- Run eval and ship: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)
