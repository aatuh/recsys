---
diataxis: how-to
tags:
  - how-to
  - integration
  - developer
  - recsys-service
---
# Cookbook: integrate with a content feed
This guide shows how to cookbook: integrate with a content feed in a reliable, repeatable way.


## Who this is for

Engineers integrating recommendations into a feed-like product (articles, videos, collections).

## What you will get

- A surface plan for feed placements (`home`, `related`, `next_up`)
- A minimal request + attribution pattern that supports offline eval and experiments
- Pitfalls to avoid for infinite scroll and delayed outcomes

## Goal

Integrate recommendations into a feed-like product (articles, videos, collections) with attribution that supports both
offline evaluation and online experiments.

## Typical surfaces

- `home`: personalized feed modules
- `related`: “related content” blocks on detail pages
- `next_up`: “watch next” / “read next”

Use surfaces to represent “where the list appears”, not “what the algorithm is”. You can change algorithms without
breaking instrumentation if surfaces stay stable.

## Minimal serving integration

1. Call `POST /v1/recommend` with a stable `request_id`.
2. Render results in the feed module.
3. Log exposure and outcomes with that same `request_id`.

Example request:

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: feed-req-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"related","k":8,"user":{"user_id":"u_9","session_id":"s_9"}}'
```

## Outcomes: decide what counts as “success”

Pick one primary outcome to start:

- click → `click`
- long read / watch completion → `conversion`

If you change the definition of “conversion”, treat it as an evaluation contract change and communicate it (otherwise
metric trends become incomparable).

## Verify

- Validate schemas with `recsys-eval validate`.
- Slice join rates by platform (web/app) and surface (`home`, `related`).
- Track “empty recs” rate per surface; it’s a fast signal of missing signals or bad rules.

## Pitfalls

- **Infinite scroll creates many recommendation calls**
  - Fix: log one exposure per rendered module instance; avoid reusing `request_id` across modules.
- **Outcome events missing `item_id`**
  - Symptom: conversions exist but can’t be attributed to ranked lists.
- **Delayed outcomes**
  - For long reads/watches, outcomes may arrive minutes later. Ensure your event pipeline preserves `request_id`.

## Read next

- Exposure logging & attribution: [Exposure logging and attribution](../../explanation/exposure-logging-and-attribution.md)
- Event join logic: [Event join logic (exposures ↔ outcomes ↔ assignments)](../../reference/data-contracts/join-logic.md)
- Run evaluation and decide ship/hold/rollback: [How-to: run evaluation and make ship decisions](../run-eval-and-ship.md)
