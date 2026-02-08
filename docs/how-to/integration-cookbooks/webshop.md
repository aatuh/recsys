---
diataxis: how-to
tags:
  - how-to
  - integration
  - developer
  - recsys-service
---
# Cookbook: integrate with a webshop
This guide shows how to cookbook: integrate with a webshop in a reliable, repeatable way.


## Who this is for

Product and platform engineers integrating recommendations into an ecommerce experience.

## What you will get

- A surface plan for common ecommerce placements (`home`, `pdp`, `cart`)
- A minimal request + logging pattern that keeps attribution correct
- Early verification checks that prevent “we can’t measure lift” failures later

## Goal

Serve recommendations on common ecommerce surfaces and log events so you can measure impact and roll back safely.

## Typical surfaces

- `home`: “popular now” / personalized modules
- `pdp`: “similar items”
- `cart`: cross-sell / “frequently bought together”

Treat each surface as a stable namespace. Changing surface names breaks rule targeting, signal routing, and evaluation
slicing.

## Minimal serving integration

1. Call `POST /v1/recommend` with a stable `request_id`.
2. Render the returned `items[]` in the UI.
3. Log an exposure event (what was shown, with ranks).

Example request:

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: shop-req-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"pdp","k":12,"user":{"user_id":"u_123","session_id":"s_456"}}'
```

## Minimal logging (what makes evaluation possible)

Log two streams:

- **Exposure**: ranked list served (join key: `request_id`)
- **Outcome**: what the user did (click and/or purchase)

Outcome schema in `recsys-eval` supports `click` and `conversion`. A common webshop mapping is:

- PDP click → `click`
- purchase → `conversion` (optionally include order value as `value`)

Example outcome JSONL (one object per line):

```json
{"request_id":"shop-req-1","user_id":"u_123","item_id":"sku_42","event_type":"click","ts":"2026-02-05T10:00:03Z"}
{"request_id":"shop-req-1","user_id":"u_123","item_id":"sku_42","event_type":"conversion","value":79.00,"ts":"2026-02-05T10:05:12Z"}
```

## Verify (early, before you scale)

- Ensure `request_id` is the same in:
  - the HTTP request (`X-Request-Id`)
  - the exposure log record
  - the outcome log record
- Validate schemas:
  - `recsys-eval validate --schema exposure.v1 --input exposures.jsonl`
  - `recsys-eval validate --schema outcome.v1 --input outcomes.jsonl`
- Compute a join rate in your warehouse:
  - % of exposures with ≥1 matching outcome by `request_id`

## Pitfalls (common failure modes)

- **CDN/UI retries generate new request IDs**
  - Symptom: outcomes don’t join, join rate collapses.
  - Fix: generate `request_id` once per rendered list and propagate it.
- **Pre-fetching without rendering**
  - Symptom: many exposures with no outcomes; misleading metrics.
  - Fix: only log exposures when you actually render to a user.
- **Logging raw PII**
  - Fix: use pseudonymous IDs; never log email/phone.

## Read next

- Exposure logging & attribution: [Exposure logging and attribution](../../explanation/exposure-logging-and-attribution.md)
- Validate determinism and joinability: [Tutorial: verify determinism](../../tutorials/verify-determinism.md)
- Run evaluation and decide ship/hold/rollback: [How-to: run evaluation and make ship decisions](../run-eval-and-ship.md)
