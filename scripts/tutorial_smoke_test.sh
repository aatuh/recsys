#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

ORIG_ENV_FILE=""

cleanup() {
  docker compose -f docker-compose.yml down --remove-orphans >/dev/null 2>&1 || true
  if [ -n "${ORIG_ENV_FILE}" ] && [ -f "${ORIG_ENV_FILE}" ]; then
    cp "${ORIG_ENV_FILE}" api/.env
    rm -f "${ORIG_ENV_FILE}"
  else
    rm -f api/.env
  fi
  rm -f /tmp/demo_config.json /tmp/demo_rules.json /tmp/recommend.json /tmp/exposures.jsonl
}
trap cleanup EXIT

echo "Preparing api/.env for the tutorial smoke test..."
if [ -f api/.env ]; then
  ORIG_ENV_FILE="$(mktemp)"
  cp api/.env "${ORIG_ENV_FILE}"
fi
cp api/.env.example api/.env
cat >> api/.env <<'ENV'

# Tutorial smoke overrides (DB-only mode)
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

echo "Starting db + recsys-service..."
docker compose -f docker-compose.yml up -d db api

echo "Waiting for service health..."
for _ in $(seq 1 60); do
  if curl -fsS http://localhost:8000/healthz >/dev/null; then
    break
  fi
  sleep 2
done
curl -fsS http://localhost:8000/healthz >/dev/null

echo "Applying migrations (idempotent)..."
(cd api && make migrate-up)

echo "Bootstrapping demo tenant..."
docker exec -i recsys-db psql -U recsys-db -d recsys-db <<'SQL'
insert into tenants (external_id, name)
values ('demo', 'Demo Tenant')
on conflict (external_id) do nothing;
SQL

echo "Upserting tenant config..."
cat > /tmp/demo_config.json <<'JSON'
{
  "weights": { "pop": 1.0, "cooc": 0.0, "emb": 0.0 },
  "flags": { "enable_rules": true },
  "limits": { "max_k": 50, "max_exclude_ids": 200 }
}
JSON

curl -fsS -X PUT http://localhost:8000/v1/admin/tenants/demo/config \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d @/tmp/demo_config.json >/dev/null

echo "Upserting tenant rules (pin item_3)..."
cat > /tmp/demo_rules.json <<'JSON'
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

curl -fsS -X PUT http://localhost:8000/v1/admin/tenants/demo/rules \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d @/tmp/demo_rules.json >/dev/null

echo "Seeding minimal DB-only signals..."
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

echo "Calling /v1/recommend..."
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: req-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1","session_id":"s_1"}}' \
  > /tmp/recommend.json

python3 - <<'PY'
import json

with open("/tmp/recommend.json", encoding="utf-8") as f:
    data = json.load(f)

items = data.get("items") or []
assert items, "expected non-empty 'items'"
assert items[0].get("item_id") == "item_3", f"expected item_3 pinned first, got {items[0]!r}"
PY

echo "Checking that an exposure log entry exists..."
docker exec recsys-svc sh -c 'test -s /app/tmp/exposures.eval.jsonl'
docker cp recsys-svc:/app/tmp/exposures.eval.jsonl /tmp/exposures.jsonl

echo "Validating exposure schema with recsys-eval..."
(cd recsys-eval && make build)
(cd recsys-eval && ./bin/recsys-eval validate --schema exposure.v1 --input /tmp/exposures.jsonl)

echo "Tutorial smoke test OK"
