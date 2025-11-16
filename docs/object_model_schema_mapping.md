# Object Model – Schema Mapping (Advanced)

This doc explains how core objects (Org, Namespace, Item, User, Event) map onto RecSys storage and database schemas.

> ⚠️ **Advanced topic**
>
> Read this after you have a basic integration working via `docs/quickstart_http.md` and you are comfortable querying the database.
>
> **Where this fits:** Ingestion & storage.

---

## 1. Items

- Stored primarily in the `items` table (see `docs/database_schema.md#items`).
- Important columns:
  - `org_id` (uuid), `namespace` (text), `item_id` (text) — composite key.
  - `available` (bool), `brand`, `category`, `category_path`.
  - `tags` (text[]), `props` (jsonb).

The `items:upsert` API writes to these columns; rules, caps, and guardrails read from them.

---

## 2. Users

- Stored in the `users` table (`docs/database_schema.md#users`).
- Important columns:
  - `org_id`, `namespace`, `user_id`.
  - `traits` (jsonb).
  - `recent_activity_at`, `created_at`, `updated_at`.

User traits are mirrored from the API and used in rules and segment definitions.

---

## 3. Events

- Stored in the `events` table (`docs/database_schema.md#events`).
- Important columns:
  - `org_id`, `namespace`, `event_id`.
  - `user_id`, `item_id`, `type`, `ts`, `value`, `meta`.

These rows feed popularity, co-visitation, and personalization signals (see `docs/concepts_and_metrics.md`).

---

## 4. Org & Namespace

- `org_id` appears on all core tables and ties back to `X-Org-ID`.
- `namespace` buckets records into logical groups per org.

Guardrails and simulations usually select subsets of data using both `org_id` and `namespace` filters.

---

## 5. Decisions & guardrail evidence

- Decision traces live in `rec_decisions` (`docs/database_schema.md#rec_decisions`).
- Fields like `final_items`, `metrics`, and `effective_config` are used when debugging guardrail failures or tuning.

Use this doc as a complement to `object_model_concepts.md` and `docs/database_schema.md` when you need column-level understanding.

