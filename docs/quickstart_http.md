# HTTP Quickstart

This guide is for teams integrating against a hosted RecSys deployment using only HTTP calls. Follow it when you cannot or do not want to run the repo locally.

---

## 1. Base URL, auth, and namespaces

- **Base URL:** obtain it from your deployment owner (e.g., `https://api.recsys.example.com`). In local demos it is `http://localhost:8000`.
- **Org header:** every request **must** include `X-Org-ID: <uuid>`. This header enforces tenancy and is validated server-side.
- **Auth (optional):** if `API_AUTH_ENABLED=true`, include `X-API-Key: <token>` or the configured `Authorization` header. Your ops team will tell you which scheme is live.
- **Namespace:** a logical tenant/surface bucket (`default`, `retail_us`, etc.). Supply it in every payload. Use different namespaces when you need isolated catalogs or experiments.

---

## 2. Ingest minimal data

These copy-paste examples assume:

```text
export BASE_URL="https://api.recsys.example.com"
export ORG_ID="00000000-0000-0000-0000-000000000001"
export NS="retail_demo"
```

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

## 3. Request recommendations (and similar items)

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

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| `400 missing_org_id` | `X-Org-ID` header absent or not a UUID | Send the header on every request, even for GETs. |
| `404 namespace_not_found` | Namespace typo or never seeded | Double-check the namespace string, seed data, or create it via admin APIs. |
| `422 invalid blend` | `overrides.blend` weights missing/negative | Normalize weights to positive floats; omit `overrides` to use defaults. |
| Empty recommendation list | Catalog empty or items `available=false` | Upsert items with `available=true`, check `/v1/items`. |
| Slow responses (>1s) | Too large `k`/`fanout`, network to region | Reduce `k`, avoid unnecessary `include_reasons`, ensure calls hit the closest region. |

---

## 5. Next steps

- **Full endpoint details:** `docs/api_reference.md`
- **Environment & algorithm knobs:** `docs/env_reference.md`
- **Concepts, metrics, guardrails primer:** `docs/concepts_and_metrics.md`
- **Business + lifecycle overview:** `docs/business_overview.md`, `docs/overview.md`
- **Simulation & guardrail workflows:** `docs/simulations_and_guardrails.md`

When you need to apply rules or debug ranking issues, jump to `docs/rules-runbook.md`. For local experiments or deeper tuning, follow `GETTING_STARTED.md` plus `docs/tuning_playbook.md`.
