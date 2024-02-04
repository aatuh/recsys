# Pilot plan (4–6 weeks)

## Who this is for

Product owners, engineering leads, and delivery teams planning a pilot of the RecSys suite.

## What you will get

- A realistic timeline for a first pilot (from “hello world” to production readiness)
- Clear deliverables and exit criteria per phase
- The minimum instrumentation needed to measure impact

## What “success” looks like

At the end of a pilot, you should be able to answer:

- Can we serve recommendations reliably for our key surfaces (latency, availability, no “empty recs”)?
- Can we explain and roll back changes (config/rules/artifacts) without drama?
- Can we measure quality and impact from real logs (offline + online)?

If you cannot measure impact, you are not “piloting a recommender” yet—you are only integrating an endpoint.

## Prerequisites (non-negotiable)

You need the ability to produce:

- exposure logs (what was served, with ranks)
- outcome logs (what the user did)
- stable IDs for joins (a `request_id`, plus a pseudonymous `user_id` or session identifier)

See: [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)

## Phase 1 (Week 1): baseline + instrumentation

Goal: ship a safe baseline and prove you can measure it.

Deliverables:

- One surface integrated end-to-end (client → `recsys-service` → response rendered)
- Exposure + outcome logging wired from production-like traffic (even if small)
- First `recsys-eval` report generated from real logs (joins validated)
- Runbooks exercised once: “service not ready”, “empty recs”

Recommended scope:

- Start in **DB-only mode** to minimize moving parts.
- Use a deterministic baseline algorithm (popularity is fine).

Exit criteria:

- `recsys-eval validate` succeeds for your logs
- P95 latency and error rate are within acceptable bounds for your product

## Phase 2 (Weeks 2–3): improve relevance safely

Goal: add one higher-signal candidate source and gain iteration speed.

Typical upgrades:

- similarity/co-visitation signals (often high ROI with modest complexity)
- basic business rules (pin/exclude, constraints) for control and trust
- segmentation (by surface, locale, tenant, or other stable context keys)

Deliverables:

- A second algorithm/config version evaluated against baseline (offline)
- A small “ship checklist” used before rollout (what changed, how to roll back)
- A rollback drill completed once (config/rules and/or artifact manifest)

Exit criteria:

- Offline evaluation shows a consistent improvement (or a clear tradeoff you accept)
- You can roll back within minutes with a known procedure

## Phase 3 (Weeks 4–6): production hardening + experimentation

Goal: move from “works” to “operationally safe”.

Typical additions:

- artifact/manifest mode (pipelines publish versioned artifacts; service reads a manifest pointer)
- A/B experiments for key surfaces with clear guardrails
- SLOs and alerting: latency, error rate, empty-recs rate, artifact freshness

Deliverables:

- A documented on-call playbook and escalation path
- Production readiness checklist completed
- One controlled experiment run (even if only to validate instrumentation)

Exit criteria:

- On-call can triage common failures from runbooks
- Changes are shipped with gates and are reversible

## Read next

- Stakeholder overview: [`start-here/what-is-recsys.md`](what-is-recsys.md)
- Responsibilities (RACI): [`start-here/responsibilities.md`](responsibilities.md)
- Security and privacy overview: [`start-here/security-privacy-compliance.md`](security-privacy-compliance.md)
- Local end-to-end tutorial: [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)
- Production readiness checklist: [`operations/production-readiness-checklist.md`](../operations/production-readiness-checklist.md)
