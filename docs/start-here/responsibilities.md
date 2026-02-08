---
diataxis: explanation
tags:
  - business
  - ops
---
# Responsibilities (RACI): who owns what
This page explains Responsibilities (RACI): who owns what and how it fits into the RecSys suite.


## Who this is for

Engineering leads, product owners, SRE/on-call, and security reviewers planning a pilot or production rollout of the
RecSys suite.

## What you will get

- A shared-responsibility model for adopting the RecSys suite
- A RACI matrix you can use during procurement and delivery planning
- Clear handoffs between integration, data, and operations

## Default responsibility model

By default, the **RecSys suite** is **self-hosted**: your organization runs the infrastructure and owns the data. The
maintainers provide the software, documentation, and support channels.

If you operate the suite as a managed service, the same responsibilities exist but the owner per row may change.

## Roles used in the RACI

- **PO**: Product owner (business outcomes, prioritization)
- **App**: Application team (frontend/backend integration)
- **Data**: Data engineering / analytics (events, pipelines, evaluation)
- **SRE**: Platform / SRE / on-call (infra, deploys, monitoring, incidents)
- **Sec**: Security / compliance (PII, access controls, risk review)
- **RecSys**: RecSys suite maintainers/support (product expertise, reviews, escalation)

Legend:

- **R** = Responsible (does the work)
- **A** = Accountable (owns the outcome)
- **C** = Consulted (gives input)
- **I** = Informed (kept in the loop)

## Pilot RACI (DB-only mode)

In DB-only mode you can validate the full “serve → log → eval” loop without object storage or pipelines.

| Activity | PO | App | Data | SRE | Sec | RecSys |
| --- | --- | --- | --- | --- | --- | --- |
| Define surfaces and success metrics | A/R | C | C | I | I | C |
| Provision dev/staging environments (DB, deploy) | I | I | C | A/R | C | C |
| Integrate `recsys-service` API into one surface | C | A/R | C | C | I | C |
| Propagate `request_id` and stable IDs | C | A/R | C | C | I | C |
| Emit exposure logs and outcome logs | C | R | A/R | C | C | C |
| Run offline evaluation (`recsys-eval`) and review results | C | C | A/R | I | I | C |
| Operate runbooks (empty recs / not ready) | I | C | C | A/R | I | C |
| Decide what to ship (gate/rollout plan) | A/R | C | C | C | C | C |

## Production RACI (artifact/manifest + pipelines)

Artifact/manifest mode adds a second “supply chain”: pipelines publish versioned signals and a manifest; the service
reads the manifest pointer.

| Activity | PO | App | Data | SRE | Sec | RecSys |
| --- | --- | --- | --- | --- | --- | --- |
| Deploy `recsys-service` to production (auth, tenancy, limits) | I | C | C | A/R | C | C |
| Run pipelines on schedule and publish manifests | I | I | A/R | C | I | C |
| Monitor freshness, latency, and empty-recs rate | I | I | C | A/R | I | C |
| Roll back config/rules and/or manifest on incident | I | C | C | A/R | I | C |
| Security review (PII, retention, access, audit) | I | C | C | C | A/R | C |
| Post-incident review and follow-ups | A/R | C | C | C | C | I |

## Read next

- Pilot plan: [Pilot plan (2–6 weeks)](pilot-plan.md)
- Security and privacy overview: [Security, privacy, and compliance (overview)](security-privacy-compliance.md)
- Local end-to-end tutorial: [local end-to-end (service → logging → eval)](../tutorials/local-end-to-end.md)
- Production readiness checklist: [Production readiness checklist (RecSys suite)](../operations/production-readiness-checklist.md)
