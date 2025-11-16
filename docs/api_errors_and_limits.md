# API Errors & Limits

This section centralizes how RecSys responds when things go wrong. Use it to implement robust retry/backoff logic and to understand the maximum payload sizes we enforce.

## Status codes & behavior

- **200 OK** — Successful write/read. Responses include `trace_id`. Action: nothing to fix.
- **400 Bad Request** — Missing `namespace`, malformed JSON, missing `X-Org-ID`. Action: validate headers and payload shape; do not retry until corrected.
- **401 / 403 Unauthorized** — API key missing/invalid when auth is enabled. Action: include the correct `X-API-Key` or `Authorization` header.
- **404 Not Found** — Namespace doesn’t exist, item/user not seeded, or wrong endpoint path. Action: confirm namespace spelling, ingestion, and URL.
- **409 Conflict** — Duplicate IDs (e.g., rule or override name). Action: fetch the existing resource, resolve the conflict, and retry once fixed.
- **422 Unprocessable Entity** — Invalid override values (e.g., `overrides.blend` sums to zero), unknown event types, invalid trait keys. Action: adjust the payload to match documented ranges/names; this is a permanent failure until corrected.
- **429 Too Many Requests** — Rate limit exceeded (`API_RATE_LIMIT_RPM`). Action: apply exponential backoff or request higher limits.
- **500 Internal Server Error** — Unexpected bug or downstream outage. Action: safe to retry with jitter; capture the `trace_id` and share with support if the issue persists.

All errors return JSON:

```json
{
  "code": "missing_org_id",
  "message": "X-Org-ID header missing",
  "details": {},
  "trace_id": "01HF3WED9Q1800KJ7Q4MJ4GB8E"
}
```

- `code` – machine-readable identifier (e.g., `missing_org_id`, `invalid_override`).
- `message` – human explanation.
- `details` – optional map with extra context.
- `trace_id` – include this when escalating issues; it links to audit logs.

## Limits & guidance

- **Batch sizes**

  - `POST /v1/items:upsert` – up to 50 items per request.
  - `POST /v1/users:upsert` – up to 100 users.
  - `POST /v1/events:batch` – up to 500 events.

- **Payload sizes**

  - Soft limit: 1 MB per request body. Larger payloads may be rejected or cause elevated latency; use batching.

- **Recommendation parameters**

  - `k` (list length): typical range 10–50 per surface. Values over ~200 increase latency substantially; default guardrails (automatic checks that block risky configurations; see `docs/concepts_and_metrics.md`) assume `k ≤ 100`.
  - `fanout`: control via env/overrides; ensure `fanout ≥ k + diversity headroom`.

- **Idempotency**

  - `items:upsert`, `users:upsert`, `events:batch` are idempotent when the same payload is sent again—duplicates overwrite by `item_id/user_id/event_id`.
  - Manual override/rule creation is **not** idempotent by name; avoid reusing IDs without checking for conflicts.

- **Retry strategy**

  - Retry 500/429 responses with exponential backoff (e.g., 500ms → 1s → 2s). Do **not** retry 4xx responses other than 429 unless you’ve corrected the payload.

- **Rate limits**

  - Default RPM (requests per minute) is configured via `API_RATE_LIMIT_RPM` and `API_RATE_LIMIT_BURST`. Consult your deployment owner for exact values. 429 responses include a `Retry-After` header when available.

## Where this applies

- Linked from [`docs/quickstart_http.md`](quickstart_http.md) and [`docs/api_reference.md`](api_reference.md) to keep the detailed quickstart and endpoint reference concise.
- Applies to both hosted and local deployments; local stacks may not enforce RPM limits but will still use the same status codes and payload size expectations.
- For symptom-based guidance (“I get empty lists”, “I see many 429s”), see [`docs/faq_and_troubleshooting.md`](faq_and_troubleshooting.md).
