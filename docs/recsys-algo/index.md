---
diataxis: explanation
tags:
  - recsys-algo
  - ranking
  - explanation
---
# recsys-algo

Deterministic recommendation engine with **explainable scoring**, optional personalization, and merchandising rules.

This module is the **ranking core** of the suite. It consumes candidate sets (popularity, co-visitation,
similarity, etc.), applies constraints/rules, and produces a ranked list with optional explain/trace details.

## Start here

- **Concepts:** [Concepts](concepts.md)
- **Integration (store ports):** [Store ports](store-ports.md)
- **Examples:** [Examples](examples.md)
- **Releases:** [Releases](releases.md)

## Where this fits

- **recsys-service** calls into `recsys-algo` to generate ranked outputs.
- **recsys-pipelines** produces artifacts/signals that the service exposes as stores.
- **recsys-eval** validates changes in ranking behavior and business KPIs.

## Read next

- Start here: [Start here](../start-here/index.md)
- Quickstart (10 minutes): [Quickstart (10 minutes)](../tutorials/quickstart.md)
