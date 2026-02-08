---
diataxis: how-to
tags:
  - how-to
  - overview
  - developer
  - ops
---
# How-to guides
Task-focused guides (prereqs → steps → verify) for integrating, operating, and tuning RecSys.


## Who this is for

- Engineers integrating RecSys into an application
- Operators running the suite in production and on-call

## What you will get

- Goal-oriented steps (prereqs → steps → verify)
- Operational runbooks for common failure modes

## Common tasks

<div class="grid cards" markdown>

- **[Integrate serving API](integrate-recsys-service.md)**  
  Tenancy, auth, request/response, and production integration notes.
- **[Integration checklist](integration-checklist.md)**  
  A compact “did we do the basics?” list for app teams.
- **[Troubleshooting for integrators](troubleshooting-integration.md)**  
  Symptom → cause → fix checklist for empty recs, auth issues, and join problems.
- **[Run eval & ship decisions](run-eval-and-ship.md)**  
  Produce reports and decide ship/hold/rollback with evidence.
- **[Validate logging & joinability](validate-logging.md)**  
  Confirm exposures and outcomes are attributable before you trust metrics.
- **[Tune ranking safely](tune-ranking.md)**  
  Change ranking behavior without losing auditability or rollback discipline.
- **[Add a signal end-to-end](add-signal-end-to-end.md)**  
  Extend ranking inputs safely and validate the impact.
- **[Deploy with Helm](deploy-helm.md)**  
  Kubernetes deployment guide and production checks.
- **[Operations](../operations/index.md)**  
  Production readiness, performance, and on-call runbooks.

</div>

## Read next

- Tutorials (happy-path walkthroughs): [Tutorials](../tutorials/index.md)
- Reference (contracts + config): [Reference](../reference/index.md)
- Explanation (mental model): [Explanation](../explanation/index.md)
