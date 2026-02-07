# recsys-algo

Deterministic recommendation engine with **explainable scoring**, optional personalization, and merchandising rules.

This module is the **ranking core** of the suite. It consumes candidate sets (popularity, co-visitation,
similarity, etc.), applies constraints/rules, and produces a ranked list with optional explain/trace details.

If you are here as a recommendation engineer, start with:
[RecSys engineering: start here](../start-here/receng.md)

## Start here

- **Concepts:** [`concepts.md`](concepts.md)
- **Integration (store ports):** [`store-ports.md`](store-ports.md)
- **Examples:** [`examples.md`](examples.md)
- **Releases:** [`releases.md`](releases.md)

## Where this fits

- **recsys-service** calls into `recsys-algo` to generate ranked outputs.
- **recsys-pipelines** produces artifacts/signals that the service exposes as stores.
- **recsys-eval** validates changes in ranking behavior and business KPIs.
