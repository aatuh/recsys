#!/usr/bin/env bash
# Demo script: boots the local stack via docker compose, runs recsys-pipelines
# on demo data, publishes the manifest to MinIO, seeds a tenant, and calls
# /v1/recommend with dev headers. Output is saved under tmp/.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEMO_DIR="$ROOT/examples/demo"
PIPELINES_CFG="${PIPELINES_CFG:-$DEMO_DIR/recsys-pipelines.minio.json}"
PIPELINES_BIN="${PIPELINES_BIN:-}"
MINIO_BUCKET="${MINIO_BUCKET:-recsys-artifacts}"
MINIO_USER="${MINIO_ROOT_USER:-minioadmin}"
MINIO_PASS="${MINIO_ROOT_PASSWORD:-minioadmin}"
mkdir -p "$ROOT/tmp"

BASE_URL="${BASE_URL:-http://localhost:8000}"
TENANT_ID="${TENANT_ID:-demo}"
SURFACE="${SURFACE:-home}"
START_DAY="${START_DAY:-2026-01-01}"
END_DAY="${END_DAY:-2026-01-01}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "✖ Missing required command: $1" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd docker
if [ -z "$PIPELINES_BIN" ]; then
  require_cmd go
fi

if [ ! -f "$PIPELINES_CFG" ]; then
  echo "✖ Pipelines config not found: $PIPELINES_CFG" >&2
  exit 1
fi

echo "▶ Starting demo stack..."
cd "$ROOT"
make dev

echo "▶ Waiting for services to start..."
sleep 15

echo "▶ Waiting for API health..."
for _ in {1..30}; do
  if curl -fsS "$BASE_URL/healthz" >/dev/null; then
    break
  fi
  sleep 4
done
curl -fsS "$BASE_URL/healthz" >/dev/null

echo "▶ Waiting for MinIO..."
for _ in {1..30}; do
  if curl -fsS "http://localhost:9000/minio/health/ready" >/dev/null; then
    break
  fi
  sleep 2
done
curl -fsS "http://localhost:9000/minio/health/ready" >/dev/null

echo "▶ Seeding tenant..."
POSTGRES_USER="${POSTGRES_USER:-recsys-db}"
POSTGRES_DB="${POSTGRES_DB:-recsys-db}"
docker exec recsys-db psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
  "insert into tenants (external_id, name) values ('$TENANT_ID', 'Demo Tenant') on conflict (external_id) do nothing;" >/dev/null

echo "▶ Running pipelines..."
cd "$ROOT/recsys-pipelines"
if [ -n "$PIPELINES_BIN" ]; then
  "$PIPELINES_BIN" run \
    --config "$PIPELINES_CFG" --tenant "$TENANT_ID" --surface "$SURFACE" \
    --start "$START_DAY" --end "$END_DAY"
else
  GOWORK=off go run ./cmd/recsys-pipelines \
    run --config "$PIPELINES_CFG" --tenant "$TENANT_ID" --surface "$SURFACE" \
    --start "$START_DAY" --end "$END_DAY"
fi

echo "▶ Publishing manifest to MinIO..."
MANIFEST_PATH="$ROOT/tmp/demo-pipelines/registry/current/$TENANT_ID/$SURFACE/manifest.json"
if [ ! -f "$MANIFEST_PATH" ]; then
  echo "✖ Manifest not found: $MANIFEST_PATH" >&2
  exit 1
fi
docker run --rm --network recsys_default \
  --entrypoint /bin/sh \
  -v "$ROOT/tmp/demo-pipelines/registry:/registry:ro" \
  minio/mc:RELEASE.2024-10-02T08-27-28Z \
  -c "mc alias set local http://minio:9000 $MINIO_USER $MINIO_PASS >/dev/null && \
    mc mb -p local/$MINIO_BUCKET >/dev/null 2>&1 || true && \
    mc cp /registry/current/$TENANT_ID/$SURFACE/manifest.json local/$MINIO_BUCKET/registry/current/$TENANT_ID/$SURFACE/manifest.json"

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
