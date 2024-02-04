# Start here

This site documents the **RecSys suite**: a production-ready set of modules for building, shipping, and operating a
recommendation system.

## Who this is for

- New evaluators of the RecSys suite
- Engineers who want the shortest path to a runnable local setup
- Stakeholders who need the “what do we get, what do we need?” overview

## What you will get

- A recommended reading and execution path (tutorial → concepts → integration)
- Role-based entry points (lead dev, recsys engineer, stakeholder, SRE)
- Links to the canonical contracts, API, and operational runbooks

## Recommended path

1. **Run the suite locally end-to-end**
   - Tutorial: [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)
   - Optional: production-like artifacts + ship/rollback: [`tutorials/production-like-run.md`](../tutorials/production-like-run.md)

2. **Understand the architecture and data flow**
   - Diagram: [`start-here/diagrams/suite-context.md`](diagrams/suite-context.md)
   - Explanation: [`explanation/suite-architecture.md`](../explanation/suite-architecture.md)
   - Repo layout: [`start-here/repo-layout.md`](repo-layout.md)
   - Known limitations: [`start-here/known-limitations.md`](known-limitations.md)

3. **Integrate, operate, and validate**
   - Integrate the API: [`how-to/integrate-recsys-service.md`](../how-to/integrate-recsys-service.md)
   - Operate pipelines: [`how-to/operate-pipelines.md`](../how-to/operate-pipelines.md)
   - Run evaluation and ship decisions: [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)

## Role-based entry points

### Lead developer / platform engineer

- Suite architecture and contracts:
  - [`explanation/suite-architecture.md`](../explanation/suite-architecture.md)
  - [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
  - [`reference/config/index.md`](../reference/config/index.md)
  - [`reference/api/openapi.yaml`](../reference/api/openapi.yaml)
  - Security overview: [`start-here/security-privacy-compliance.md`](security-privacy-compliance.md)

### Recommendation engineer / ML engineer

- Ranking core:
  - [`recsys-algo/index.md`](../recsys-algo/index.md)
- Pipelines outputs:
  - [`recsys-pipelines/docs/reference/output-layout.md`](../recsys-pipelines/docs/reference/output-layout.md)

### Product / business stakeholder

- What the suite is and what it takes to pilot:
  - [`start-here/what-is-recsys.md`](what-is-recsys.md)
  - [`start-here/pilot-plan.md`](pilot-plan.md)
  - [`start-here/responsibilities.md`](responsibilities.md)
  - Security and privacy overview: [`start-here/security-privacy-compliance.md`](security-privacy-compliance.md)

- What to expect from evaluation and decisions:
  - [`recsys-eval/docs/interpreting_results.md`](../recsys-eval/docs/interpreting_results.md)
  - [`recsys-eval/docs/metrics.md`](../recsys-eval/docs/metrics.md)

### SRE / on-call

- Operational runbooks:
  - [`operations/runbooks/service-not-ready.md`](../operations/runbooks/service-not-ready.md)
  - [`operations/runbooks/empty-recs.md`](../operations/runbooks/empty-recs.md)
  - [`operations/runbooks/rollback-config-rules.md`](../operations/runbooks/rollback-config-rules.md)
  - Security overview: [`start-here/security-privacy-compliance.md`](security-privacy-compliance.md)
  - [`recsys-pipelines/docs/operations/runbooks/pipeline-failed.md`](../recsys-pipelines/docs/operations/runbooks/pipeline-failed.md)
  - [`recsys-pipelines/docs/operations/runbooks/validation-failed.md`](../recsys-pipelines/docs/operations/runbooks/validation-failed.md)
  - [`recsys-pipelines/docs/operations/runbooks/limit-exceeded.md`](../recsys-pipelines/docs/operations/runbooks/limit-exceeded.md)
  - [`recsys-pipelines/docs/operations/runbooks/stale-artifacts.md`](../recsys-pipelines/docs/operations/runbooks/stale-artifacts.md)

## Read next

- Local end-to-end tutorial: [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)
- Stakeholder overview: [`start-here/what-is-recsys.md`](what-is-recsys.md)
- Suite architecture: [`explanation/suite-architecture.md`](../explanation/suite-architecture.md)
