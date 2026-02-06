---
tags:
  - reference
  - integration
  - developer
  - evaluation
  - recsys-service
---

# Integration spec (one surface)

## Who this is for

- Application and platform engineers integrating `recsys-service`
- Data/analytics owners validating that attribution will work (join-rate is meaningful)

## What you will get

- The minimal request/response contract for `POST /v1/recommend`
- The exact tenant scoping and `request_id` propagation rules you must satisfy
- The logging invariants required for evaluation and troubleshooting

## Serving contract

### Endpoint

- `POST /v1/recommend`

### Headers (tenant scope + auth)

Tenant scope can come from **one** trusted source (claim/API key) or from a header.
If you send multiple sources, they must match.

Reference: [`reference/auth-and-tenancy.md`](../reference/auth-and-tenancy.md)

Local dev (dev headers enabled):

- Required:
  - `X-Dev-User-Id: <id>`
  - `X-Dev-Org-Id: <tenant>` (dev auth tenant context)
  - `X-Org-Id: <tenant>` (tenant scope)
- Optional:
  - `X-Request-Id: <request_id>` (recommended)

Production (JWT / API key):

- Required:
  - `Authorization: Bearer <token>` (JWT), or `X-API-Key: <key>` (API key mode)
- Recommended:
  - Omit `X-Org-Id` if tenant scope comes from a claim/key (prevents mismatch bugs).
  - If you do send `X-Org-Id`, it must equal the tenant from the claim/key.
- Optional:
  - `X-Request-Id: <request_id>` (recommended)

### Body (minimum)

Required fields:

- `surface` (stable name, for example `home_feed`)
- `k` (count requested; subject to server limits)
- `user.user_id` and/or `user.session_id` (stable, pseudonymous)

Example minimal body:

```json
{
  "surface": "home",
  "k": 10,
  "user": { "user_id": "u_123", "session_id": "s_456" }
}
```

### Response (minimum)

Your integration must handle:

- `items[]` (may be empty)
- `meta.request_id` (must be captured and propagated to logs)
- `warnings[]` (treat as actionable: missing signals, filtered candidates, etc.)

## Attribution contract (`request_id`)

This is the most important contract in the suite.

Rules:

1. Generate **one `request_id` per rendered list**.
2. Use the same `request_id` across:
   - recommend request (`X-Request-Id`), or capture `meta.request_id` from the response
   - exposure logs (what you showed)
   - outcome logs (what the user did)
   - assignment logs (if you bucket experiments)

Reference:

- Join logic (exposures ↔ outcomes): [`reference/data-contracts/join-logic.md`](../reference/data-contracts/join-logic.md)

## Logging contract (minimum)

You must produce exposures and outcomes that can be joined by `request_id`.

Canonical schemas and examples:

- Data contracts hub: [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
- Eval events schemas: [`reference/data-contracts/eval-events.md`](../reference/data-contracts/eval-events.md)
- Minimum instrumentation spec: [`reference/minimum-instrumentation.md`](../reference/minimum-instrumentation.md)

Security baseline:

- Do not log raw PII. Use stable pseudonymous IDs.

## Verification

### Validate request shape during development

Use the validation endpoint to catch missing fields and common warnings early:

```bash
curl -fsS http://localhost:8000/v1/recommend/validate \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: integ-spec-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1"}}'
```

### Validate logs against schemas

```bash
(cd recsys-eval && ./bin/recsys-eval validate --schema exposure.v1 --input /tmp/exposures.jsonl)
(cd recsys-eval && ./bin/recsys-eval validate --schema outcome.v1 --input /tmp/outcomes.jsonl)
```

## Read next

- Integration checklist: [`how-to/integration-checklist.md`](../how-to/integration-checklist.md)
- Minimal integration walkthrough: [`how-to/integrate-recsys-service.md`](../how-to/integrate-recsys-service.md)
- Exposure logging & attribution: [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)
- Runbooks (empty recs): [`operations/runbooks/empty-recs.md`](../operations/runbooks/empty-recs.md)
