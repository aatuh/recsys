# Zero to First Recommendation

This narrative walks through a fictional “Acme Outfitters” storefront from zero data to the first recommendation response. Follow it when you need a story-oriented tour before diving into the detailed quickstarts or API reference.

> Already comfortable with the concepts? Jump to `docs/quickstart_http.md` for the full hosted API guide or `GETTING_STARTED.md` if you want to run the stack locally.

---

## 1. Scenario Setup

Acme Outfitters sells outdoor gear. We’ll work inside the namespace `acme_demo` and use the hosted API endpoint `https://api.example.com`.

We only need a handful of records to see meaningful output:

- **Items** – a hiking backpack and a rain jacket.
- **User** – Sam, an outdoors enthusiast.
- **Events** – Sam views and purchases the backpack.

All requests include the `X-Org-ID` header (tenant UUID) and the `namespace` field so data stays isolated.

---

## 2. Ingest Items

```bash
curl -X POST https://api.example.com/v1/items:upsert \
  -H "Content-Type: application/json" \
  -H "X-Org-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
        "namespace": "acme_demo",
        "items": [
          {
            "item_id": "pack_01",
            "title": "SummitPro 45L Hiking Pack",
            "category": "Packs",
            "brand": "TrailSmith",
            "price": 159.99,
            "available": true,
            "tags": ["category:packs", "brand:trailsmith", "activity:hiking"],
            "props": { "capacity_liters": 45 }
          },
          {
            "item_id": "jacket_01",
            "title": "StormGuard Rain Shell",
            "category": "Outerwear",
            "brand": "TrailSmith",
            "price": 189.99,
            "available": true,
            "tags": ["category:outerwear", "brand:trailsmith", "activity:hiking"]
          }
        ]
      }'
```

- `item_id` can be any unique string; we often reuse the catalog SKU.
- `tags` power rules, personalization, and diversity guardrails.

---

## 3. Ingest a User

```bash
curl -X POST https://api.example.com/v1/users:upsert \
  -H "Content-Type: application/json" \
  -H "X-Org-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
        "namespace": "acme_demo",
        "users": [
          {
            "user_id": "sam_0001",
            "traits": {
              "segment": "trail_seekers",
              "location": "PNW",
              "loyalty_tier": "gold"
            }
          }
        ]
      }'
```

- `traits.segment` is optional but makes guardrail dashboards more informative.

---

## 4. Record Behavior

```bash
curl -X POST https://api.example.com/v1/events:batch \
  -H "Content-Type: application/json" \
  -H "X-Org-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
        "namespace": "acme_demo",
        "events": [
          {
            "user_id": "sam_0001",
            "item_id": "pack_01",
            "type": 0,
            "value": 1,
            "timestamp": "2025-01-10T12:00:00Z",
            "meta": { "surface": "home" }
          },
          {
            "user_id": "sam_0001",
            "item_id": "pack_01",
            "type": 3,
            "value": 1,
            "timestamp": "2025-01-10T12:15:00Z",
            "meta": { "surface": "checkout" }
          }
        ]
      }'
```

- `type` codes: 0 = view, 1 = click, 2 = add-to-cart, 3 = purchase.
- The `meta.surface` field is useful later when investigating guardrails (`docs/simulations_and_guardrails.md`).

---

## 5. Request Recommendations

```bash
curl -s -X POST https://api.example.com/v1/recommendations \
  -H "Content-Type: application/json" \
  -H "X-Org-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
        "namespace": "acme_demo",
        "surface": "home",
        "user_id": "sam_0001",
        "k": 8,
        "include_reasons": true
      }' | jq
```

Example response (shortened):

```json
{
  "items": [
    {
      "item_id": "jacket_01",
      "score": 0.87,
      "reasons": ["personalization", "co_visitation"]
    },
    {
      "item_id": "pack_01",
      "score": 0.65,
      "reasons": ["recent_purchase", "diversity"]
    }
  ],
  "trace": {
    "namespace": "acme_demo",
    "surface": "home",
    "policy": "blend:v1_mmr",
    "extras": {
      "candidate_sources": {
        "popularity": 120,
        "co_visitation": 85,
        "embedding": 60
      },
      "personalized_items": 4
    }
  }
}
```

**How to read it**

- `k` is the list length; tune it per surface. Larger `k` means more compute.
- `score` is a blended value (see `docs/concepts_and_metrics.md` for how popularity, co-visitation, and embeddings combine).
- `reasons` tell you why an item appeared. They’re designed for merchandising reviews and guardrail evidence.
- `trace.extras.candidate_sources` show how many candidates each retriever contributed before ranking.

---

## 6. Connecting to Guardrails & Rules

- Before rollout, run `make scenario-suite` (documented in `docs/simulations_and_guardrails.md`) to ensure cold-start, rule, and exposure guardrails pass for your namespace.
- Use the rule engine (`/v1/admin/rules` and `/v1/admin/manual_overrides`) to pin/boost items when campaigns require extra control; see `docs/rules_runbook.md`.

---

## 7. Where to Go Next

1. **Hosted quickstart** – `docs/quickstart_http.md` for a comprehensive reference with troubleshooting.
2. **API reference** – `docs/api_reference.md` for every endpoint, error codes, and behavioral guarantees.
3. **Safety & simulations** – `docs/simulations_and_guardrails.md` to understand guardrails, fixtures, and CI.
4. **Local development** – `GETTING_STARTED.md` if you want to run RecSys from source.
