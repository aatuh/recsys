---
tags:
  - quickstart
  - overview
  - developer
  - business
---

# Docs map

## Who this is for

- Anyone evaluating the RecSys suite for adoption
- Teams onboarding new engineers, analysts, or on-call staff

## What you will get

- A “pick your path” map: **if you are X, read Y**
- The shortest links to the tutorial, core explanations, and integration/operations references

## Pick your path (by goal)

- Prove the loop locally (copy/paste): [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)
- Do a production-like ship/rollback run: [`tutorials/production-like-run.md`](../tutorials/production-like-run.md)
- Integrate the API into your product: [`how-to/integrate-recsys-service.md`](../how-to/integrate-recsys-service.md)
- Deploy on Kubernetes (Helm): [`how-to/deploy-helm.md`](../how-to/deploy-helm.md)
- Operate pipelines day-to-day: [`how-to/operate-pipelines.md`](../how-to/operate-pipelines.md)
- Evaluate and make ship/rollback decisions: [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)

## Pick your path (by role)

### Lead developer / platform engineer

- Architecture + boundaries: [`explanation/suite-architecture.md`](../explanation/suite-architecture.md)
- Config + contracts:
  - [`reference/config/index.md`](../reference/config/index.md)
  - [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
- API schema (canonical): [`reference/api/openapi.yaml`](../reference/api/openapi.yaml)

### Recommendation engineer / ML engineer

- Ranking core (ports/adapters, signals): [`recsys-algo/index.md`](../recsys-algo/index.md)
- Pipelines lifecycle + artifacts: [`recsys-pipelines/docs/index.md`](../recsys-pipelines/docs/index.md)

### Product / business stakeholder

- What you get + what you need: [`start-here/what-is-recsys.md`](what-is-recsys.md)
- Pilot plan and ownership: [`start-here/pilot-plan.md`](pilot-plan.md) and [`start-here/responsibilities.md`](responsibilities.md)
- How success is measured: [`recsys-eval/docs/metrics.md`](../recsys-eval/docs/metrics.md)

### SRE / on-call

- Suite runbooks: [`operations/runbooks/service-not-ready.md`](../operations/runbooks/service-not-ready.md)
- Pipeline runbooks: [`recsys-pipelines/docs/operations/runbooks/pipeline-failed.md`](../recsys-pipelines/docs/operations/runbooks/pipeline-failed.md)
- Capacity guidance: [`operations/performance-and-capacity.md`](../operations/performance-and-capacity.md)

## Read next

- Start here (overview): [`start-here/index.md`](index.md)
- Stakeholder overview: [`start-here/what-is-recsys.md`](what-is-recsys.md)
