# API Reference

This reference groups every public REST endpoint by domain. For payload schemas see `api/swagger/swagger.yaml`; this document explains what each route is for, who uses it, and notable parameters or behaviors.

> **Who should read this?** Integration engineers and developers implementing against the API. Pair it with `docs/env_vars.md` for algorithm knobs and `docs/database_schema.md` for storage details.

## Ingestion & Data Management

| Endpoint            | Method | Purpose                                         | Notes                                                                                                                  |
|---------------------|--------|-------------------------------------------------|------------------------------------------------------------------------------------------------------------------------|
| `/v1/items:upsert`  | POST   | Bulk insert/update catalog items.               | Accepts `items` array (50 per request). Include `tags`, `props`, `available`. Writes to Postgres and refreshes caches. |
| `/v1/users:upsert`  | POST   | Bulk insert/update user profiles.               | Provide `traits` JSON with segments, locale, etc. Up to 100 per request.                                               |
| `/v1/events:batch`  | POST   | Record behavioral events (view/click/purchase). | Up to 500 per batch. `type` codes: 0=view…3=purchase. Drives personalization & bandits.                                |
| `/v1/items:delete`  | POST   | Delete items by namespace / ID filter.          | Pass `delete_request.namespace` plus optional `item_id`.                                                               |
| `/v1/users:delete`  | POST   | Delete users by namespace / filter.             | Use when resetting a tenant namespace.                                                                                 |
| `/v1/events:delete` | POST   | Delete historical events in a namespace.        | Supports filters such as `user_id`, `event_type`, time ranges.                                                         |
| `/v1/items`         | GET    | Paginated list of items.                        | Supports filters (`item_id`, created_after/before). Useful for QA and audits.                                          |
| `/v1/users`         | GET    | Paginated list of users.                        | Filter by `user_id` or creation timestamps.                                                                            |
| `/v1/events`        | GET    | Paginated list of events.                       | Filter by `user_id`, `item_id`, `event_type`, time window.                                                             |

**Common request fields**

- `namespace` (string, required): logical tenant. Every request must supply it explicitly.
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

| Endpoint                      | Method | Purpose                                                 | Notes                                                                                                                                                                                                                    |
|-------------------------------|--------|---------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `/v1/recommendations`         | POST   | Core ranking endpoint returning top-K items for a user. | Requires `namespace`, `k`, optional `user_id`. Supports `overrides` (blend/MMR/profile knobs) and `include_reasons`. Returns trace extras (policy summary, starter profile) and feeds audit/coverage guardrails.         |
| `/v1/rerank`                  | POST   | Re-score a caller-supplied candidate list.              | Send up to ~200 `items[]` (optional prior scores) plus user/context. Engine reuses blend/MMR/policy/personalization but never injects new IDs, so search/cart services can control retrieval while inheriting telemetry. |
| `/v1/items/{item_id}/similar` | GET    | Fetch similar items by collaborative/content signals.   | Needs `namespace` and `item_id`. Used for “related products.”                                                                                                                                                            |
| `/v1/explain/llm`             | POST   | Ask the LLM explainer for narrative summaries.          | Provide target type (`recommendation`, `order`), time window, question. Requires LLM env vars.                                                                                                                           |

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

## Bandit (Multi-arm) API

| Endpoint                     | Method | Purpose                                             | Notes                                                                                     |
|------------------------------|--------|-----------------------------------------------------|-------------------------------------------------------------------------------------------|
| `/v1/bandit/decide`          | POST   | Allocate a user/session into an experiment arm.     | Use before rendering a surface; response includes `arm_id`, `policy_version`.             |
| `/v1/bandit/recommendations` | POST   | Retrieve recommendations with exploration baked in. | Same payload as `/v1/recommendations`; response also includes bandit metadata.            |
| `/v1/bandit/reward`          | POST   | Send reward signals (click/purchase) for an arm.    | Provide `decision_id`, `reward` (0‑1), optional metadata.                                 |
| `/v1/bandit/policies`        | GET    | List policies for an org/namespace.                 | Returns active + historical policies with arm config.                                     |
| `/v1/bandit/policies:upsert` | POST   | Create/update a policy definition.                  | Supply arms, traffic splits, eligibility rules. Used by ops tooling when launching tests. |

**Typical sequence**

1. `POST /v1/bandit/decide` with `namespace`, `surface`, optional traits → response includes `decision_id`, `arm_id`, `policy_version`.
2. Render the surface using `/v1/bandit/recommendations` (bandit-managed blend) or `/v1/recommendations` (pass `arm_id` in `context` to pick a preset).
3. When the user acts, `POST /v1/bandit/reward` with the `decision_id` and `reward` (0/1 or scaled value). Rewards update the policy’s arm statistics.

Policies are defined via `/v1/bandit/policies:upsert`. Each policy lists arms, traffic percentages, and optional eligibility filters (segments, namespaces, surfaces). Use `GET /v1/bandit/policies` to audit what’s live.

## Configuration (Event Types, Segments, Presets)

| Endpoint                           | Method | Purpose                                         | Notes                                                     |
|------------------------------------|--------|-------------------------------------------------|-----------------------------------------------------------|
| `/v1/event-types`                  | GET    | List effective event-type weights & half-lives. | Shows what the ranking engine currently uses.             |
| `/v1/event-types:upsert`           | POST   | Configure event-type weights/half-lives.        | E.g., change purchase weight or deactivate custom events. |
| `/v1/segments`                     | GET    | List behavioral segments.                       | Segments drive cohort-specific tuning.                    |
| `/v1/segments:upsert`              | POST   | Create/update segments.                         | Provide `segment_id`, description, eligibility rules.     |
| `/v1/segments:delete`              | POST   | Remove segments.                                | Use when retiring cohorts.                                |
| `/v1/segment-profiles`             | GET    | Preset starter profiles per segment.            | Useful for cold-start curation.                           |
| `/v1/segment-profiles:upsert`      | POST   | Create/update starter profile weights.          | Map categories/tags to weights.                           |
| `/v1/segment-profiles:delete`      | POST   | Remove starter profiles.                        |                                                           |
| `/v1/segments:dry-run`             | POST   | Test segment definitions without saving.        |                                                           |
| `/v1/admin/recommendation/presets` | GET    | Fetch recommended MMR presets per surface.      | UI tooling can show validated values.                     |
| `/v1/admin/recommendation/config`  | GET/POST | Fetch/apply the active recommendation config.   | Use `analysis/scripts/recommendation_config.py export/apply` to keep git-backed JSON templates per namespace. |

**Usage tips**

- Start with `GET` endpoints to inspect defaults after provisioning a namespace.
- Use the `:dry-run` endpoints (`segments:dry-run`, rules dry-run) before committing changes to avoid breaking guardrails.
- Starter profiles feed cold-start personalization; align them with the data seeded via fixtures/templates.
- Event-type weights/half-lives should mirror the importance of downstream KPIs (e.g., purchases > clicks). Adjust them in tandem with guardrail thresholds.

## Rules & Manual Overrides

| Endpoint                     | Method         | Purpose                                            | Notes                                                                |
|------------------------------|----------------|----------------------------------------------------|----------------------------------------------------------------------|
| `/v1/admin/rules`            | GET/POST       | List/create merchandising rules (boost/block/pin). | POST body defines targets, actions, priority, namespace/surface.     |
| `/v1/admin/rules/{rule_id}`  | GET/PUT/DELETE | Inspect or update a specific rule.                 | Use PUT to adjust windows/priority. DELETE removes the rule.         |
| `/v1/admin/rules/dry-run`    | POST           | Test a rule against synthetic input.               | Returns what would happen without saving.                            |
| `/v1/admin/manual_overrides` | GET/POST       | Manage ad-hoc overrides (short-lived boosts/pins). | Great for campaigns; overrides translate to rules behind the scenes. |
| `/v1/admin/manual_overrides/{override_id}/cancel` | POST | Cancel an active manual override. |

Rules are long-lived merchandising controls. Manual overrides map to temporary rules behind the scenes. Both obey namespace/surface scoping and appear in decision traces (`trace.extras.policy`). Always dry-run complex rules before enabling them in production.

### Rule testing & evidence

1. `POST /v1/admin/rules/dry-run` (or run `analysis/scripts/test_rules.py --base-url <url> --org-id <uuid> --namespace <ns>`) to see how the new rule would modify a seeded namespace. The script writes before/after payloads plus metric samples to `analysis/results/rules_effect_sample.json`.
2. Inspect the dry-run response (`preview.items`, `policy.preview`) and verify the expected rule IDs show up under `trace.extras.policy.rules`.
3. Enable the rule with `POST /v1/admin/rules` and watch `/metrics` for `policy_rule_blocked_items_total{rule_id="<id>"}` and `policy_constraint_leak_total` so you can prove the change behaves safely.

Include the JSON artifacts when filing guardrail evidence or peer reviews.

## Audit & Coverage

| Endpoint                            | Method   | Purpose                                 | Notes                                                                                       |
|-------------------------------------|----------|-----------------------------------------|---------------------------------------------------------------------------------------------|
| `/v1/audit/decisions`               | GET/POST | List audit records or enqueue new ones. | GET supports filters (namespace, time window). POST used internally when tracing decisions. |
| `/v1/audit/decisions/{decision_id}` | GET      | Fetch a specific audit record.          | Contains request, config, response, policy summary.                                         |
| `/v1/audit/search`                  | POST     | Query audits with richer filters.       | Filter by rule IDs, leakage flags, user IDs, time windows, surfaces, etc.                   |

Traces include the full request, resolved algorithm config, policy summaries, and coverage telemetry. Use them to debug guardrail failures or zero-effect overrides.

## Data Governance / Maintenance

| Endpoint   | Method | Purpose                                               | Notes                                                                                         |
|------------|--------|-------------------------------------------------------|-----------------------------------------------------------------------------------------------|
| `/version` | GET    | Emit git commit, build timestamp, and model version.  | Relies on `RECSYS_GIT_COMMIT` / `RECSYS_BUILD_TIME` env vars; falls back to runtime defaults. |
| `/health`  | GET    | Liveness/readiness probe.                             | Returns `{ "status": "ok" }`. Used by Docker/CI.                                              |
| `/docs`    | GET    | Swagger UI / API docs.                                | Serves `swagger.json` / `swagger.yaml`.                                                       |
| `/metrics` | GET    | Prometheus metrics (if enabled via observability env) | Scrape `policy_*`, HTTP latency, DB stats, etc.                                               |

Pair `/version` with determinism artifacts (see `analysis/results/determinism_ci.json` or `make determinism`) when capturing evidence for evaluations or incident reviews.

### Load & chaos workflows

- `make load-test LOAD_BASE_URL=<url> LOAD_ORG_ID=<uuid> LOAD_NAMESPACE=<ns> LOAD_RPS=10,100,1000` – runs the k6 script `analysis/load/recommendations_k6.js`, ramps through each RPS stage, and writes `analysis/results/load_test_summary.json` (latency percentiles, iteration count, error rate).
- `LOAD_USER_POOL`, `LOAD_SURFACE`, and `LOAD_STAGE_DURATION` let you tailor the traffic mix without editing code; set `SUMMARY_PATH` to capture multiple environments (e.g., staging vs. prod) side by side.
- `python analysis/scripts/chaos_toggle.py api pause 15` (or `db stop 20`) temporarily pauses a docker-compose service so you can observe how `/v1/recommendations` behaves during cache/DB outages while the load test is running.
- Share the resulting JSON summaries alongside determinism and scenario evidence when answering evaluation rubrics.

## Using the reference

- Payload schemas: see `api/swagger/swagger.yaml` or the generated docs served at `/docs`.
- Authentication: by default API keys are disabled (`API_AUTH_ENABLED=false`). If enabled, set `X-API-Key` header.
- Namespacing: every write/read call requires `namespace` (explicit field or inferred from user/item). Guardrails and env profiles apply per namespace.
- Rate limiting: configure via `API_RATE_LIMIT_RPM`/`BURST`; admins should mention limits in partner docs once keys are issued.
For deeper walkthroughs and examples, see:
- `README.md` (operational checklist, overrides, bandit flows)
- `docs/bespoke_simulations.md` (seeding + simulation) 
- `docs/rules-runbook.md` (override troubleshooting)
