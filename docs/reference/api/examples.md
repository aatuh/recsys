# API examples (HTTP files)

These examples are written in `.http` format for tools like:

- JetBrains HTTP Client (built into IntelliJ/GoLand)
- VS Code REST Client (or similar)

They are also useful as **copy/paste reference** for curl clients.

## Recommend

```http
POST https://example.com/v1/recommend
Authorization: Bearer {{token}}
Content-Type: application/json

{ "surface": "home", "k": 20, "user": { "user_id": "u_1", "anonymous_id": null, "session_id": "s_1" } }
```

Notes:

- `surface` selects the surface namespace (and typically maps to a UI placement like `home`, `pdp`, etc).
- `k` is the number of items requested.
- `user` can carry one or more identifiers. Prefer a stable `user_id` when available.
- Tenant scope + `surface` determine which config/rules/signals are used.
- Use `POST /v1/recommend/validate` during development to normalize requests and surface warnings early.

## Similar items

```http
POST https://example.com/v1/similar
Authorization: Bearer {{token}}
Content-Type: application/json

{ "surface": "pdp", "item_id": "item_1", "k": 20 }
```

Notes:

- `item_id` is the anchor item you want neighbors for.
- Similarity requires similarity/co-vis signals. If unavailable, the API may return warnings or empty results depending
  on the configured algorithm and available signals.

## Admin: config, rules, cache invalidation

These endpoints are for **operators** (bootstrap + config management).

See also: [`admin.md`](admin.md).

```http
### Get config (admin)
GET https://example.com/v1/admin/tenants/demo/config
Authorization: Bearer {{admin_token}}

### Update config (admin)
PUT https://example.com/v1/admin/tenants/demo/config
Authorization: Bearer {{admin_token}}
Content-Type: application/json
If-Match: {{etag}}

{ "weights": { "pop": 0.5, "cooc": 0.2, "emb": 0.3 } }

### Get rules (admin)
GET https://example.com/v1/admin/tenants/demo/rules
Authorization: Bearer {{admin_token}}

### Update rules (admin)
PUT https://example.com/v1/admin/tenants/demo/rules
Authorization: Bearer {{admin_token}}
Content-Type: application/json
If-Match: {{etag}}

[
  {
    "action": "pin",
    "target_type": "item",
    "item_ids": ["item_1", "item_2"],
    "surface": "home",
    "priority": 10
  },
  {
    "action": "block",
    "target_type": "tag",
    "target_key": "brand:nike"
  }
]

### Invalidate caches (admin)
POST https://example.com/v1/admin/tenants/demo/cache/invalidate
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{ "targets": ["config", "rules"], "surface": "home" }
```

Notes:

- Admin endpoints require elevated scope (viewer/operator/admin roles in JWT mode).
- `If-Match` enables optimistic concurrency using the current `ETag`/version value.
- Cache invalidation is safe to run after updates to reduce “stale config” confusion during incidents.
