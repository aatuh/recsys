---
tags:
  - how-to
  - integration
  - developer
  - evaluation
  - security
---

# How-to: Integration checklist (one surface)

## Who this is for

- Application and platform engineers integrating one recommendation surface end-to-end.

## Goal

Make one surface “evaluation-ready”: calls to `POST /v1/recommend` are attributed correctly, outcomes join, and you can
produce a first report.

## Prereqs

- A tenant exists and admin bootstrap is complete (config + rules).
- You can run the local stack (use [`Tutorial: Quickstart`](../tutorials/quickstart.md) if you want the fastest setup).

!!! info "Canonical spec"
    For the request/response + logging contract (headers, tenant scope, `request_id` rules), see:
    [`developers/integration-spec.md`](../developers/integration-spec.md).

## Checklist

### 1) Tenant and surface

- [ ] Choose a stable `surface` name (for example: `home_feed`, `pdp_similar`).
- [ ] Ensure the tenant exists (and you know its external id).

### 2) Auth and tenancy

- [ ] Local dev: if `DEV_AUTH_ENABLED=true`, send the dev headers + `X-Org-Id` tenant scope.
- [ ] Production: disable dev headers (`DEV_AUTH_ENABLED=false`) and use JWT and/or API keys.
- [ ] Tenant scope comes from trusted auth claims (`AUTH_TENANT_CLAIMS`) or a trusted gateway (no “user-supplied” tenant).

Reference: [`reference/auth-and-tenancy.md`](../reference/auth-and-tenancy.md)

### 3) Request contract (recommend)

- [ ] Always send `surface` in the JSON body.
- [ ] Send pseudonymous `user_id` and/or `session_id` (no raw PII).
- [ ] Send a stable `X-Request-Id` (recommended) and treat it as the join key.
- [ ] During development, call `POST /v1/recommend/validate` to catch warnings early.

### 4) Response handling

- [ ] Treat `warnings[]` as actionable signals (missing signals, filtered candidates, etc.).
- [ ] Handle empty results safely (fallback UX or a baseline list).
- [ ] Capture `meta.request_id` (or the header value you supplied) and propagate it to downstream logs.

### 5) Logging and attribution (minimum)

- [ ] Emit exposure logs and outcome logs that share the same `request_id`.
- [ ] Ensure every record includes `tenant_id` and `surface` (in context fields or top-level fields, depending on schema).
- [ ] Validate logs against schemas before computing metrics.

Reference: [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)

### 6) Evaluation readiness

- [ ] Join-rate is not near-zero (broken joins invalidate metrics).
- [ ] You can produce at least one report that compares baseline vs candidate.

Start here: [`how-to/run-eval-and-ship.md`](run-eval-and-ship.md)

### 7) Operational behavior (minimum)

- [ ] Handle `429` throttling and respect `Retry-After`.
- [ ] Decide retry behavior to avoid double-counting downstream events.
- [ ] Practice one rollback drill (config/rules and/or manifest pointer).

Start here: [`start-here/operational-reliability-and-rollback.md`](../start-here/operational-reliability-and-rollback.md)

## Minimal integration example (copy/paste)

For a full walkthrough, see:
[Minimal integration example](integrate-recsys-service.md#minimal-integration-example-one-surface)

Serve one request:

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: req-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":10,"user":{"user_id":"u_1","session_id":"s_1"}}'
```

Then:

- Log one exposure record per rendered list (join key: `request_id`).
- Log outcome events with that same `request_id`.

## Verify

Validate a request:

```bash
curl -fsS http://localhost:8000/v1/recommend/validate \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: integ-check-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1"}}'
```

Then call recommend and confirm `meta.request_id` is present:

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: integ-check-2' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1"}}'
```

## Pitfalls

- Missing/unstable `request_id` (you can’t attribute outcomes).
- Missing tenant scope headers in dev mode (`X-Dev-Org-Id` vs `X-Org-Id` mismatch).
- Retrying without thinking about downstream counting (double outcomes).
- Logging raw PII instead of pseudonymous identifiers.

## Read next

- Run eval and ship decisions: [`how-to/run-eval-and-ship.md`](run-eval-and-ship.md)
- Data contracts (schemas + examples): [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
- Exposure logging & attribution: [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)
