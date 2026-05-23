---
title: "Recommendation API observability: what to log before incidents"
description: "Recommendation API observability guidance for request IDs, tenant and surface labels, empty recommendation rates, latency, warnings, and rollback debugging."
language: "en"
pubDate: "2026-04-25"
translationKey: "api-observability"
tags: ["recommendation API", "observability", "operations"]
---

Recommendation API observability is easiest to add before the first production incident. Once a team is already
debugging empty responses, stale artifacts, or unexplained ranking changes, missing request context becomes the problem
inside the problem.

The goal is not noisy logging. A recommendation API should emit enough structured evidence to answer three questions:
what did the service receive, what decision path did it use, and which artifact or rule version produced the response?

## Recommendation API observability starts with request context

Every served recommendation response should be traceable through a stable request ID. That request ID needs to appear in
the API response metadata, exposure event, evaluation input, and operational logs.

Useful request context usually includes:

- request ID
- tenant, surface, and environment
- pseudonymous user or session identifier when available
- candidate or item count after eligibility filters
- active artifact manifest version
- ranking path, rule path, or fallback path
- response size and empty-response reason
- latency bucket and timeout status

Avoid raw personal data in operational logs. The [security overview](/security/) describes the current self-hosted and
pseudonymous identifier posture at a product level.

## Metrics that catch recommendation failures

Recommendation systems fail in ways that ordinary API uptime can miss. A service can return HTTP 200 while serving only
fallbacks, stale artifacts, or empty lists.

Track metrics that map to user-visible quality and operator action:

- empty recommendation rate by tenant and surface
- fallback usage rate
- artifact freshness age
- p95 and p99 latency by endpoint
- timeout and degraded-mode rate
- exposure event write failures
- outcome join coverage for evaluated traffic

These metrics should be visible before a launch and watched during every ranking or pipeline rollout.

## Debug empty recommendations without guessing

Empty recommendations are one of the fastest ways to lose trust in a recommender. The debug path should be mechanical:
find the request ID, inspect the decision path, confirm the active artifact, check eligibility filters, and compare the
result to the expected fallback behavior.

If the service can only say "no recommendations found," the operator has to reconstruct the incident from scattered
pipeline and API state. A better empty-response record says whether the problem came from missing candidates, stale
features, tenant scoping, rule filters, artifact load failure, or an intentional fallback.

The technical operations docs include a dedicated
[empty recommendations runbook](/documentation/technical/operations/runbooks/empty-recommendations/) and
[service readiness runbook](/documentation/technical/operations/runbooks/service-not-ready/).

## Preserve observability across rollback

Rollback should not erase the evidence needed to explain the failure. When a manifest pointer, config, or ranking rule is
rolled back, keep the previous and restored versions in the decision record.

That is where recommendation API observability connects to evaluation. A rollback is more credible when the team can
show which traffic saw the changed path, which guardrail failed, and which lever restored the previous behavior.

For implementation detail, start with the
[operations documentation](/documentation/technical/operations/) and the
[artifacts and pipelines guide](/documentation/technical/artifacts-and-pipelines/). For a commercial or pilot discussion,
use the [contact page](/contact/).
