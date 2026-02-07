---
tags:
  - how-to
  - overview
  - developer
  - ops
---

# How-to guides

## Who this is for

- Engineers integrating RecSys into an application
- Operators running the suite in production and on-call

## What you will get

- Goal-oriented steps (prereqs → steps → verify)
- Operational runbooks for common failure modes

--8<-- "_snippets/key-terms.list.snippet"
--8<-- "_snippets/key-terms.defs.one-up.snippet"

## Common tasks

<div class="grid cards" markdown>

- **[First surface end-to-end](first-surface-end-to-end.md)**  
  One stitched path: choose mode → integrate API → instrument → run first eval.
- **[Integrate serving API](integrate-recsys-service.md)**  
  Tenancy, auth, request/response, and production integration notes.
- **[Integration checklist](integration-checklist.md)**  
  A compact “did we do the basics?” list for app teams.
- **[Run eval & ship decisions](run-eval-and-ship.md)**  
  Produce reports and decide ship/hold/rollback with evidence.
- **[Deploy with Helm](deploy-helm.md)**  
  Kubernetes deployment guide and production checks.
- **[Operations](../operations/index.md)**  
  Production readiness, performance, and on-call runbooks.

</div>

## Read next

- Reference (contracts + config): [`reference/index.md`](../reference/index.md)
- Explanation (mental model): [`explanation/index.md`](../explanation/index.md)
