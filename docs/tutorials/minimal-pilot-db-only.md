---
diataxis: tutorial
tags:
  - tutorial
  - quickstart
  - developer
  - recsys-service
---
# Tutorial: minimal pilot mode (DB-only, popularity baseline)
In this tutorial you will follow a guided walkthrough and verify a working result.


## Who this is for

Product owners and engineering leads who want a low-friction pilot of the **RecSys suite** without object storage or
offline pipelines.

## What you will get

- A popularity-only baseline you can ship safely
- The minimum instrumentation needed to measure impact
- A clear list of what this mode proves (and what it does not)

## What this pilot answers (business + engineering)

With DB-only + popularity, you can answer:

- Can we integrate a recommendation placement end-to-end (API → UI render)?
- Can we attribute outcomes to exposures (do joins work)?
- Are latency, error rate, and “empty recs” within acceptable bounds?
- Do we have the data and IDs needed to evaluate and iterate?

This pilot does not answer “how good is personalization”. It answers “is the loop operationally real”.

## Prereqs

- One surface to start (for example: `home`)
- A way to generate or propagate a stable `request_id` per rendered list
- A pseudonymous `user_id` or `session_id` you can use consistently

## Minimal data you need (DB-only signals)

Populate these Postgres tables for your tenant:

- `item_tags` (catalog + tags used for constraints and filters)
- `item_popularity_daily` (daily popularity score per item and surface/namespace)

The fastest approach is to backfill a small set of top items (for one surface) and refresh daily.

## Steps (recommended)

1. Run the runnable local tutorial once (proves the loop end-to-end):
   - [local end-to-end (service → logging → eval)](local-end-to-end.md)

2. Replace the seed SQL with your own data loading:
   - Load `item_tags` from your catalog.
   - Load `item_popularity_daily` from page views / purchases / clicks (your choice).

3. Integrate one placement in your product:
   - Use `POST /v1/recommend` with a stable `request_id`.
   - Log exposures and outcomes with that same `request_id`.

4. Run evaluation on real logs early:
   - Validate schemas and compute join rates.

## Verify

- `POST /v1/recommend` returns a non-empty `items[]` list for your surface.
- Exposures and outcomes share the same stable `request_id`.
- Join-rate is not near-zero (otherwise KPIs are not trustworthy).

See:

- Integration checklist: [How-to: Integration checklist (one surface)](../how-to/integration-checklist.md)
- Minimum instrumentation spec: [Minimum instrumentation spec (for credible evaluation)](../reference/minimum-instrumentation.md)

## Troubleshooting (runbooks)

- Service not ready: [Runbook: Service not ready](../operations/runbooks/service-not-ready.md)
- Empty recs: [Runbook: Empty recs](../operations/runbooks/empty-recs.md)
- Database migration issues: [Runbook: Database migration issues](../operations/runbooks/db-migration-issues.md)

## Read next

- Integrate service: [How-to: integrate recsys-service into an application](../how-to/integrate-recsys-service.md)
- Define success metrics and exit criteria: [Success metrics (KPIs, guardrails, and exit criteria)](../for-businesses/success-metrics.md)
- Evaluation, pricing, and licensing (buyer guide): [Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)
