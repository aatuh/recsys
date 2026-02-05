---
tags:
  - quickstart
  - checklist
  - business
  - ops
---

# Customer onboarding checklist

## Who this is for

- Customer lead developer / platform engineer
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
  - [`start-here/responsibilities.md`](responsibilities.md)
- [ ] Pick a pilot plan and timeline:
  - [`start-here/pilot-plan.md`](pilot-plan.md)

### 2) Security and privacy

- [ ] Confirm PII handling and retention expectations.
- [ ] Decide where exposure/outcome logs live and who can access them.
- [ ] Review the suite security overview:
  - [`start-here/security-privacy-compliance.md`](security-privacy-compliance.md)

### 3) Integration (serving API)

- [ ] Run the local end-to-end tutorial once (proves the loop):
  - [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)
- [ ] Implement `request_id` propagation and exposure/outcome logging:
  - [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)
- [ ] Integrate `/v1/recommend` into your product:
  - [`how-to/integrate-recsys-service.md`](../how-to/integrate-recsys-service.md)

### 4) Data and pipelines

- [ ] Decide DB-only vs artifact/manifest mode:
  - [`explanation/data-modes.md`](../explanation/data-modes.md)
- [ ] If using pipelines, run the pipelines quickstart and confirm artifacts/manifest publishing:
  - [`recsys-pipelines/docs/tutorials/local-quickstart.md`](../recsys-pipelines/docs/tutorials/local-quickstart.md)

### 5) Operations

- [ ] Deploy strategy agreed (Helm, manifests, secrets):
  - [`how-to/deploy-helm.md`](../how-to/deploy-helm.md)
- [ ] Run the production readiness checklist:
  - [`operations/production-readiness-checklist.md`](../operations/production-readiness-checklist.md)
- [ ] Ensure on-call has the runbooks bookmarked:
  - [`operations/runbooks/service-not-ready.md`](../operations/runbooks/service-not-ready.md)
  - [`operations/runbooks/empty-recs.md`](../operations/runbooks/empty-recs.md)
  - [`operations/runbooks/stale-manifest.md`](../operations/runbooks/stale-manifest.md)

## Read next

- Docs map: [`start-here/docs-map.md`](docs-map.md)
- Start here (overview): [`start-here/index.md`](index.md)
