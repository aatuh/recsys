# API Reference

This reference groups every public REST endpoint by domain. For payload schemas see `api/swagger/swagger.yaml`; this document explains what each route is for, who uses it, and notable parameters or behaviors.

> **Where this fits:** Client integration.
>
> **Who should read this?** Integration engineers and developers implementing against the API. Pair it with [`docs/env_reference.md`](env_reference.md) for algorithm knobs and [`docs/database_schema.md`](database_schema.md) for storage details.

## TL;DR

- Use this doc when you already know the basics from [`docs/quickstart_http.md`](quickstart_http.md) and need per-endpoint details.
- It is organized by domain (ingestion, ranking, admin, audit) so you can jump directly to the routes you care about.
- On first read, focus on ingestion and `/v1/recommendations`; you can skip admin and audit sections until you are wiring rules or debugging incidents.

## Ingestion & Data Management

- **`POST /v1/items:upsert`** — Bulk insert/update catalog items (50 per request). Include `tags`, `props`, `available`; writes to Postgres and refreshes caches.
- **`POST /v1/users:upsert`** — Bulk insert/update user profiles (≤100 per request). Provide `traits` JSON with segments, locale, etc.
- **`POST /v1/events:batch`** — Record behavioral events (view/click/purchase). Up to 500 events per batch; `type` codes 0=view…3=purchase.
- **`POST /v1/items:delete`** — Delete items by namespace/ID filter. Supply `delete_request.namespace` plus optional `item_id`.
- **`POST /v1/users:delete`** — Delete users by namespace/filter. Handy when resetting a namespace for a given org.
- **`POST /v1/events:delete`** — Delete historical events for a namespace. Supports filters (`user_id`, `event_type`, time ranges).
- **`GET /v1/items`** — Paginated list of items with filters (ID, created_after/before). Useful for QA/audits.
- **`GET /v1/users`** — Paginated list of users, filterable by ID or creation timestamps.
- **`GET /v1/events`** — Paginated events feed with filters (`user_id`, `item_id`, `event_type`, time window).

**Common request fields**

- `namespace` (string, required): logical bucket under an org. Every request must supply it explicitly.
- `items[]/users[]/events[]`: see Swagger for full schema; recommended properties:
  - Items: `item_id`, `category`, `brand`, `price`, `available`, `tags[]`, `props{}`
  - Users: `user_id`, `traits{ segment, locale, device }`
  - Events: `user_id`, `item_id`, `type`, `ts`, optional `meta{ surface, session_id }`
- Deletion payloads support optional filters (`created_after`, `user_id`, etc.). Omitting filters deletes everything for the namespace.

```bash
curl -X POST https://api.example.com/v1/items:upsert \
  -H "Content-Type: application/json" \
  -d '{
        "namespace": "retail_us",
        "items": [
          {
            "item_id": "sku_123",
            "category": "Beauty",
            "brand": "PureBloom",
            "price": 19.99,
            "available": true,
            "tags": ["beauty","skincare","brand:purebloom"],
            "props": {"margin":0.35,"novelty":0.4}
          }
        ]
      }'
```

## Ranking & Explainability

- **`POST /v1/recommendations`** — Core ranking endpoint returning top-K items. Requires `namespace`, `k`, optional `user_id`. Supports `overrides` (blend/MMR/profile knobs) and `include_reasons`; returns detailed traces for guardrails.
- **`POST /v1/rerank`** — Re-score a caller-supplied candidate list (≤200 items). Reuses the same personalization and telemetry but never injects new IDs, so search/cart services keep control over retrieval.
- **`GET /v1/items/{item_id}/similar`** — Fetch similar items via collaborative/content signals. Supply `namespace` + `item_id`; powers “related products.”
- **`POST /v1/explain/llm`** — Ask the explainer LLM for narrative summaries (`recommendation`, `order`). Provide question + time window; requires LLM env vars.

**Key request fields (`/v1/recommendations`)**

- `namespace` (required) and `k` (required, typical values 10–50).
- `surface`: informs guardrails (e.g., `home`, `search`, `email`).
- User context: `user_id`, `recent_event_ids`, or `recent_item_ids`.
- `constraints.include/exclude`: limit to specific tags, brands, or explicit lists.
- `overrides`: temporary tuning knobs (blend weights, `mmr_lambda`, `profile_boost`, `profile_starter_blend_weight`, `brand_cap`, `category_cap`, etc.).
- `include_reasons` + `explain_level`: request textual reasons or numeric breakdowns for debugging.

```bash
curl -X POST https://api.example.com/v1/recommendations \
  -H "Content-Type: application/json" \
  -d '{
        "namespace": "retail_us",
        "surface": "home",
        "user_id": "user_42",
        "k": 20,
        "overrides": {
          "mmr_lambda": 0.35,
          "profile_starter_blend_weight": 0.45
        },
        "include_reasons": true
      }'
```

### Rerank workflow

- **When to use it:** downstream systems already have a candidate list (search/browse/cart) but want the same personalization, policy, and telemetry as `/v1/recommendations`.
- **Payload:** identical top-level fields plus an `items[]` array (≤200 entries) supplying `item_id` and optional `score`. The service never injects new IDs; it only reorders what you send.
- **Response:** same schema as `/v1/recommendations` (items + `trace.extras`, audit ID) so guardrail automation and evidence capture work unchanged.

```bash
curl -X POST https://api.example.com/v1/rerank \
  -H "Content-Type: application/json" \
  -d '{
        "namespace": "retail_us",
        "user_id": "shopper_42",
        "k": 4,
        "context": {"surface": "search", "query": "wireless earbuds"},
        "items": [
          {"item_id": "sku_a", "score": 0.71},
          {"item_id": "sku_b", "score": 0.65},
          {"item_id": "sku_c"}
        ]
      }'
```

## Behavioral guarantees

- **Idempotency**
  - `POST /v1/items:upsert`, `POST /v1/users:upsert`, and `POST /v1/events:batch` overwrite existing rows based on `item_id` / `user_id` / `event_id`. Resending the same payload is safe and acts as an upsert.
  - Rules/overrides are not idempotent by name—use unique IDs or fetch existing resources before recreating.
- **Consistency & freshness**
  - Ingested items/users/events typically show up in `/v1/recommendations` within seconds; caches refresh automatically. Heavy batch jobs can take up to ~1 minute to propagate retriever signals.
  - Derived components (embedding retrievers, bandit policies) depend on nightly/batch pipelines—consult the relevant analytics docs for expected lag.
- **Pagination**
  - `GET /v1/items`, `/v1/users`, `/v1/events` use cursor-based pagination (`page_token`). Responses include `next_page_token`; pass it to fetch the next page.
  - Default sort is `updated_at desc`. Replaying a request with the same `page_token` yields stable results unless data changed.
  - Max page size is 100 records.

## Bandit (Multi-arm) API

> Need a refresher on what “multi-armed bandit” means? See the plain-language definition in `docs/concepts_and_metrics.md`.

- **`POST /v1/bandit/decide`** — Allocate a user/session into an experiment arm. Call before rendering; response includes `arm_id` and `policy_version`.
- **`POST /v1/bandit/recommendations`** — Retrieve recommendations with exploration baked in. Same payload as `/v1/recommendations`, plus bandit metadata in the response.
- **`POST /v1/bandit/reward`** — Send reward signals (click/purchase) for an arm. Provide `decision_id`, `reward` (0–1), optional metadata.
- **`GET /v1/bandit/policies`** — List policies for an org/namespace, including active + historical arm configs.
- **`POST /v1/bandit/policies:upsert`** — Create or update policy definitions (arms, traffic splits, eligibility rules). Used when launching new tests.

**Typical sequence**

1. `POST /v1/bandit/decide` with `namespace`, `surface`, optional traits → response includes `decision_id`, `arm_id`, `policy_version`.
2. Render the surface using `/v1/bandit/recommendations` (bandit-managed blend) or `/v1/recommendations` (pass `arm_id` in `context` to pick a preset).
3. When the user acts, `POST /v1/bandit/reward` with the `decision_id` and `reward` (0/1 or scaled value). Rewards update the policy’s arm statistics.

Policies are defined via `/v1/bandit/policies:upsert`. Each policy lists arms, traffic percentages, and optional eligibility filters (segments, namespaces, surfaces). Use `GET /v1/bandit/policies` to audit what’s live.

## Configuration (Event Types, Segments, Presets)

- **`GET /v1/event-types`** — List current event-type weights and half-lives.
- **`POST /v1/event-types:upsert`** — Configure event-type weights/half-lives (e.g., change purchase weight, disable custom events).
- **`GET /v1/segments`** — List behavioral segments driving cohort-specific tuning.
- **`POST /v1/segments:upsert`** — Create/update segments (`segment_id`, description, eligibility rules).
- **`POST /v1/segments:delete`** — Remove segments when retiring cohorts.
- **`GET /v1/segment-profiles`** — Fetch preset starter profiles per segment for cold-start curation.
- **`POST /v1/segment-profiles:upsert`** — Create/update starter profile weights (map categories/tags to weights).
- **`POST /v1/segment-profiles:delete`** — Delete starter profiles.
- **`POST /v1/segments:dry-run`** — Test segment definitions without saving.
- **`GET /v1/admin/recommendation/presets`** — Fetch recommended MMR presets per surface for UI tooling.
- **`GET/POST /v1/admin/recommendation/config`** — Fetch/apply the active recommendation config. Use `analysis/scripts/recommendation_config.py` to manage git-backed JSON templates.

**Usage tips**

- Start with `GET` endpoints to inspect defaults after provisioning a namespace.
- Use the `:dry-run` endpoints (`segments:dry-run`, rules dry-run) before committing changes to avoid breaking guardrails.
- Starter profiles feed cold-start personalization; align them with the data seeded via fixtures/templates.
- Event-type weights/half-lives should mirror the importance of downstream KPIs (e.g., purchases > clicks). Adjust them in tandem with guardrail thresholds.

## Rules & Manual Overrides

- **`GET/POST /v1/admin/rules`** — List or create merchandising rules (boost/block/pin). POST body defines targets, actions, priority, namespace/surface.
- **`GET/PUT/DELETE /v1/admin/rules/{rule_id}`** — Inspect, update, or delete a specific rule. Use PUT to adjust windows/priority.
- **`POST /v1/admin/rules/dry-run`** — Test a rule against synthetic input; returns what would happen without saving.
- **`GET/POST /v1/admin/manual_overrides`** — Manage ad-hoc overrides (short-lived boosts/pins). Overrides compile to rules behind the scenes.
- **`POST /v1/admin/manual_overrides/{override_id}/cancel`** — Cancel an active manual override.

Rules are long-lived merchandising controls. Manual overrides map to temporary rules behind the scenes. Both obey namespace/surface scoping and appear in decision traces (`trace.extras.policy`). Always dry-run complex rules before enabling them in production.

### Rule testing & evidence

1. `POST /v1/admin/rules/dry-run` (or run `analysis/scripts/test_rules.py --base-url <url> --org-id <uuid> --namespace <ns>`) to see how the new rule would modify a seeded namespace. The script writes before/after payloads plus metric samples to `analysis/results/rules_effect_sample.json`.
2. Inspect the dry-run response (`preview.items`, `policy.preview`) and verify the expected rule IDs show up under `trace.extras.policy.rules`.
3. Enable the rule with `POST /v1/admin/rules` and watch `/metrics` for `policy_rule_blocked_items_total{rule_id="<id>"}` and `policy_constraint_leak_total` so you can prove the change behaves safely.

Include the JSON artifacts when filing guardrail evidence or peer reviews.

## Audit & Coverage

- **`GET/POST /v1/audit/decisions`** — List audit records or enqueue new ones. GET supports namespace/time filters; POST is used internally during tracing.
- **`GET /v1/audit/decisions/{decision_id}`** — Fetch a specific audit record, including request, config, response, policy summary.
- **`POST /v1/audit/search`** — Query audits with richer filters (rule IDs, leakage flags, user IDs, time windows, surfaces).

Traces include the full request, resolved algorithm config, policy summaries, and coverage telemetry. Use them to debug guardrail failures or zero-effect overrides.

## Data Governance / Maintenance

- **`GET /version`** — Emit git commit, build timestamp, and model version (from `RECSYS_GIT_COMMIT` / `RECSYS_BUILD_TIME`; falls back to runtime defaults).
- **`GET /health`** — Liveness/readiness probe returning `{ "status": "ok" }`; used by Docker/CI.
- **`GET /docs`** — Swagger UI / API docs (serves `swagger.json` / `swagger.yaml`).
- **`GET /metrics`** — Prometheus metrics (when observability env vars enabled) covering `policy_*`, HTTP latency, DB stats, etc.

Pair `/version` with determinism artifacts (see `analysis/results/determinism_ci.json` or `make determinism`) when capturing evidence for evaluations or incident reviews.

### Load & chaos workflows

> Local-only note: the commands below assume you cloned this repo and can run Docker/Make. Hosted API consumers can skip this section.

- `make load-test LOAD_BASE_URL=<url> LOAD_ORG_ID=<uuid> LOAD_NAMESPACE=<ns> LOAD_RPS=10,100,1000` – runs the k6 script `analysis/load/recommendations_k6.js`, ramps through each RPS stage, and writes `analysis/results/load_test_summary.json` (latency percentiles, iteration count, error rate).
- `LOAD_USER_POOL`, `LOAD_SURFACE`, and `LOAD_STAGE_DURATION` let you tailor the traffic mix without editing code; set `SUMMARY_PATH` to capture multiple environments (e.g., staging vs. prod) side by side.
- `python analysis/scripts/chaos_toggle.py api pause 15` (or `db stop 20`) temporarily pauses a docker-compose service so you can observe how `/v1/recommendations` behaves during cache/DB outages while the load test is running.
- Share the resulting JSON summaries alongside determinism and scenario evidence when answering evaluation rubrics.

## Error handling & status codes

- **`200 OK`** — Successful write/read (responses include `trace_id`). Nothing to fix.
- **`400 Bad Request`** — Missing `namespace`, malformed payload, missing `X-Org-ID`. Fix headers and JSON shape.
- **`401/403 Unauthorized`** — API key missing/invalid when auth enabled. Add `X-API-Key` or `Authorization`.
- **`404 Not Found`** — Namespace absent, item/user not seeded, or endpoint typo. Confirm namespace spelling and ingestion.
- **`409 Conflict`** — Duplicate IDs (manual override/rule creation). Fetch existing resource, resolve conflict, retry.
- **`422 Unprocessable Entity`** — Invalid override values (`overrides.blend` sums to zero, unknown event type, etc.). Refer to `docs/env_reference.md` for valid ranges.
- **`429 Too Many Requests`** — Rate limit exceeded (`API_RATE_LIMIT_RPM`). Back off or request higher limits.
- **`500 Internal Server Error`** — Unexpected bug or downstream outage. Retry with jitter; capture `trace_id` and escalate.

All errors return JSON with `code`, `message`, and sometimes `details`. Include the `trace_id` when reporting incidents.

## Common patterns

- **Recommendations vs rerank:** Use `/v1/recommendations` when the service can assemble candidates from its own data; use `/v1/rerank` when your application already has a candidate list but wants consistent blend/MMR/personalization. Both return the same trace schema, so guardrails and audits work identically.
- **Minimal ingestion loop:** Follow [`docs/quickstart_http.md`](quickstart_http.md) for the canonical order (items → users → events → recommendations). Missing steps are the #1 cause of empty lists.
- **Per-surface overrides:** Pass `context.surface` (home, pdp, email) and supply `overrides.blend`/`overrides.mmr` per request. When an override sticks, capture it in env profiles via `analysis/scripts/env_profile_manager.py`.
- **Namespace resets:** Use `analysis/scripts/reset_namespace.py` → `seed_dataset.py` before tuning or running scenarios to eliminate leftover catalog noise.
- **Rules dry-run workflow:** `POST /v1/admin/rules/dry-run` with the proposed payload, review the before/after samples, then create the rule. [`docs/rules_runbook.md`](rules_runbook.md) explains how to monitor telemetry afterward.
- **Error triage loop:** Capture `trace_id`, query `/v1/audit/decisions/{trace_id}`, and cross-reference with guardrail dashboards. Refer to [`docs/concepts_and_metrics.md`](concepts_and_metrics.md) to explain coverage/diversity terminology in reports.

Need more narrative examples? Start with [`GETTING_STARTED.md`](../GETTING_STARTED.md) for local walkthroughs or [`docs/quickstart_http.md`](quickstart_http.md) for hosted integrations. For detailed error codes, limits, and retry guidance, see [`docs/api_errors_and_limits.md`](api_errors_and_limits.md).

## Using the reference

- Payload schemas: see `api/swagger/swagger.yaml` or the generated docs served at `/docs`.
- Authentication: by default API keys are disabled (`API_AUTH_ENABLED=false`). If enabled, set `X-API-Key` header.
- Namespacing: every write/read call requires `namespace` (explicit field or inferred from user/item). Guardrails and env profiles apply per namespace.
- Rate limiting: configure via `API_RATE_LIMIT_RPM`/`BURST`; admins should mention limits in partner docs once keys are issued.
- Security & data handling: see [`docs/security_and_data_handling.md`](security_and_data_handling.md) for TLS/auth, PII guidance, and retention expectations.
For deeper walkthroughs and examples, see:
- [`GETTING_STARTED.md`](../GETTING_STARTED.md) / [`docs/quickstart_http.md`](quickstart_http.md) (hands-on ingestion + recommendation flow)
- [`docs/simulations_and_guardrails.md`](simulations_and_guardrails.md) (seeding + simulation)
- [`docs/rules_runbook.md`](rules_runbook.md) (override troubleshooting)
- [`docs/concepts_and_metrics.md`](concepts_and_metrics.md) (terminology used in traces/guardrails)
