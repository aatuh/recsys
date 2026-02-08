---
diataxis: explanation
tags:
  - limitations
  - risk
  - evaluation
  - security
---
# Limitations and risks

This page explains the **common failure modes** when evaluating and operating RecSys.

!!! info "Scope"
    For the blunt list of product boundaries (what is not implemented or intentionally out of scope), see
    [Known limitations and non-goals](../start-here/known-limitations.md).

## Product boundaries (what you may hit quickly)

Known current boundaries include:

- Tenant creation is DB-only today (no tenant-create admin endpoint yet)
- Pipelines manifest registry is filesystem-based by default
- Kafka ingestion is scaffolded but not implemented as a streaming consumer

Source of truth:

- [Known limitations and non-goals](../start-here/known-limitations.md)

## Evaluation risks (how pilots go wrong)

### You can serve results but cannot measure impact

Symptoms:

- No stable join key (`request_id`) across exposure and outcome logs
- Low join rate (outcomes cannot be attributed to what was shown)
- Reports exist but are not trusted by stakeholders

What to do:

- Follow the minimum instrumentation spec: [Minimum instrumentation spec](../reference/minimum-instrumentation.md)
- Validate joinability early: [Verify joinability](../tutorials/verify-joinability.md)
- Use evaluation reasoning guardrails: [Evaluation validity](eval-validity.md)

### Metrics are treated as truth instead of signals

Offline and online metrics have limits. A single metric rarely tells the whole story.

What to read:

- Evaluation reasoning and pitfalls: [Evaluation reasoning and pitfalls](evaluation-reasoning.md)
- Interpreting results: [Interpreting results](../recsys-eval/docs/interpreting_results.md)

## Data risks (privacy, retention, and accidental leakage)

Typical risks:

- Logging raw identifiers where pseudonymous IDs would suffice
- Keeping event logs longer than necessary
- Mixing customer tenants in the same storage without clear boundaries

What to read:

- Security posture (overview): [Security, privacy, compliance](../start-here/security-privacy-compliance.md)
- Security artifacts (canonical pack): [Security pack](../security/security-pack.md)

## Operational risks (ship/rollback discipline)

Typical risks:

- No rollback drill (the lever is untested until it is urgent)
- Treating artifacts/manifests as mutable state (breaking reproducibility)
- Running pipelines without freshness and limit guardrails

What to read:

- Rollback model: [Operational reliability & rollback](../start-here/operational-reliability-and-rollback.md)
- Manifest lifecycle: [Artifacts and manifest lifecycle](artifacts-and-manifest-lifecycle.md)
- Pipelines invariants: [Pipelines operational invariants](pipelines-operational-invariants.md)

## Licensing and procurement risks

Typical risks:

- Starting a pilot without clarifying the license path (AGPL vs commercial)
- Procurement starts late and blocks shipping even after a successful pilot

What to do:

- Use the buyer flow: [Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)
- Confirm file-level rules: [Licensing](../licensing/index.md)

## Read next

- Known limitations (canonical): [Known limitations and non-goals](../start-here/known-limitations.md)
- Buyer flow: [Evaluation, pricing, and licensing](../pricing/evaluation-and-licensing.md)
- Production readiness checklist: [Production readiness checklist](../operations/production-readiness-checklist.md)
