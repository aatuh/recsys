---
diataxis: reference
tags:
  - quickstart
  - overview
  - developer
  - business
---
# Docs map
This page is the canonical reference for Docs map.


## Who this is for

- Anyone evaluating the RecSys suite for adoption
- Teams onboarding new engineers, analysts, or on-call staff

## What you will get

- A “pick your path” map: **if you are X, read Y**
- The shortest links to the tutorial, core explanations, and integration/operations references

## Pick your path (by goal)

- Prove the loop locally (copy/paste): [local end-to-end (service → logging → eval)](../tutorials/local-end-to-end.md)
- Do a production-like ship/rollback run: [production-like run (pipelines → object store → ship/rollback)](../tutorials/production-like-run.md)
- Integrate the API into your product: [How-to: integrate recsys-service into an application](../how-to/integrate-recsys-service.md)
- Troubleshoot integration issues: [Troubleshooting for integrators](../how-to/troubleshooting-integration.md)
- Deploy on Kubernetes (Helm): [Deploy with Helm (production-ish)](../how-to/deploy-helm.md)
- Operate pipelines day-to-day: [How-to: operate recsys-pipelines](../how-to/operate-pipelines.md)
- Evaluate and make ship/rollback decisions: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)

## Pick your path (by role)

### Developer / platform engineer

- Architecture + boundaries: [Suite architecture](../explanation/suite-architecture.md)
- Config + contracts:
  - [Configuration reference](../reference/config/index.md)
  - [Data contracts](../reference/data-contracts/index.md)
- API schema (Swagger UI): [API Reference](../reference/api/api-reference.md)

### Recommendation engineer / ML engineer

- [RecSys engineering hub](../recsys-engineering/index.md)
- Ranking core (ports/adapters, signals): [recsys-algo](../recsys-algo/index.md)
- Pipelines lifecycle + artifacts: [recsys-pipelines docs](../recsys-pipelines/docs/index.md)

### Product / business stakeholder

- What you get + what you need: [What the RecSys suite is (stakeholder overview)](what-is-recsys.md)
- Pilot plan and ownership: [Pilot plan (2–6 weeks)](pilot-plan.md) and [Responsibilities (RACI): who owns what](responsibilities.md)
- How success is measured: [Metrics: what we measure and why](../recsys-eval/docs/metrics.md)

### SRE / on-call

- Suite runbooks: [Runbook: Service not ready](../operations/runbooks/service-not-ready.md)
- Pipeline runbooks: [Runbook: Pipeline failed](../recsys-pipelines/docs/operations/runbooks/pipeline-failed.md)
- Capacity guidance: [Performance and capacity guide](../operations/performance-and-capacity.md)

## Read next

- Start here (overview): [Start here](index.md)
- Stakeholder overview: [What the RecSys suite is (stakeholder overview)](what-is-recsys.md)
