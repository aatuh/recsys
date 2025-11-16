# HTTP Quickstart

This guide is for teams integrating against a hosted RecSys deployment using only HTTP calls. Follow it when you consume RecSys as a managed service and do **not** need to run the repo locally. If you prefer a story-driven tour before diving into reference material, read [`docs/zero_to_first_recommendation.md`](zero_to_first_recommendation.md) first and then return here for the exhaustive details. Want to run the entire stack from source? Switch to [`GETTING_STARTED.md`](../GETTING_STARTED.md) instead.

> **Where this fits:** Client integration.

## TL;DR

- Use this doc when you are integrating with a hosted RecSys API using HTTP only.
- **Part 1** shows the minimum three calls you need to prove connectivity and a working recommendation loop.
- **Part 2** covers making ingestion and error handling production-ready.
- **Part 3** introduces more advanced topics like personalization and guardrails—safe to read in Week 2+.

If you only want to prove that your base URL, org ID, and namespace are wired correctly, start with **Part 1 – Hello RecSys in 3 calls** and come back for the rest later.

---

## Part 1 – Hello RecSys in 3 calls

### 0. Hello RecSys in 3 calls

These three calls verify that you can reach the API, ingest a single item, and receive at least one recommendation.

Set common variables:

```bash
export BASE_URL="https://api.recsys.example.com"
export ORG_ID="00000000-0000-0000-0000-000000000001"
export NS="hello_demo"
```

1. **Check health**

```bash
curl -s "$BASE_URL/health" \
  -H "X-Org-ID: $ORG_ID"
```

Expected: HTTP 200 response with a small JSON body indicating the service is healthy.

2. **Upsert a single item**

```bash
curl -s -X POST "$BASE_URL/v1/items:upsert" \
  -H "Content-Type: application/json" \
  -H "X-Org-ID: $ORG_ID" \
  -d '{
        "namespace": "'"$NS"'",
        "items": [
          {
            "item_id": "hello_sku_1",
            "available": true,
            "tags": ["category:demo"],
            "props": { "title": "Hello RecSys item" }
          }
        ]
      }'
```

3. **Request minimal recommendations**

```bash
curl -s -X POST "$BASE_URL/v1/recommendations" \
  -H "Content-Type: application/json" \
  -H "X-Org-ID: $ORG_ID" \
  -d '{
        "namespace": "'"$NS"'",
        "k": 4
      }'
```

If you receive a non-empty `items` list, you have proven the full loop (health → ingest → recommend). A minimal successful response might look like:

```json
{
  "items": [
    { "item_id": "hello_sku_1", "score": 0.42 }
  ]
}
```

The rest of this doc focuses on improving quality, personalization, and safety.

---

## Part 2 – Base URL, auth, and ingestion

### 1. Base URL, auth, and namespaces

- **Base URL:** obtain it from your deployment owner (e.g., `https://api.recsys.example.com`). In local demos it is `http://localhost:8000`.
- **Org header:** every request **must** include `X-Org-ID: <uuid>`. This header enforces tenancy and is validated server-side.
- **Auth (optional):** if `API_AUTH_ENABLED=true`, include `X-API-Key: <token>` or the configured `Authorization` header. Your ops team will tell you which scheme is live.
- **Namespace:** a logical per-org/surface bucket (`default`, `retail_us`, etc.). Supply it in every payload. Use different namespaces when you need isolated catalogs or experiments.

---

### 2. Ingest minimal data

These copy-paste examples assume:

```text
export BASE_URL="https://api.recsys.example.com"
export ORG_ID="00000000-0000-0000-0000-000000000001"
export NS="retail_demo"
```

If you are still deciding how to map your catalog, users, and events into these shapes, read [`object_model_concepts.md`](object_model_concepts.md) first for guidance.

### 2.1 Upsert items

```bash
curl -X POST "$BASE_URL/v1/items:upsert" \
  -H "Content-Type: application/json" \
  -H "X-Org-ID: $ORG_ID" \
  -d '{
        "namespace": "'"$NS"'",
        "items": [
          {
            "item_id": "sku_123",
            "available": true,
            "price": 29.99,
            "tags": ["brand:acme", "category:fitness", "color:blue"],
            "props": { "title": "Acme Smart Bottle", "inventory": 56 }
          }
        ]
      }'
```

### 2.2 Upsert users (optional but recommended)

```bash
curl -X POST "$BASE_URL/v1/users:upsert" \
  -H "Content-Type: application/json" \
  -H "X-Org-ID: $ORG_ID" \
  -d '{
        "namespace": "'"$NS"'",
        "users": [
          {
            "user_id": "user_001",
            "traits": { "segment": "fitness_seekers", "loyalty_tier": "gold" }
          }
        ]
      }'
```

### 2.3 Batch events

```bash
curl -X POST "$BASE_URL/v1/events:batch" \
  -H "Content-Type: application/json" \
  -H "X-Org-ID: $ORG_ID" \
  -d '{
        "namespace": "'"$NS"'",
        "events": [
          {
            "user_id": "user_001",
            "item_id": "sku_123",
            "type": 0,
            "value": 1,
            "timestamp": "2024-05-01T12:00:00Z",
            "meta": { "surface": "home" }
          },
          {
            "user_id": "user_001",
            "item_id": "sku_987",
            "type": 3,
            "value": 1,
            "timestamp": "2024-05-03T15:04:00Z",
            "meta": { "surface": "pdp" }
          }
        ]
      }'
```

Events drive personalization, guardrails, and coverage metrics. Use your real event type codes; defaults are `0=view`, `1=cart`, `2=wishlist`, `3=purchase`.

---

## Part 3 – Request recommendations and go deeper

### 3. Request recommendations (and similar items)

### 3.1 Personalized feed

```bash
curl -s -X POST "$BASE_URL/v1/recommendations" \
  -H "Content-Type: application/json" \
  -H "X-Org-ID: $ORG_ID" \
  -d '{
        "namespace": "'"$NS"'",
        "user_id": "user_001",
        "k": 12,
        "include_reasons": true,
        "overrides": {
          "blend": { "alpha": 0.45, "beta": 0.35, "gamma": 0.20 },
          "mmr": { "lambda": 0.18 }
        }
      }'
```

Key fields:

- `k`: number of items to return.
- `surface`: optional label (home, pdp, email) that ties into guardrails and overrides.
- `include_reasons`: adds `trace` + per-item reason codes for debugging.
- `overrides`: per-call tweaks to blend/MMR/profile knobs without changing env vars.

### 3.2 Similar items

```bash
curl -s "$BASE_URL/v1/items/sku_123/similar?namespace=$NS&k=8" \
  -H "X-Org-ID: $ORG_ID"
```

Useful for PDP “you may also like” modules. This endpoint relies heavily on embeddings and co-visitation data; ensure your catalog and events cover the source item.

---

## 4. Common mistakes

- **`400 missing_org_id`** — `X-Org-ID` header absent or not a UUID. Fix: send the header on every request, even GETs.
- **`404 namespace_not_found`** — Namespace typo or never seeded. Fix: double-check the namespace string, seed data, or create it via admin APIs.
- **`422 invalid blend`** — `overrides.blend` weights missing or negative. Fix: normalize weights to positive floats; omit `overrides` to use defaults.
- **Empty recommendation list** — Catalog empty or items `available=false`. Fix: upsert items with `available=true` and check `/v1/items`.
- **Slow responses (>1s)** — `k`/fanout too large or wrong region. Fix: reduce `k`, skip unnecessary `include_reasons`, ensure calls hit the nearest region.

For more scenarios and step-by-step debugging tips, see `docs/faq_and_troubleshooting.md`.

---

## 5. Next steps

- **Full endpoint details:** `docs/api_reference.md`
- **Language examples:** `docs/client_examples.md` (Python + Node.js snippets)
- **Errors & limits:** `docs/api_errors_and_limits.md` (status codes, payload limits, retry guidance)
- **FAQ & troubleshooting:** `docs/faq_and_troubleshooting.md` when something doesn’t work as expected
- **Environment & algorithm knobs:** `docs/env_reference.md`
- **Concepts, metrics, guardrails primer:** `docs/concepts_and_metrics.md`
- **Business + lifecycle overview:** `docs/business_overview.md`, `docs/overview.md`, `docs/system_overview.md`, `docs/doc_map.md`
- **Simulation & guardrail workflows:** `docs/simulations_and_guardrails.md`

When you need to apply rules or debug ranking issues, jump to `docs/rules_runbook.md`. For local experiments or deeper tuning, follow `GETTING_STARTED.md` plus `docs/tuning_playbook.md`.

---

## Where to go next

- If you’re integrating HTTP calls → keep `docs/api_reference.md` and `docs/client_examples.md` nearby.
- If you’re a PM → skim `docs/business_overview.md` and `docs/recsys_in_plain_language.md`.
- If you’re tuning quality → read `docs/tuning_playbook.md` and `docs/simulations_and_guardrails.md`.
