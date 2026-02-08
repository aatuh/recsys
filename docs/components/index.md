---
diataxis: explanation
tags:
  - components
  - overview
  - developer
  - ops
---
# Components

This section contains **deep, component-level documentation**. If you’re new,
start with:

- Start here: [Start here](../start-here/index.md)
- Tutorials: [Tutorials](../tutorials/index.md)

## Suite modules

<div class="grid cards" markdown>

- **[recsys-service (serving API)](../how-to/integrate-recsys-service.md)**  
  Integration guide + the canonical API reference.
- **[recsys-algo (ranking logic)](../recsys-algo/index.md)**  
  Signals, blending, constraints, and determinism notes.
- **[recsys-pipelines (offline layer)](../recsys-pipelines/docs/index.md)**  
  Artifacts, scheduling, backfills, and operations runbooks.
- **[recsys-eval (evaluation)](../recsys-eval/docs/index.md)**  
  Metrics, workflows, CI gates, and experiment analysis.

</div>

!!! info "Layering and contracts"
    The suite is intentionally layered to keep ownership boundaries clear:

    - `recsys-algo` is the ranking core (no HTTP, no DB).
    - `recsys-service` is serving + tenancy + rules/config.
    - `recsys-pipelines` produces versioned artifacts and a manifest (optional in DB-only mode).
    - `recsys-eval` consumes logs to produce evaluation reports.

    If you're choosing an integration shape, start with: [Choose your data mode](../start-here/choose-data-mode.md)

## Read next

- API reference: [API Reference](../reference/api/api-reference.md)
- How it works (mental model): [How it works: architecture and data flow](../explanation/how-it-works.md)
