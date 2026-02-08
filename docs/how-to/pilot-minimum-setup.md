---
diataxis: how-to
tags:
  - how-to
  - pilot
  - developer
  - business
---
# How-to: minimum pilot setup (one surface)

This guide is the shortest practical path to a **credible pilot**: one recommendation surface, measurable lift, and an auditable ship/rollback loop.

## When to use this

- You want a 2–6 week pilot with clear exit criteria.
- You have time for **one** surface and **one** KPI.
- You need an internal story that passes engineering + security review.

## Outcome

By the end, you will have:

- one surface served by RecSys (DB-only or artifact/manifest mode)
- exposure + outcome logs that join reliably (`request_id`)
- one offline report you can share
- a written ship/hold/rollback rule for changes

## Prerequisites

- A chosen surface (examples: home feed, PDP similar-items)
- A stable join key you can carry end-to-end: `request_id`
- A pseudonymous `user_id` (no raw PII)

If you haven’t picked a data mode yet, follow: [How-to: choose a data mode](../start-here/choose-data-mode.md)

## Steps

### 1) Choose the surface and success metrics

Define:

- **Primary KPI** (choose one): CTR, conversion rate, revenue per mille, add-to-cart rate
- **Guardrail** (choose at least one): latency p95, error rate, catalog coverage, content diversity

Use templates:

- [Success metrics](../for-businesses/success-metrics.md)
- [Value model (ROI)](../for-businesses/value-model.md)

### 2) Implement the integration contract

Minimum contract:

- call `POST /v1/recommend` with `tenant_id`, `surface`, `k`, and a stable user/session id
- keep `request_id` from the response and attach it to outcomes
- log exposures (what you showed) and outcomes (what the user did)

Start here:

- [Integration spec](../reference/integration-spec.md)
- [Reference: recommend request fields](../reference/api/recommend-request.md)

### 3) Turn on exposure logging

If you are integrating a product, you can start by letting the service emit `exposure.v1` compatible events:

- enable `EXPOSURE_LOG_ENABLED=true`
- use `EXPOSURE_LOG_FORMAT=eval_v1`

See:

- [Exposure logging and attribution](../explanation/exposure-logging-and-attribution.md)
- [Reference: exposure/outcome schemas and examples](../reference/data-contracts/exposure-outcome-assignment.md)

### 4) Generate outcome events

Emit at least one outcome type for the pilot:

- `click`
- `conversion` (purchase/signup)

Make sure your outcome events carry:

- the same `request_id` you got from the recommend call
- `item_id`
- `ts` (RFC3339)

### 5) Validate joins and produce the first report

Run schema validation and a basic join-rate check:

- validate JSONL schemas locally
- compute join rate in your warehouse (or a one-off script)

Then run the default evaluation pack and generate a report you can share:

- [How-to: run evaluation and make ship decisions](run-eval-and-ship.md)

### 6) Decide: ship/hold/rollback

Write the simplest policy that prevents accidental regressions:

- ship only if KPI improves and guardrails do not regress
- roll back immediately if join rate drops below a threshold
- keep a written trail (what changed + why)

Use:

- [Operational reliability and rollback](../start-here/operational-reliability-and-rollback.md)
- [Decision playbook (ship/hold/rollback)](../recsys-eval/docs/decision-playbook.md)

## Read next

- Plan the pilot end-to-end: [How-to: run a pilot](../start-here/pilot-plan.md)
- Integrate one surface: [How-to: minimal integration (one surface)](minimal-integration.md)
- Debug empty results: [Runbook: Empty recs](../operations/runbooks/empty-recs.md)
- Understand the architecture: [How it works](../explanation/how-it-works.md)
