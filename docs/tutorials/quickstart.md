---
tags:
  - tutorial
  - quickstart
  - developer
  - recsys-service
---

# Tutorial: Quickstart (10 minutes)

## Who this is for

- Developers who want the fastest path to a non-empty `POST /v1/recommend` response.

## What you will get

- `recsys-service` running locally in DB-only mode
- one successful `POST /v1/recommend`
- one saved exposure log file you can later evaluate

> **Choose your data mode**
>
> This tutorial uses **DB-only mode** (fastest path to first success).
>
> - Choose **DB-only** to validate API integration, tenancy, and exposure logging with the smallest moving parts.
> - Choose **artifact/manifest mode** when you want atomic publish and rollback (pipelines produce artifacts and a
>   manifest pointer drives serving).
>
> See: [`explanation/data-modes.md`](../explanation/data-modes.md). For an artifact-mode walkthrough, jump to
> [`tutorials/production-like-run.md`](production-like-run.md).

## Prereqs

- Docker + Docker Compose (v2)
- `make`
- `curl`
- POSIX shell
- Optional: `jq` (prettier output)

Verify you have them:

```bash
docker compose version
make --version
curl --version
```

!!! info "Key terms (2 minutes)"
    - **[Tenant](../project/glossary.md#tenant)**: a configuration + data isolation boundary (usually one organization).
    - **[Surface](../project/glossary.md#surface)**: where recommendations are shown (home, PDP, cart, ...).
    - **[Request ID](../project/glossary.md#request-id)**: the join key that ties together responses, exposures, and outcomes.
    - **[Exposure log](../project/glossary.md#exposure-log)**: what was shown (audit trail + evaluation input).

## 1) Start Postgres + `recsys-service` (DB-only mode)

From repo root, create a clean tutorial environment file:

```bash
if [ -f api/.env ]; then
  cp api/.env "/tmp/recsys-api.env.$(date +%s).bak"
fi

cp api/.env.example api/.env
```

<details markdown="1" open>
<summary>Quickstart tutorial defaults (required)</summary>

Append these values to `api/.env`:

```bash
cat >> api/.env <<'ENV'

# Quickstart overrides (DB-only mode)
RECSYS_ARTIFACT_MODE_ENABLED=false
RECSYS_ALGO_MODE=popularity
RECSYS_ALGO_RULES_ENABLED=true

# Enable eval-compatible exposure logs
EXPOSURE_LOG_ENABLED=true
EXPOSURE_LOG_FORMAT=eval_v1
EXPOSURE_LOG_PATH=/app/tmp/exposures.eval.jsonl

# Local/dev: disable admin RBAC roles (dev headers don’t carry roles)
AUTH_VIEWER_ROLE=
AUTH_OPERATOR_ROLE=
AUTH_ADMIN_ROLE=
ENV
```

</details>

Start only the DB + API:

```bash
docker compose up -d db api
```

Wait until the service is healthy:

```bash
for _ in $(seq 1 60); do
  if curl -fsS http://localhost:8000/healthz >/dev/null; then
    break
  fi
  sleep 2
done
curl -fsS http://localhost:8000/healthz >/dev/null
```

Apply database migrations (idempotent):

```bash
(cd api && make migrate-up)
```

Expected:

- The final `curl -fsS http://localhost:8000/healthz` exits 0.
- `(cd api && make migrate-up)` exits 0.

## 2) Bootstrap a demo tenant + minimal data

Insert a tenant row:

```bash
docker exec -i recsys-db psql -U recsys-db -d recsys-db <<'SQL'
insert into tenants (external_id, name)
values ('demo', 'Demo Tenant')
on conflict (external_id) do nothing;
SQL
```

Upsert a minimal config:

```bash
curl -fsS -X PUT http://localhost:8000/v1/admin/tenants/demo/config \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d @- <<'JSON'
{
  "weights": { "pop": 1.0, "cooc": 0.0, "emb": 0.0 },
  "flags": { "enable_rules": true },
  "limits": { "max_k": 50, "max_exclude_ids": 200 }
}
JSON
```

Upsert rules (pin `item_3` to prove control works):

```bash
curl -fsS -X PUT http://localhost:8000/v1/admin/tenants/demo/rules \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d @- <<'JSON'
[
  {
    "action": "pin",
    "target_type": "item",
    "item_ids": ["item_3"],
    "surface": "home",
    "priority": 10
  }
]
JSON
```

Seed `item_tags` and `item_popularity_daily` for surface `home`:

```bash
docker exec -i recsys-db psql -U recsys-db -d recsys-db <<'SQL'
with t as (
  select id as tenant_id
    from tenants
   where external_id = 'demo'
)
insert into item_tags (tenant_id, namespace, item_id, tags, price, created_at)
select tenant_id, 'home', 'item_1', array['brand:nike','category:shoes'], 99.90, now() from t
union all
select tenant_id, 'home', 'item_2', array['brand:nike','category:shoes'], 79.00, now() from t
union all
select tenant_id, 'home', 'item_3', array['brand:acme','category:socks'], 12.00, now() from t
on conflict (tenant_id, namespace, item_id)
do update set tags = excluded.tags,
              price = excluded.price,
              created_at = excluded.created_at;

with t as (
  select id as tenant_id
    from tenants
   where external_id = 'demo'
)
insert into item_popularity_daily (tenant_id, namespace, item_id, day, score)
select tenant_id, 'home', 'item_1', current_date, 10 from t
union all
select tenant_id, 'home', 'item_2', current_date, 7 from t
union all
select tenant_id, 'home', 'item_3', current_date, 3 from t
on conflict (tenant_id, namespace, item_id, day)
do update set score = excluded.score;
SQL
```

Expected:

- Each command exits 0 (`docker exec ... psql` and both `curl -fsS -X PUT ...` calls).

## 3) Call `POST /v1/recommend`

Send a request with deterministic `request_id`:

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: req-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1","session_id":"s_1"}}'
```

Expected:

- Response has a non-empty `items` list containing `item_1`, `item_2`, `item_3`.
- Because you pinned `item_3`, it appears first.

## 4) Save the exposure log (audit trail)

Confirm the service wrote an exposure log, and copy it locally:

```bash
docker compose exec -T api sh -c 'test -s /app/tmp/exposures.eval.jsonl'
docker compose cp api:/app/tmp/exposures.eval.jsonl /tmp/exposures.eval.jsonl
head -n 1 /tmp/exposures.eval.jsonl
```

Expected:

- The log file exists and is non-empty.
- The first line is JSON and contains `request_id`.

## Verify (Definition of Done)

- [ ] `curl -fsS http://localhost:8000/healthz` succeeds
- [ ] `POST /v1/recommend` returns a non-empty `items` list
- [ ] `/tmp/exposures.eval.jsonl` exists and contains a `request_id`

## Troubleshooting (common failures)

- `/healthz` never becomes healthy → [Service not ready](../operations/runbooks/service-not-ready.md)
- `POST /v1/recommend` returns empty `items` → [Empty recs](../operations/runbooks/empty-recs.md)
- Migrations fail or tables are missing → [Database migration issues](../operations/runbooks/db-migration-issues.md)
- Exposure log file is missing → confirm `EXPOSURE_LOG_*` in `api/.env`, then restart:
  `docker compose up -d --force-recreate api`

## Read next

- Integrate into an app: [`how-to/integrate-recsys-service.md`](../how-to/integrate-recsys-service.md)
- Full walkthrough (serving → logging → eval): [`tutorials/local-end-to-end.md`](local-end-to-end.md)
- API reference (Swagger UI + OpenAPI spec): [`reference/api/api-reference.md`](../reference/api/api-reference.md)
