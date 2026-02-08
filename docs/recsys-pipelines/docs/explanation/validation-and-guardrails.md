---
diataxis: explanation
tags:
  - recsys-pipelines
---
# Validation and guardrails
This page explains Validation and guardrails and how it fits into the RecSys suite.


## Validation gate

The pipeline validates canonical data before computing/publishing.

Builtin checks include:

- event parsing and required fields
- timestamp inside the window
- maximum events processed
- maximum distinct sessions/items

Artifacts are also validated:

- correct type/version/window
- version matches recomputed hash
- maximum sizes (items/neighbors)

## Guardrails

Resource limits protect the pipeline from unbounded inputs:

- max events per run
- max sessions per run
- max items per session
- max distinct items per run
- max neighbors per item
- max items per artifact

If you see "limit exceeded", raise limits only after understanding why.

Operational guidance: `operations/runbooks/limit-exceeded.md`.

## Read next

- Limit exceeded runbook: [Runbook: Limit exceeded](../operations/runbooks/limit-exceeded.md)
- Config reference (limits and inputs): [Config reference](../reference/config.md)
- Debug failures: [How-to: Debug a failed pipeline run](../how-to/debug-failures.md)
- Operate pipelines daily: [How-to: Operate pipelines daily](../how-to/operate-daily.md)
- Glossary: [Glossary](../glossary.md)
