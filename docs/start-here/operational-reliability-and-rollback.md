---
diataxis: explanation
tags:
  - overview
  - business
  - ops
  - rollback
---
# Operational reliability and rollback
This page explains Operational reliability and rollback and how it fits into the RecSys suite.


## Who this is for

- Product owners and stakeholders who need confidence that “we can ship this safely”
- Engineering leads who want a shared mental model of what can go wrong and how we recover
- On-call/SRE who need the shortest path to the right runbook

## What you will get

- A clear model for what “healthy” means across serving, pipelines, and evaluation
- The rollback levers that exist in each layer (and when to use which)
- A first-incident checklist with links to the right runbooks

## Reliability model (plain language)

RecSys is designed so that:

- **Serving stays available** even when offline pipelines fail.
- **Changes are reversible**: you can roll back config/rules or artifact versions without redeploying everything.
- **Decisions are auditable**: logs and reports can explain what was shipped and why.

The most important invariants are:

- Pipelines publish artifacts first and update the “current” manifest pointer last.
- The service reads “current” and can fall back safely when data is missing.

## Rollback levers (what can we reverse?)

Use the smallest lever that fixes the user-facing issue.

### 1) Config/rules rollback (fast, common)

When to use:

- A bad rule or constraint caused empty or surprising recommendations.
- You need to revert request defaults or weights.

How:

- Use the config/rules rollback runbook: [Runbook: Roll back config/rules](../operations/runbooks/rollback-config-rules.md)

### 2) Artifact/manifest rollback (pipelines layer)

When to use:

- A published artifact version is wrong (bad data, wrong window, bad computation).
- Freshness is OK, but relevance regressed immediately after a publish.

How:

- Roll back artifacts safely: [How-to: Roll back artifacts safely](../recsys-pipelines/docs/how-to/rollback-safely.md)
- Roll back the manifest pointer: [How-to: Roll back to a previous artifact version](../recsys-pipelines/docs/how-to/rollback-manifest.md)

### 3) “Stop shipping” (hold changes)

When to use:

- Data quality is unreliable (join-rate is bad, validation is failing, SRM indicates instrumentation issues).
- You need to stabilize observability before trying new algorithms.

How:

- Follow the evaluation workflow: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)
- Use `recsys-eval` troubleshooting when metrics don’t make sense:
  [Troubleshooting: symptom -> cause -> fix](../recsys-eval/docs/troubleshooting.md)

## First incident checklist (start here under pressure)

Pick the symptom that best matches what you see:

- **Service is up, but recommendations are empty**
  - Runbook: [Runbook: Empty recs](../operations/runbooks/empty-recs.md)
- **Service is up, but data looks stale**
  - Runbook: [Runbook: Stale manifest (artifact mode)](../operations/runbooks/stale-manifest.md)
  - Pipelines runbook: [Runbook: Stale artifacts](../recsys-pipelines/docs/operations/runbooks/stale-artifacts.md)
- **Pipelines run failed**
  - Runbook: [Runbook: Pipeline failed](../recsys-pipelines/docs/operations/runbooks/pipeline-failed.md)
- **Evaluation report looks wrong (joins low, SRM warning, impossible lift)**
  - Start with: [Interpreting results: how to go from report to decision](../recsys-eval/docs/interpreting_results.md)
  - Then: [Troubleshooting: symptom -> cause -> fix](../recsys-eval/docs/troubleshooting.md)

## Read next

- Production readiness checklist: [Production readiness checklist (RecSys suite)](../operations/production-readiness-checklist.md)
- Rollback config/rules runbook: [Runbook: Roll back config/rules](../operations/runbooks/rollback-config-rules.md)
- Pipelines rollback: [How-to: Roll back artifacts safely](../recsys-pipelines/docs/how-to/rollback-safely.md)
- Interpreting eval results: [Interpreting results: how to go from report to decision](../recsys-eval/docs/interpreting_results.md)
