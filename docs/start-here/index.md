---
diataxis: explanation
tags:
  - quickstart
  - business
  - developer
---
# Start here

This site documents the **RecSys suite**: a production-ready set of modules for building, shipping, and operating a
recommendation system.

## Quick paths

<div class="grid cards" markdown>

- **[Quickstart (minimal)](../tutorials/quickstart-minimal.md)**  
  Fastest path to a non-empty `POST /v1/recommend` + an exposure log.
- **[Minimum components by goal](minimum-components-by-goal.md)**  
  Decide DB-only vs artifact/manifest mode and what you need to run.
- **[Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)**  
  Recommended evaluation path + procurement-ready links.

</div>

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
   - Tutorial: [local end-to-end (service → logging → eval)](../tutorials/local-end-to-end.md)
   - If you want the fastest possible first success: [Quickstart (minimal)](../tutorials/quickstart-minimal.md)
   - Optional: production-like artifacts + ship/rollback: [production-like run (pipelines → object store → ship/rollback)](../tutorials/production-like-run.md)

2. **Understand the architecture and data flow**
   - Diagram: [Suite Context](diagrams/suite-context.md)
   - One-page mental model: [How it works: architecture and data flow](../explanation/how-it-works.md)
   - Explanation: [Suite architecture](../explanation/suite-architecture.md)
   - Repo layout: [Repo layout and Go module paths](repo-layout.md)
   - Known limitations: [Known limitations and non-goals (current)](known-limitations.md)

3. **Integrate, operate, and validate**
   - Integrate the API: [How-to: integrate recsys-service into an application](../how-to/integrate-recsys-service.md)
   - Operate pipelines: [How-to: operate recsys-pipelines](../how-to/operate-pipelines.md)
   - Run evaluation and ship decisions: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)

## Role-based entry points

### Lead developer / platform engineer

- Suite architecture and contracts:
  - [Suite architecture](../explanation/suite-architecture.md)
  - [Data contracts](../reference/data-contracts/index.md)
  - [Configuration reference](../reference/config/index.md)
  - [OpenAPI spec (YAML)](../reference/api/openapi.yaml)
  - Security overview: [Security, privacy, and compliance (overview)](security-privacy-compliance.md)

### Recommendation engineer / ML engineer

Start here: [RecSys engineering hub](../recsys-engineering/index.md)

- Ranking core:
  - [recsys-algo](../recsys-algo/index.md)
- Pipelines outputs:
  - [Output layout (local filesystem)](../recsys-pipelines/docs/reference/output-layout.md)

### Product / business stakeholder

- What the suite is and what it takes to pilot:
  - [What the RecSys suite is (stakeholder overview)](what-is-recsys.md)
  - [Pilot plan (2–6 weeks)](pilot-plan.md)
  - [ROI and risk model](roi-and-risk-model.md)
  - [Responsibilities (RACI): who owns what](responsibilities.md)
  - [Operational reliability and rollback](operational-reliability-and-rollback.md)
  - Security and privacy overview: [Security, privacy, and compliance (overview)](security-privacy-compliance.md)
  - Licensing and pricing: [Licensing](../licensing/index.md)
  - Support model: [Support](../project/support.md)

- What to expect from evaluation and decisions:
  - [Interpreting results: how to go from report to decision](../recsys-eval/docs/interpreting_results.md)
  - [Metrics: what we measure and why](../recsys-eval/docs/metrics.md)

### SRE / on-call

- Operational runbooks:
  - [Runbook: Service not ready](../operations/runbooks/service-not-ready.md)
  - [Runbook: Empty recs](../operations/runbooks/empty-recs.md)
  - [Runbook: Roll back config/rules](../operations/runbooks/rollback-config-rules.md)
  - Security overview: [Security, privacy, and compliance (overview)](security-privacy-compliance.md)
  - [Runbook: Pipeline failed](../recsys-pipelines/docs/operations/runbooks/pipeline-failed.md)
  - [Runbook: Validation failed](../recsys-pipelines/docs/operations/runbooks/validation-failed.md)
  - [Runbook: Limit exceeded](../recsys-pipelines/docs/operations/runbooks/limit-exceeded.md)
  - [Runbook: Stale artifacts](../recsys-pipelines/docs/operations/runbooks/stale-artifacts.md)

## Read next

- Local end-to-end tutorial: [local end-to-end (service → logging → eval)](../tutorials/local-end-to-end.md)
- Stakeholder overview: [What the RecSys suite is (stakeholder overview)](what-is-recsys.md)
- Suite architecture: [Suite architecture](../explanation/suite-architecture.md)
