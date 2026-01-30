Below is a complete, implementable REST API spec for **recsys-svc**
(`recsys-service`). It’s written so you can directly translate it into OpenAPI.

I’m using:

* **Problem Details** for errors (`application/problem+json`). ([RFC Editor][1])
* **W3C Trace Context** for distributed tracing headers (`traceparent`,
  `tracestate`). ([W3C][2])
* OWASP guidance that **tenant/object scoping must be enforced everywhere**
  (BOLA risk). ([OWASP Foundation][3])
* Modern **RateLimit** headers (draft; widely adopted patterns exist). ([IETF Datatracker][4])

---

## Global conventions (apply to all endpoints)

### Base URL and versioning

* Base: `https://<host>`
* All endpoints are under `/v1/...`

### Authentication and tenancy

**Required for all `/v1/*` endpoints except health/metrics/version:**

* `Authorization: Bearer <JWT>` (OIDC)
* Tenancy is derived from token claim (example: `tenant_id` / `org_id`).
* If client supplies `X-Org-Id`, it must match the token claim (otherwise 403).

This is specifically to prevent cross-tenant access (broken object level
authorization / BOLA). ([OWASP Foundation][3])

### Correlation and tracing headers

**Request (optional but recommended):**

* `X-Request-Id: <uuid>` (client-generated, echoed back)
* `traceparent: ...` and optional `tracestate: ...` per W3C. ([W3C][2])

**Response:**

* `X-Request-Id: <uuid>`
* `traceparent` may be echoed/returned (implementation choice)

### Content negotiation

* Requests with JSON body must set:

  * `Content-Type: application/json`
* Responses:

  * success: `Content-Type: application/json`
  * errors: `Content-Type: application/problem+json` ([RFC Editor][1])

### Rate limiting headers (non-business middleware)

If enabled, responses may include (draft standard):

* `RateLimit-Policy: ...`
* `RateLimit: ...` ([IETF Datatracker][4])
  and/or legacy style `RateLimit-Limit/Remaining/Reset`. ([IETF][5])

On throttling:

* `429 Too Many Requests`
* `Retry-After: <seconds>`

### Error format (Problem Details)

All non-2xx errors return:

```json
{
  "type": "https://errors.recsys.example/problem/invalid-request",
  "title": "Invalid request",
  "status": 400,
  "detail": "Field 'k' must be between 1 and 200.",
  "instance": "/v1/recommend",
  "code": "RECSYS_INVALID_REQUEST",
  "request_id": "2b1c0e6e-3f12-4b4f-9f5f-2c8c0f2d2f8e"
}
```

This follows RFC 9457 (extensions like `code` and `request_id` are allowed). ([RFC Editor][1])

---

# Public serving endpoints

## 1) Recommend

### HTTP method

`POST`

### URL

`/v1/recommend`

### URL parameters

None.

### Business logic goal

Return a ranked list of recommended item IDs for a given tenant, surface, and
user/session context, applying constraints, rules, and diversification.

### Request headers

* `Authorization: Bearer <JWT>` (required)
* `Content-Type: application/json` (required)
* `Accept: application/json` (recommended)
* `X-Request-Id: <uuid>` (optional)
* `traceparent: ...` (optional) ([W3C][2])
* `X-Org-Id: <string>` (optional; must match token claim)

### Request JSON

```json
{
  "surface": "home",
  "segment": "default",
  "k": 20,

  "user": {
    "user_id": "u_123",
    "anonymous_id": "a_456",
    "session_id": "s_789"
  },

  "context": {
    "locale": "fi-FI",
    "device": "web",
    "country": "FI",
    "now": "2026-01-29T10:00:00Z"
  },

  "anchors": {
    "item_ids": ["item_1", "item_2"],
    "max_anchors": 50
  },

  "candidates": {
    "include_ids": [],
    "exclude_ids": ["item_999"]
  },

  "constraints": {
    "required_tags": ["winter", "outdoor"],
    "forbidden_tags": ["adult"],
    "max_per_tag": {"brand:nike": 2}
  },

  "weights": {
    "pop": 0.40,
    "cooc": 0.20,
    "emb": 0.40
  },

  "options": {
    "include_reasons": true,
    "explain": "summary",
    "include_trace": false,
    "seed": 123456789
  },

  "experiment": {
    "id": "exp_home_2026_01",
    "variant": "B"
  }
}
```

Notes:

* `user_id` may be omitted for anonymous use; `anonymous_id` or `session_id`
  is then required.
* `options.seed` makes stochastic steps reproducible.
* `candidates.include_ids` is a hard allow-list: if provided, only those IDs
  can appear in the response (after ranking/constraints).
* `candidates.exclude_ids` always removes those items from the response.
* Pin rules can inject items even if they are not in the base candidate pool.

### Response headers

* `Content-Type: application/json`
* `X-Request-Id: <uuid>`
* Optional rate limit headers ([IETF Datatracker][4])
* Optional `Cache-Control` (typically `no-store` for personalized responses)

### Response JSON (200)

```json
{
  "items": [
    {
      "item_id": "item_42",
      "rank": 1,
      "score": 0.8123,
      "reasons": ["embedding_similarity", "recent_popularity"],
      "explain": {
        "signals": {
          "pop": 0.70,
          "cooc": 0.00,
          "emb": 0.93
        },
        "rules": ["boost:campaign_winter"]
      }
    }
  ],
  "meta": {
    "tenant_id": "tenant_abc",
    "surface": "home",
    "segment": "default",
    "algo_version": "recsys-algo@v0.9.0",
    "config_version": "cfg_2026-01-29_001",
    "rules_version": "rules_2026-01-29_003",
    "request_id": "2b1c0e6e-3f12-4b4f-9f5f-2c8c0f2d2f8e",
    "timings_ms": {
      "candidates": 12,
      "scoring": 4,
      "postrank": 3,
      "total": 25
    },
    "counts": {
      "candidates_in": 400,
      "filtered": 120,
      "returned": 20
    }
  },
  "warnings": [
    {
      "code": "SIGNAL_UNAVAILABLE",
      "detail": "cooc signal unavailable; scoring used pop+emb only"
    }
  ]
}
```

### Status codes (business + common failures)

**Success**

* `200 OK`: recommendations returned (possibly empty list).
* `200 OK` with `warnings[]`: partial degradation (recommended pattern).

**Client errors**

* `400 Bad Request`: schema/validation failed (Problem Details body). ([RFC Editor][1])
* `401 Unauthorized`: missing/invalid token.
* `403 Forbidden`: tenant mismatch / insufficient scopes (BOLA protection). ([OWASP Foundation][3])
* `413 Payload Too Large`: request exceeds limits.
* `415 Unsupported Media Type`: not `application/json`.
* `422 Unprocessable Entity`: semantically invalid request (example: unknown
  surface, impossible constraints).

**Throttling**

* `429 Too Many Requests`: rate limited (middleware); include `Retry-After`
  and RateLimit headers. ([IETF Datatracker][4])

**Server/dependency**

* `500 Internal Server Error`: unexpected.
* `503 Service Unavailable`: dependency unavailable (store/feature index).
* `504 Gateway Timeout`: exceeded deadline.

### Non-business logic goals (middleware)

* Authn/authz + tenant isolation (prevent BOLA). ([OWASP Foundation][3])
* Rate limiting / quotas. ([IETF Datatracker][4])
* Request size caps, max `k`, max anchors, max excludes.
* Timeouts, retries (careful: retries only for idempotent dependency calls).
* Observability: tracing (`traceparent`), correlation id (`X-Request-Id`). ([W3C][2])
* Response compression (`gzip`) when `Accept-Encoding: gzip`.

---

## 2) Similar items

### HTTP method

`POST`

### URL

`/v1/similar`

### URL parameters

None.

### Business logic goal

Given an `item_id`, return the top-K similar items using embedding similarity
and/or co-occurrence, with optional constraints.

### Request headers

Same as `/v1/recommend`.

### Request JSON

```json
{
  "surface": "pdp",
  "segment": "default",
  "item_id": "item_42",
  "k": 20,
  "constraints": {
    "required_tags": [],
    "forbidden_tags": []
  },
  "options": {
    "include_reasons": true,
    "explain": "summary",
    "seed": 42
  }
}
```

### Response JSON (200)

Same shape as recommend:

```json
{
  "items": [
    { "item_id": "item_77", "rank": 1, "score": 0.91, "reasons": ["embedding_similarity"] }
  ],
  "meta": { "request_id": "..." },
  "warnings": []
}
```

### Status codes

Same as `/v1/recommend`, plus:

* `404 Not Found`: `item_id` does not exist (only if you choose strict mode;
  alternatively return `200` with empty list).

### Non-business logic goals

Same as `/v1/recommend`.

---

## 3) Validate request (no recommendations)

### HTTP method

`POST`

### URL

`/v1/recommend/validate`

### Business logic goal

Validate and normalize a recommend request without executing store calls or
ranking. Useful for integration testing and client debugging.

### Request/response

* Same request JSON as `/v1/recommend`.

### Response JSON (200)

```json
{
  "normalized_request": {
    "surface": "home",
    "segment": "default",
    "k": 20,
    "options": { "include_reasons": false, "explain": "none", "seed": 0 }
  },
  "warnings": [
    { "code": "DEFAULT_APPLIED", "detail": "segment defaulted to 'default'" }
  ],
  "meta": { "request_id": "..." }
}
```

### Status codes

* `200 OK`: valid.
* `400/401/403/413/415/422`: same meaning as above.

### Non-business goals

Same as `/v1/recommend`, but typically still subject to rate limits.

---

# Health and ops endpoints

## 4) Liveness

* **Method:** `GET`
* **URL:** `/healthz`
* **Goal:** process is up.
* **Auth:** none.
* **Response:** `200 OK` plain text or JSON.

## 5) Readiness

* **Method:** `GET`
* **URL:** `/readyz`
* **Goal:** ready to serve traffic (deps reachable, warm caches ok).
* **Auth:** none (or restrict in some environments).
* **Response:** `200 OK` ready, `503` not ready.

## 6) Metrics (Prometheus)

* **Method:** `GET`
* **URL:** `/metrics`
* **Goal:** expose metrics scrape endpoint.
* **Auth:** typically restricted at network layer.
* **Response:** `200 OK` text format.

## 7) Version/build info

* **Method:** `GET`
* **URL:** `/version`
* **Goal:** return build metadata.
* **Auth:** optional.
* **Response (200):**

```json
{ "service": "recsys-svc", "version": "1.2.3", "git_sha": "abc123", "built_at": "..." }
```

---

# Optional admin/control-plane endpoints (recommended if you manage config/rules here)

These must be protected with strong authz (admin scope) and strict tenant
checks (BOLA/BFLA class risks). ([OWASP Foundation][3])

## 8) Get tenant config

* **Method:** `GET`
* **URL:** `/v1/admin/tenants/{tenant_id}/config`
* **URL params:** `tenant_id` (string)
* **Goal:** fetch current config (weights, limits, feature flags).
* **Headers:** `Authorization` required; admin scope required.
* **200 response:**

```json
{
  "tenant_id": "tenant_abc",
  "config_version": "cfg_2026-01-29_001",
  "defaults": { "weights": { "pop": 0.4, "cooc": 0.2, "emb": 0.4 } },
  "limits": { "max_k": 200, "max_exclude_ids": 5000 }
}
```

* **Status codes:** `200`, `401`, `403`, `404` (tenant not found)

## 9) Update tenant config (atomic, versioned)

* **Method:** `PUT`
* **URL:** `/v1/admin/tenants/{tenant_id}/config`
* **Headers:**

  * `If-Match: "<config_version>"` (optional but recommended for safe updates)
* **Goal:** update config with optimistic concurrency.
* **200/204:** updated; returns new `config_version`.
* **409 Conflict:** version mismatch.
* **400/422:** invalid config (negative weights, impossible limits).

## 10) Get rules

* **Method:** `GET`
* **URL:** `/v1/admin/tenants/{tenant_id}/rules`
* **Goal:** fetch current merchandising/ranking rules.
* **Status codes:** `200/401/403/404`

## 11) Update rules (versioned)

* **Method:** `PUT`
* **URL:** `/v1/admin/tenants/{tenant_id}/rules`
* **Headers:** `If-Match` recommended
* **Status codes:** `200/204`, `409`, `400/422`

## 12) Invalidate caches

* **Method:** `POST`
* **URL:** `/v1/admin/tenants/{tenant_id}/cache/invalidate`
* **Body:**

```json
{ "targets": ["rules", "config", "popularity"], "surface": "home" }
```

* **Goal:** force cache invalidation for tenant/surface.
* **Status codes:** `202 Accepted` (async invalidation) or `200 OK`.

---

If you want, I can also provide:

* an OpenAPI 3.1 skeleton that matches the above,
* a minimal “public-only” subset (recommend + similar + health) if you plan to
  manage config/rules elsewhere.

[1]: https://www.rfc-editor.org/rfc/rfc9457.html?utm_source=chatgpt.com "RFC 9457: Problem Details for HTTP APIs"
[2]: https://www.w3.org/TR/trace-context/?utm_source=chatgpt.com "Trace Context"
[3]: https://owasp.org/API-Security/editions/2023/en/0xa1-broken-object-level-authorization/?utm_source=chatgpt.com "API1:2023 Broken Object Level Authorization"
[4]: https://datatracker.ietf.org/doc/draft-ietf-httpapi-ratelimit-headers/?utm_source=chatgpt.com "RateLimit header fields for HTTP - Datatracker - IETF"
[5]: https://www.ietf.org/archive/id/draft-polli-ratelimit-headers-02.html?utm_source=chatgpt.com "RateLimit Header Fields for HTTP"
