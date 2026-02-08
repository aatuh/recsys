---
diataxis: how-to
tags:
  - quickstart
  - checklist
  - business
  - ops
---
# Customer onboarding checklist
This guide shows how to customer onboarding checklist in a reliable, repeatable way.


## Who this is for

- Customer developer / platform engineer
- Security/compliance reviewer
- Data engineer / analytics owner
- SRE / on-call owner
- Product owner / stakeholder

## What you will get

- A practical checklist you can use to run a pilot and move to production safely
- Links to the exact docs pages that answer “how do we do that?”

## Checklist

### 1) Scope and ownership

- [ ] Agree on the first surface(s) and success metrics (CTR, conversion, revenue, retention).
- [ ] Confirm roles and responsibilities (RACI):
  - [Responsibilities (RACI): who owns what](responsibilities.md)
- [ ] Pick a pilot plan and timeline:
  - [Pilot plan (2–6 weeks)](pilot-plan.md)

### 2) Security and privacy

- [ ] Confirm PII handling and retention expectations.
- [ ] Decide where exposure/outcome logs live and who can access them.
- [ ] Review the suite security overview:
  - [Security, privacy, and compliance (overview)](security-privacy-compliance.md)

### 3) Integration (serving API)

- [ ] Run the local end-to-end tutorial once (proves the loop):
  - [local end-to-end (service → logging → eval)](../tutorials/local-end-to-end.md)
- [ ] Implement `request_id` propagation and exposure/outcome logging:
  - [Exposure logging and attribution](../explanation/exposure-logging-and-attribution.md)
- [ ] Integrate `/v1/recommend` into your product:
  - [How-to: integrate recsys-service into an application](../how-to/integrate-recsys-service.md)

### 4) Data and pipelines

- [ ] Decide DB-only vs artifact/manifest mode:
  - [Data modes: DB-only vs artifact/manifest](../explanation/data-modes.md)
- [ ] If using pipelines, run the pipelines quickstart and confirm artifacts/manifest publishing:
  - [Run locally (filesystem mode)](../recsys-pipelines/docs/tutorials/local-quickstart.md)

### 5) Operations

- [ ] Deploy strategy agreed (Helm, manifests, secrets):
  - [Deploy with Helm (production-ish)](../how-to/deploy-helm.md)
- [ ] Run the production readiness checklist:
  - [Production readiness checklist (RecSys suite)](../operations/production-readiness-checklist.md)
- [ ] Ensure on-call has the runbooks bookmarked:
  - [Runbook: Service not ready](../operations/runbooks/service-not-ready.md)
  - [Runbook: Empty recs](../operations/runbooks/empty-recs.md)
  - [Runbook: Stale manifest (artifact mode)](../operations/runbooks/stale-manifest.md)

## Read next

- Minimum pilot setup (one surface): [How-to: minimum pilot setup (one surface)](../how-to/pilot-minimum-setup.md)
- Security, privacy, compliance: [Security, privacy, compliance](security-privacy-compliance.md)
- Start evaluation: [Start an evaluation](../evaluate/index.md)
