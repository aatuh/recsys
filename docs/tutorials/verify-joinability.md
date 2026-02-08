---
diataxis: tutorial
tags:
  - tutorial
  - joinability
  - evaluation
  - data-contracts
---
# Tutorial: Verify joinability (request IDs → outcomes)
In this tutorial you will follow a guided walkthrough and verify a working result.


## Who this is for

- Developers and analysts who need evaluation-ready data.
- Teams validating that exposures and outcomes can be joined correctly.

## What you will get

- A concrete test to prove that your logs can be joined for evaluation
- A checklist for common join-breakers

## Prereqs

- You can run a small controlled test in a staging or local environment.
- You can access exposure logs and (optionally) outcome logs.

If you're starting from scratch, run: [Quickstart (10 minutes)](quickstart.md)

## What “joinable” means (one screen)

A recommendation request is joinable when:

- The recommendation response and the exposure log share the same `request_id`
- Outcome events can be linked back to the same `request_id` (directly or via stable user/session identity)
- The join fields are stable across platforms and time windows

Reference:

- Join logic: [Event join logic (exposures ↔ outcomes ↔ assignments)](../reference/data-contracts/join-logic.md)
- Minimum instrumentation: [Minimum instrumentation spec (for credible evaluation)](../reference/minimum-instrumentation.md)

## Step 1: Run a controlled recommendation

Pick one request ID and keep it consistent:

```bash
curl -fsS http://localhost:8000/v1/recommend   -H 'Content-Type: application/json'   -H 'X-Request-Id: join-1'   -H 'X-Dev-User-Id: dev-user-1'   -H 'X-Dev-Org-Id: demo'   -H 'X-Org-Id: demo'   -d '{"surface":"home","k":5,"user":{"user_id":"u_1","session_id":"s_1"}}'   > /tmp/recsys.join.response.json
```

## Step 2: Confirm the exposure log contains the same `request_id`

In DB-only quickstart mode, the tutorial writes a local JSONL file. Confirm:

```bash
grep -n '"request_id"' /tmp/exposures.eval.jsonl | head
grep -n '"request_id":"join-1"' /tmp/exposures.eval.jsonl | head
```

Expected:

- At least one exposure line contains `"request_id":"join-1"`.

## Step 3: Produce a synthetic outcome event (optional)

If you have an outcome event stream, emit one event that references `join-1` (or otherwise can be joined).

Then validate:

- The outcome stream contains the same `request_id`, or
- The outcome can be joined via stable user/session identity + time window

Reference:

- Eval event schema: [recsys-eval event schemas (v1)](../reference/data-contracts/eval-events.md)

## Step 4: Run a “mini join” sanity check

Goal:

- Each exposure row for `join-1` should join to 0+ outcomes.
- There should be no outcomes referencing unknown `request_id` values.

See join pitfalls:

- [Event join logic (exposures ↔ outcomes ↔ assignments)](../reference/data-contracts/join-logic.md)

## Verify (Definition of Done)

- [ ] `request_id` appears in both the response and exposure logs
- [ ] Outcomes can be joined to exposures for at least one controlled test request

## Read next

- Exposure logging & attribution: [Exposure logging and attribution](../explanation/exposure-logging-and-attribution.md)
- Run evaluation and decide ship/hold/rollback: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)
- Verify determinism: [Tutorial: verify determinism](verify-determinism.md)
