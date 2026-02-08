---
diataxis: tutorial
tags:
  - tutorial
  - quickstart
  - developer
  - recsys-service
---
# Tutorial: Quickstart (minimal)

This tutorial gets you a **non-empty** `POST /v1/recommend` response in the fewest steps.

## Who this is for

- Developers who want the fastest possible proof that the service runs and returns ranked items.

## What you will get

- `recsys-service` running locally
- one successful recommendation response with **non-empty** `items[]`
- a pointer to the integration contract and the required logging

## Prereqs

- Docker + Docker Compose v2
- `curl`

## 1) Start the stack

From the repository root:

```bash
docker compose up -d db api
```

Verify containers are running:

```bash
docker compose ps
```

Wait for the API to become healthy:

```bash
for _ in $(seq 1 60); do
  if curl -fsS http://localhost:8000/healthz >/dev/null; then
    break
  fi
  sleep 2
done
curl -fsS http://localhost:8000/healthz >/dev/null
```

## 2) Seed demo data

Apply migrations (idempotent), then seed tenant + minimal DB-only signals:

```bash
(cd api && make migrate-up)

docker exec -i recsys-db psql -U recsys-db -d recsys-db <<'SQL'
insert into tenants (external_id, name)
values ('demo', 'Demo Tenant')
on conflict (external_id) do nothing;

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

## 3) Request recommendations

```bash
curl -fsS -X POST 'http://localhost:8000/v1/recommend' \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1"}}'
```

**Success looks like:**

- HTTP 200
- `items[]` is non-empty
- each item has `item_id`, `rank`, and `score`

You may also see `warnings[]`; in this tutorial that is expected:

- `SIGNAL_UNAVAILABLE` for collaborative/content/session means those optional signals were not seeded in this minimal DB-only setup.
- `DEFAULT_APPLIED` means omitted request fields (for example `segment` and `options`) were filled with defaults.
- These warnings are non-fatal for this tutorial. The result is valid as long as you get HTTP 200 and a non-empty `items[]`.

## What you just proved

- The service runs locally.
- Tenancy scoping works (`X-Org-Id`).
- Ranking produces an ordered list.

## Read next

- Full local loop (serve → log → eval): [Tutorial: Local end-to-end (20–30 minutes)](local-end-to-end.md)
- Minimal integration contract: [Integration spec](../reference/integration-spec.md)
- Evaluation-ready logging requirements: [Minimum instrumentation spec](../reference/minimum-instrumentation.md)
- Architecture and data flow: [How it works](../explanation/how-it-works.md)
