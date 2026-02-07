---
tags:
  - how-to
  - integration
  - developer
  - evaluation
  - recsys-service
---

# How-to: first surface end-to-end

## Who this is for

- Lead developers / platform engineers integrating RecSys into an application
- Teams that want one recommendation surface to be **measurable** (logs join; a first report exists)

## Goal

Get one surface live end-to-end:

1. `POST /v1/recommend` is wired into your app for a stable `surface`.
2. Exposure and outcome logs share the same `request_id`.
3. You can produce a first `recsys-eval` report and make a ship/hold decision.

## Prereqs

- You picked **one** surface name (for example: `home_feed`) and can keep it stable.
- You can generate or propagate a stable `request_id` per rendered list.
- You can run RecSys locally (recommended) or in a dev/staging environment.

Recommended (fastest):

- Tutorial: [Quickstart (10 minutes)](../tutorials/quickstart.md)
- Decision guide: [Choose your data mode](../start-here/choose-data-mode.md)

## Steps

### 1) Choose your serving mode (DB-only vs artifact/manifest)

Use the decision guide:

- [Choose your data mode](../start-here/choose-data-mode.md)

If you just want the fastest first integration, start with **DB-only**.

### 2) Prove the API works before you touch your app

Before integrating, ensure you can get a non-empty response for your target tenant and surface.

Local runnable paths:

- DB-only: [Tutorial: Quickstart (10 minutes)](../tutorials/quickstart.md)
- Artifact/manifest: [Tutorial: production-like run](../tutorials/production-like-run.md)

### 3) Wire `POST /v1/recommend` into your app (contract first)

Use the canonical integration contract:

- [Integration spec (one surface)](../developers/integration-spec.md)

Minimal request example (local dev headers shown):

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: req-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home_feed","k":10,"user":{"user_id":"u_1","session_id":"s_1"}}'
```

Notes:

- In production, do not use dev headers. Use JWT/API keys and a trusted tenant scope.
- Capture and persist `meta.request_id` (or the `X-Request-Id` you supplied) so it can be joined to logs.

### 4) Instrument exposures + outcomes (minimum, joinable by `request_id`)

Use the canonical requirements:

- [Minimum instrumentation spec](../reference/minimum-instrumentation.md)

Definition of done for this step:

- Every rendered list emits **one** exposure record with the `request_id`.
- Every outcome event you care about (click/conversion) includes that same `request_id`.

### 5) Validate logs and generate your first report

Validate event logs against schemas:

```bash
recsys-eval validate --schema exposure.v1 --input exposures.jsonl
recsys-eval validate --schema outcome.v1 --input outcomes.jsonl
```

Then run the suite workflow:

- [Run eval & ship decisions](run-eval-and-ship.md)

### 6) Do one rollback drill (before you need it)

Practice one rollback path in a low-risk environment:

- [Operational reliability & rollback](../start-here/operational-reliability-and-rollback.md)

## Verify

- `POST /v1/recommend` returns a non-empty `items[]` list for your surface (or you have a defined fallback UX).
- Exposures and outcomes join by the same `request_id` (join-rate is not near-zero).
- You can produce and share a first evaluation report.

## Pitfalls

- Generating two request IDs (one for the API call, another for logging) → joins fail.
- Logging raw PII instead of stable pseudonymous identifiers.
- Letting `surface` names drift (you can’t compare like-for-like over time).
- Treating `warnings[]` as ignorable (they often explain empty recs or missing signals).

## Read next

- Production readiness checklist: [`operations/production-readiness-checklist.md`](../operations/production-readiness-checklist.md)
- Minimum instrumentation spec: [`reference/minimum-instrumentation.md`](../reference/minimum-instrumentation.md)
- How it works (architecture + data flow): [`explanation/how-it-works.md`](../explanation/how-it-works.md)
