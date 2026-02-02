# DB-only signals: schema + seed examples

Use these tables when running recsys-service in DB-only mode (no artifacts).

## 1) Resolve tenant UUID

`external_id` should match your tenant claim or dev header value.

```sql
select id from tenants where external_id = 'demo';
```

Assume the result is `:tenant_id`.

## 2) item_tags

Columns:

- tenant_id (uuid)
- namespace (surface, text)
- item_id (text)
- tags (text[])
- price (numeric, optional)
- created_at (timestamptz)

```sql
insert into item_tags (tenant_id, namespace, item_id, tags, price, created_at)
values
  (:tenant_id, 'home', 'item-1', array['brand:nike','category:shoes'], 99.90, now()),
  (:tenant_id, 'home', 'item-2', array['brand:nike','category:shoes'], 79.00, now())
on conflict (tenant_id, namespace, item_id)
do update set tags = excluded.tags,
              price = excluded.price,
              created_at = excluded.created_at;
```

## 3) item_popularity_daily

Columns:

- tenant_id (uuid)
- namespace (surface, text)
- item_id (text)
- day (date)
- score (numeric)

```sql
insert into item_popularity_daily (tenant_id, namespace, item_id, day, score)
values
  (:tenant_id, 'home', 'item-1', '2026-01-30', 10),
  (:tenant_id, 'home', 'item-2', '2026-01-30', 8)
on conflict (tenant_id, namespace, item_id, day)
do update set score = excluded.score;
```

Note: popularity is a **decayed sum across days**, so newer days dominate when
multiple days are present.

## 4) item_covisit_daily (for /v1/similar)

Columns:

- tenant_id (uuid)
- namespace (surface, text)
- item_id (anchor)
- neighbor_id
- day (date)
- score (numeric)

```sql
insert into item_covisit_daily (tenant_id, namespace, item_id, neighbor_id, day, score)
values
  (:tenant_id, 'home', 'item-1', 'item-2', '2026-01-30', 3),
  (:tenant_id, 'home', 'item-1', 'item-3', '2026-01-30', 1)
on conflict (tenant_id, namespace, item_id, neighbor_id, day)
do update set score = excluded.score;
```

If `/v1/similar` returns empty:

- ensure co-vis rows exist **for the same surface** (namespace)
- ensure the anchor item exists in `item_covisit_daily`
