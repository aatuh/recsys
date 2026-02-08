---
diataxis: how-to
tags:
  - logging
  - evaluation
  - instrumentation
---
# How-to: validate logging and joinability

Use this guide to confirm your logs are sufficient for attribution and evaluation.

## What you are validating

You want to be able to answer, reliably:

- What items were shown to a user (exposures)
- What the user did (outcomes)
- Whether each outcome can be attributed to a specific exposure (`request_id`)

## Step 1 — Confirm you log the minimal fields

Follow the canonical spec:

- Minimum instrumentation spec: [Minimum instrumentation spec](../reference/minimum-instrumentation.md)

At minimum, verify:

- **Exposure events** include `request_id`, `user_id` (pseudonymous), `ts`, and the ranked `items` list with rank
- **Outcome events** include `request_id`, `user_id` (pseudonymous), `ts`, and the outcome event type (click/conversion/etc.)

## Step 2 — Validate schema versions and contracts

- Data contracts (schemas): [Data contracts](../reference/data-contracts/index.md)
- Exposure/outcome schema details: [Exposure/outcome/assignment schemas](../reference/data-contracts/exposure-outcome-assignment.md)

## Step 3 — Run a joinability check

You need a quick "does this join" sanity check in your storage.

Checklist:

- [ ] For a random sample of outcomes, an exposure exists with the same `request_id`
- [ ] `request_id` is stable across services in the request path (no regeneration)
- [ ] One `request_id` maps to one recommendation response (no reuse across unrelated requests)

If joinability fails, stop and fix instrumentation before interpreting metric deltas.

Start here:

- Tutorial: [Verify joinability](../tutorials/verify-joinability.md)
- Join logic reference: [Join logic](../reference/data-contracts/join-logic.md)

## Step 4 — Confirm attribution logic (what counts as “influenced”)

Attribution is a product decision. Decide and document:

- which outcome types count (click, add_to_cart, purchase, etc.)
- the attribution window (time between exposure and outcome)
- multi-touch rules (first-touch, last-touch, linear, etc.)

Reading:

- Explanation: [Exposure logging and attribution](../explanation/exposure-logging-and-attribution.md)

## Step 5 — Produce one evidence kit bundle

Produce one minimal bundle you can share internally:

- sample recommendation response with `request_id`
- one exposure sample and schema version
- one outcome sample and schema version
- a short join-rate sanity summary

Template:

- Evidence kit template: [Evidence](../for-businesses/evidence.md)

## Read next

- Minimum instrumentation spec: [Minimum instrumentation spec](../reference/minimum-instrumentation.md)
- Interpreting metrics and reports: [Interpreting metrics and reports](../explanation/metric-interpretation.md)
- Run eval and make ship decisions: [Run eval and ship](run-eval-and-ship.md)
- Data contracts: [Data contracts](../reference/data-contracts/index.md)
