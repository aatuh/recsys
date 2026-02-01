#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEMO_DIR="$ROOT/examples/demo"
DATA_DIR="$ROOT/examples/data/tiny"
PIPELINES_CFG="$DEMO_DIR/recsys-pipelines.minio.json"
mkdir -p "$ROOT/tmp"

BASE_URL="${BASE_URL:-http://localhost:8000}"
TENANT_ID="${TENANT_ID:-demo}"
SURFACE="${SURFACE:-home}"
START_DAY="${START_DAY:-2026-01-01}"
END_DAY="${END_DAY:-2026-01-01}"

echo "▶ Starting demo stack..."
cd "$ROOT"
make dev >/dev/null

echo "▶ Waiting for API health..."
for i in {1..30}; do
  if curl -fsS "$BASE_URL/healthz" >/dev/null; then
    break
  fi
  sleep 2
done
curl -fsS "$BASE_URL/healthz" >/dev/null

echo "▶ Seeding tenant..."
POSTGRES_USER="${POSTGRES_USER:-recsys-db}"
POSTGRES_DB="${POSTGRES_DB:-recsys-db}"
docker exec recsys-db psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
  "insert into tenants (external_id, name) values ('$TENANT_ID', 'Demo Tenant') on conflict (external_id) do nothing;" >/dev/null

echo "▶ Running pipelines..."
cd "$ROOT/recsys-pipelines"
go run ./cmd/recsys-pipelines \
  run --config "$PIPELINES_CFG" --tenant "$TENANT_ID" --surface "$SURFACE" \
  --start "$START_DAY" --end "$END_DAY"

echo "▶ Calling /v1/recommend..."
payload=$(cat <<JSON
{
  "surface": "$SURFACE",
  "segment": "default",
  "k": 5,
  "user": { "user_id": "demo-user", "session_id": "demo-session" },
  "options": { "include_reasons": true }
}
JSON
)
curl -fsS "$BASE_URL/v1/recommend" \
  -H "Content-Type: application/json" \
  -H "X-Dev-User-Id: demo-user" \
  -H "X-Dev-Org-Id: $TENANT_ID" \
  -H "X-Org-Id: $TENANT_ID" \
  -d "$payload" | tee "$ROOT/tmp/demo-recommend.json"

echo
echo "✅ Demo complete. Output saved to tmp/demo-recommend.json"
