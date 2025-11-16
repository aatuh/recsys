# Database Schema Guide

This guide summarizes the key Postgres tables Recsys uses, describing the
columns, types, and how each table powers the system. Use it when mapping your
catalog/users/events, troubleshooting audits, or exporting evidence.

> **Who should read this?** Integration engineers and developers who need to map existing data into Recsys (or export traces). Pair with `docs/api_reference.md` for the ingest APIs and `docs/env_reference.md` for tuning behaviour. Commands referencing `make` or `analysis/scripts` assume local repo access; hosted API-only readers can focus on `docs/quickstart_http.md`.

### TL;DR

- **Purpose:** Explain what lives in each Postgres table (items, users, events, traces, overrides, etc.) and how RecSys uses those columns.
- **Use this when:** You are mapping data pipelines, debugging ingestion/audit issues, or exporting evidence for stakeholders.
- **Outcome:** Clear guidance on required fields, useful SQL snippets, and tips for keeping namespaces/data healthy.
- **Not for:** Learning the API schema or high-level concepts—see `docs/api_reference.md` and `docs/concepts_and_metrics.md` instead.

## Catalog & User Tables

### `items`

- **`org_id` (uuid)** — Tenant identifier; rows scoped to org + namespace.
- **`namespace` (text)** — Logical namespace (e.g., `default`, `retail_us`).
- **`item_id` (text)** — Unique item key; primary key with org/namespace.
- **`title`, `description` (text)** — Optional metadata shown downstream.
- **`brand`, `category`, `category_path` (text/text[])** — Used for caps/overrides; `category_path` holds hierarchical breadcrumbs.
- **`price` (numeric)** — Powers ranking features (margin, etc.).
- **`available` (bool)** — Drives eligibility filters.
- **`tags` (text[])** — Consumed by rules engine and personalization.
- **`props` (jsonb)** — Schemaless attributes (margin, novelty hints).
- **`updated_at` (timestamptz)** — Auto-updated timestamp.

### `users`

- **`org_id`, `namespace`, `user_id` (uuid/text)** — Composite primary key.
- **`traits` (jsonb)** — Segments, locale, device, and other metadata.
- **`recent_activity_at` (timestamptz)** — Optional auditing timestamp.
- **`created_at`, `updated_at` (timestamptz)** — Lifecycle tracking.

### `events`

- **`org_id`, `namespace`, `event_id` (uuid/text)** — Primary key (`event_id` often generated client-side).
- **`user_id`, `item_id` (text)** — References to users/items (not enforced) required for personalization.
- **`type` (smallint)** — 0=view, 1=click, 2=add-to-cart, 3=purchase, 4=custom.
- **`ts` (timestamptz)** — Event timestamp.
- **`value` (double precision)** — Optional scalar (quantity, revenue).
- **`meta` (jsonb)** — Surface, session_id, campaign info, etc.

### `event_type_config`

Stores the weight/half-life for each event type per namespace. Used when building popularity and co-vis features.

## Merchandising & Overrides

### `rules`

- **`rule_id` (uuid)** — Primary key.
- **`org_id`, `namespace`, `surface` (uuid/text)** — Scope of the rule.
- **`action` (BOOST/PIN/BLOCK)** — What the rule does.
- **`target_type`, `target_key` (enums/text)** — Target item/tag/brand/category.
- **`priority` (int)** — Higher number wins on conflict.
- **`start_at`, `end_at` (timestamptz)** — Optional schedule window.
- **`boost_value` (float)** — Multiplier for boost actions.
- **`metadata` (jsonb)** — Free-form payload for auditing/UI.

### `manual_overrides`

Short-lived boosts/suppressions that compile to rules internally.

- **`override_id` (uuid)** — Primary key.
- **`action` (text: boost/suppress)** — Determines rule type.
- **`item_id` (text)** — Target item.
- **`boost_value` (float)** — Optional numeric value.
- **`expires_at` (timestamptz)** — TTL for the override.
- **`rule_id` (uuid)** — ID of the generated rule (traceability).

## Segments & Starter Profiles

### `segment_profiles`

- **`profile_id` (text)** — Identifier referenced in the API.
- **`blend_alpha/beta/gamma` (floats)** — Starter blend weights for pop/co-vis/emb.
- **`mmr_lambda` (float)** — Starter MMR setting.
- **`profile` (jsonb)** — Map of categories/tags to weights.

### `segments`

Defines rules-based cohorts (used for starter-profile guardrails and overrides).

- **`segment_id` (text)** — Identifier used in configs.
- **`description` (text)** — Human-readable explanation.
- **`criteria` (jsonb)** — Rule set evaluated at request time.

## Bandit Tables

### `bandit_policies`

- **`policy_id`, `name` (text)** — Unique policy identifiers.
- **`config` (jsonb)** — Arm definitions, weights, surfaces.
- **`is_active` (bool)** — Rollout toggle.
- **`created_at`, `updated_at` (timestamptz)** — Audit columns.

### `bandit_decisions` / `bandit_rewards`

Depending on the migration version, decisions and rewards may be stored directly in `rec_decisions` (see below) or in dedicated tables. Each decision holds `decision_id`, `policy_id`, `arm_id`, `context`, and timestamps; rewards reference `decision_id` + reward value.

## Decision Traces & Coverage

### `rec_decisions`

Audit table capturing the full recommendation context.

- **`decision_id` (uuid)** — Primary key referenced by audit APIs.
- **`org_id`, `namespace`, `surface` (uuid/text)** — Scope of the request.
- **`request_id`, `user_hash` (text)** — Client identifiers (hashed).
- **`k`, `constraints` (int/jsonb)** — Request parameters.
- **`effective_config` (jsonb)** — Algorithm config resolved after overrides.
- **`bandit` (jsonb)** — Optional bandit decision info.
- **`candidates_pre` (jsonb)** — Candidate lists before rules/MMR.
- **`final_items` (jsonb)** — Final ranked items.
- **`metrics` (jsonb)** — Coverage stats, leakage flags, guardrail checkpoints.

## Embedding Factors

### `recsys_item_factors`, `recsys_user_factors`

- **`item_id` / `user_id` (text)** — Identifiers matching the main tables.
- **`factors` (vector(384))** — ALS embeddings used by the retriever.
- **`updated_at` (timestamptz)** — Timestamp of last factor refresh.

## Usage Tips

- Seed data via `/v1/items:upsert`, `/v1/users:upsert`, `/v1/events:batch`; tables update immediately. Use `/v1/items`/`/v1/users`/`/v1/events` to audit what was stored.
- Namespaces partition all tables; avoid reusing namespace names across customers unless intentional.
- Decision traces (`rec_decisions`) are the source of truth for guardrail debugging and can be exported via `/v1/audit/decisions`.
- When running CI or simulations, `analysis/scripts/reset_namespace.py` wipes items/users/events but does not delete historical decisions/rules—clean them manually if needed.

For column-by-column definitions, search `api/migrations/*.up.sql`. This document highlights the tables most relevant to integrators and operations.

## Developer & Integration Guidance

### Seeding and verifying data

1. Use the ingestion APIs (`/v1/items:upsert`, `/v1/users:upsert`, `/v1/events:batch`) to load data; avoid writing directly to the tables.
2. Verify the stored rows via `/v1/items`, `/v1/users`, `/v1/events` or by querying the tables:

```sql
SELECT item_id, brand, category, available
FROM items
WHERE org_id = :org AND namespace = 'retail_us'
ORDER BY updated_at DESC
LIMIT 10;
```

3. For large seeding jobs, run `analysis/scripts/seed_dataset.py --fixture-path analysis/fixtures/customers/<customer>.json` and inspect `analysis/evidence/seed_segments.json` to confirm segment distributions.

### Troubleshooting guardrails

- Fetch decision traces from `rec_decisions` (or via `/v1/audit/decisions`) to inspect `effective_config`, `candidates_pre`, `final_items`, and `metrics`.

```sql
SELECT decision_id, metrics->'coverage' AS coverage, metrics->'leakage' AS leakage
FROM rec_decisions
WHERE org_id = :org
  AND namespace = 'retail_us'
  AND ts >= now() - interval '24 hours';
```

- Coverage guardrails use `metrics.coverage.*`; zero-effect overrides log in `metrics.policy`.

### Cleaning namespaces

- `analysis/scripts/reset_namespace.py` deletes rows from items/users/events via the APIs but does **not** remove rules, manual overrides, or decision traces. To fully purge a namespace:

```sql
DELETE FROM manual_overrides WHERE org_id = :org AND namespace = 'retail_us';
DELETE FROM rules WHERE org_id = :org AND namespace = 'retail_us';
DELETE FROM rec_decisions WHERE org_id = :org AND namespace = 'retail_us';
```

- After deleting, rerun seeding and simulations to rebuild coverage metrics.

### Schema evolution

- New columns/tables are added via `api/migrations/*.up.sql`. Run `make migrate-up` (or `docker-compose run api make migrate-up`) to apply them. For rollbacks, use the paired `.down.sql`.
- Keep fixtures/templates (under `analysis/fixtures/`) in sync with schema changes (e.g., new item props or user traits).

### Exporting data

- For reporting, use SELECT queries filtered by `org_id` + `namespace`. Example: export all events for a timeframe:

```sql
COPY (
  SELECT user_id, item_id, type, ts, meta
  FROM events
  WHERE org_id = :org
    AND namespace = 'retail_us'
    AND ts BETWEEN :start AND :end
) TO STDOUT WITH CSV HEADER;
```

- Decision traces can be bulk-exported the same way for offline audits.

### Summary

- Use the APIs for ingest/delete operations whenever possible.
- Query the tables for verification, analytics, and troubleshooting.
- Always include `org_id` + `namespace` filters to avoid leaking data between tenants.
- After schema or config changes, run the simulation suite to ensure guardrails still pass.
