---
diataxis: how-to
tags:
  - how-to
  - integration
  - developer
  - reference
---
# How-to: Minimal integration (one surface)
This guide shows how to how-to: Minimal integration (one surface) in a reliable, repeatable way.


## Who this is for

- Developers who want the smallest “real integration” that still allows evaluation

## Goal

Integrate **one surface** end-to-end so that you can:

- call `POST /v1/recommend` in your app
- log exposures (what was shown) with a stable `request_id`
- log outcomes (what happened) joinable by the same `request_id`
- run at least one offline evaluation report

This is the minimum viable integration for a credible pilot.

## Prereqs

- You can call the service API from your app or gateway
- You understand tenancy/auth headers in your environment:
  [Auth and tenancy reference](../reference/auth-and-tenancy.md)

## Steps

### 1) Choose your surface id (namespace)

Pick a stable identifier for where recs are shown. Example: `home`.

Rule: keep the surface id stable over time; treat changes as a breaking analytics event.

See: [Surface namespaces](../explanation/surface-namespaces.md)

### 2) Generate a stable `request_id` per render

A `request_id` is the join key across:

- request → response → exposure log → outcome log → evaluation report

Rule of thumb:

- generate one `request_id` **per render** (per page view / widget render)
- reuse it for every outcome attributed to that render

Canonical spec: [Integration spec (one surface)](../reference/integration-spec.md)

### 3) Call `POST /v1/recommend`

Minimal request:

```bash
curl -fsS https://YOUR_RECSYS_HOST/v1/recommend   -H 'Content-Type: application/json'   -H 'X-Request-Id: req-123'   -H 'X-Org-Id: TENANT_EXTERNAL_ID'   -d '{"surface":"home","k":10,"user":{"user_id":"u_1","session_id":"s_1"}}'
```

Verify:

- response contains `meta.request_id` equal to the `X-Request-Id`
- you store the returned `items` (at least `item_id` + `rank`) for exposure logging

API reference: [API Reference](../reference/api/api-reference.md)

### 4) Emit an exposure event (what was shown)

On the client or server (wherever is easiest), emit an exposure event that includes:

- `request_id`
- `user_id` (pseudonymous)
- `surface`
- the ranked list (`item_id`, `rank`)
- timestamp

Schema and examples: [recsys-eval event schemas (v1)](../reference/data-contracts/eval-events.md)

Minimum instrumentation (canonical): [Minimum instrumentation spec (for credible evaluation)](../reference/minimum-instrumentation.md)

### 5) Emit outcome events (what happened)

Track at least one outcome type for the pilot (common: `click`, `add_to_cart`, `purchase`).

Rules:

- include the same `request_id`
- include `item_id`
- include timestamp
- avoid raw PII (use pseudonymous identifiers)

Canonical join logic: [Event join logic (exposures ↔ outcomes ↔ assignments)](../reference/data-contracts/join-logic.md)

## Verify

- [ ] For a sample user session, you can produce:
  - one recommend response
  - one exposure event with the same `request_id`
  - one or more outcome events joinable by `request_id`
- [ ] Join-rate (exposures ↔ outcomes by `request_id`) is acceptable for your surface

## Pitfalls

- Generating a new `request_id` for every click (breaks attribution)
- Missing `rank` in exposures (prevents position-bias corrections)
- Changing surface ids mid-pilot (invalidates comparisons)

## Read next

- Integration checklist: [How-to: Integration checklist (one surface)](integration-checklist.md)
- Troubleshooting integration: [Troubleshooting for integrators](troubleshooting-integration.md)
- Run eval and make ship decisions: [How-to: run evaluation and make ship decisions](run-eval-and-ship.md)
