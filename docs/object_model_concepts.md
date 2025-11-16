# Object Model – Concepts

This doc explains how RecSys sees your items, users, events, orgs, and namespaces in business and API terms.

> **Who should read this?** Week 1 integrators and PMs mapping their catalog and user base into RecSys.
>
> **Where this fits:** Ingestion & storage.

---

## TL;DR

- Understand the core objects: **Org, Namespace, Item, User, Event, Surface**.
- See how they relate and how they appear in API calls.
- Use this when designing your initial data mapping; skip schema/DB details on first pass.

---

## Contents

- [Org & Namespace](#org--namespace)
- [Items](#items)
- [Users](#users)
- [Events](#events)
- [Surfaces / placements](#surfaces--placements)
- [How this maps to API calls](#how-this-maps-to-api-calls)

---

## Org & Namespace

- **Org** – the customer/company identifier (`X-Org-ID`, `org_id`).
- **Namespace** – a logical bucket under an org (for example, per surface or region).

Typical patterns:

- One org per customer, with separate namespaces for surfaces (`home`, `pdp`, `email`).
- One org per business unit, with namespaces for regions (`retail_us`, `retail_eu`).

Conceptually:

```text
Org (X-Org-ID)
 ├─ Namespace: retail_us
 │   ├─ items
 │   ├─ users
 │   └─ events
 └─ Namespace: retail_eu
     ├─ items
     ├─ users
     └─ events
```

Guardrails, rules, and simulations are configured per org + namespace pair.

---

## Items

Items are the things you recommend: products, listings, shows, articles.

Minimal `items:upsert` payload:

```json
{
  "namespace": "retail_demo",
  "items": [
    {
      "item_id": "sku_123",
      "available": true,
      "price": 29.99,
      "tags": ["brand:acme", "category:fitness", "color:blue"],
      "props": {
        "title": "Acme Smart Bottle",
        "inventory": 56
      }
    }
  ]
}
```

- `item_id` – your stable product/content ID.
- `available` – whether the item can be shown.
- `tags` – simple labels like `category:*` or `brand:*` powering rules and guardrails.
- `props` – free-form attributes for traces and UIs.

---

## Users

Users represent the audiences you recommend to.

Minimal `users:upsert` payload:

```json
{
  "namespace": "retail_demo",
  "users": [
    {
      "user_id": "user_001",
      "traits": {
        "segment": "fitness_seekers",
        "loyalty_tier": "gold",
        "country": "FI"
      }
    }
  ]
}
```

- `user_id` – your stable user/customer/account identifier.
- `traits` – attributes you care about for analysis or rules (segments, country, tier).

---

## Events

Events connect users to items and provide the signals for popularity and personalization.

Minimal `events:batch` payload:

```json
{
  "namespace": "retail_demo",
  "events": [
    {
      "user_id": "user_001",
      "item_id": "sku_123",
      "type": 0,
      "value": 1,
      "timestamp": "2024-05-01T12:00:00Z",
      "meta": { "surface": "home" }
    }
  ]
}
```

- `user_id` / `item_id` – link back to your users and items.
- `type` – event kind (0=view, 1=click, 2=add-to-cart, 3=purchase in the default mapping).
- `value` – optional numeric weight (quantity, revenue).
- `meta.surface` – where the event happened (`home`, `pdp`, `search`, `email`).

---

## Surfaces / placements

Surfaces (sometimes called “placements”) describe **where** recommendations appear:

- `home` – home page feed.
- `pdp` – product detail page widgets.
- `search` – search result re-ranking.
- `email` – outbound campaigns.

You typically:

- Include a `surface` field in recommendation requests.
- Use namespaces and guardrails to tune behavior per surface.

---

## How this maps to API calls

- Ingestion:
  - `/v1/items:upsert` — sends items scoped by `namespace`.
  - `/v1/users:upsert` — sends users.
  - `/v1/events:batch` — sends events that connect users and items.
- Ranking:
  - `/v1/recommendations` — uses the items, users, and events for a given org + namespace and surface to build a ranked list.

For schema-level details (tables, columns, indexes), see [`object_model_schema_mapping.md`](object_model_schema_mapping.md) and [`docs/database_schema.md`](database_schema.md).

---

## Where to go next

- If you’re integrating HTTP calls → see [`docs/quickstart_http.md`](quickstart_http.md).
- If you’re a PM → skim [`docs/business_overview.md`](business_overview.md).
- If you’re tuning quality → read [`docs/tuning_playbook.md`](tuning_playbook.md).
