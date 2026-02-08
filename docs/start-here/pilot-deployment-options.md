---
diataxis: how-to
tags:
  - start-here
  - deployment
  - pilot
  - operations
---
# Pilot deployment options
This guide shows how to pilot deployment options in a reliable, repeatable way.


## Who this is for

- Developers and platform engineers planning a pilot environment.
- Buyers who want to understand what “running it” entails.

## What you will get

- A concrete set of pilot deployment shapes
- What each option proves (and what it doesn’t)
- A recommended progression from pilot to production

## Option A: Local / developer sandbox (fastest)

Use when you want to prove:

- API integration and request/response semantics
- Tenancy isolation and headers/claims handling
- Exposure log creation for later evaluation

Start with:

- [Quickstart (10 minutes)](../tutorials/quickstart.md)
- [local end-to-end (service → logging → eval)](../tutorials/local-end-to-end.md)

## Option B: Single-tenant pilot environment (staging)

Use when you want to prove:

- End-to-end integration from your app to RecSys in a networked environment
- Observability, basic SRE runbooks, and rollback procedures
- A real(istic) traffic shadow or small production slice

Recommended building blocks:

- Deploy serving: [Deploy with Helm (production-ish)](../how-to/deploy-helm.md)
- Ops checklist: [Production readiness checklist (RecSys suite)](../operations/production-readiness-checklist.md)

## Option C: Production-like (artifacts + manifest + scheduled pipelines)

Use when you want to prove:

- Daily refresh of signals with guardrails
- Atomic publish/rollback and freshness SLOs
- A repeatable evaluation workflow

Start with:

- Production-like tutorial: [production-like run (pipelines → object store → ship/rollback)](../tutorials/production-like-run.md)
- Pipelines daily operation: [How-to: operate recsys-pipelines](../how-to/operate-pipelines.md)

## Recommended progression

1. **Local sandbox** → proves integration semantics cheaply.
2. **Staging pilot** → proves ops readiness with real integration constraints.
3. **Production-like** → proves long-term operating model (pipelines, ship/rollback, evaluation).

## Read next

- Pilot plan: [Pilot plan (2–6 weeks)](pilot-plan.md)
- Data modes: [Data modes: DB-only vs artifact/manifest](../explanation/data-modes.md)
- Reliability & rollback: [Operational reliability and rollback](operational-reliability-and-rollback.md)
