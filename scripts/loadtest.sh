#!/usr/bin/env bash
# Runs the Go-based load test client against recsys-service.
# - Uses env vars for target URL, tenant/surface, and auth headers
# - Exercises /v1/recommend (or another endpoint) with configurable concurrency
# - Sends a mix of user_ids for realistic cache and personalization behavior
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

BASE_URL="${BASE_URL:-http://localhost:8000}"
ENDPOINT="${ENDPOINT:-/v1/recommend}"
TENANT_ID="${TENANT_ID:-demo}"
SURFACE="${SURFACE:-home}"
SEGMENT="${SEGMENT:-}"
ITEM_ID="${ITEM_ID:-item_1}"
K="${K:-20}"
REQUESTS="${REQUESTS:-200}"
CONCURRENCY="${CONCURRENCY:-10}"
USER_PREFIX="${USER_PREFIX:-user}"
USER_CARDINALITY="${USER_CARDINALITY:-1000}"

DEV_HEADERS="${DEV_HEADERS:-true}"
DEV_TENANT_HEADER="${DEV_TENANT_HEADER:-X-Dev-Org-Id}"
DEV_USER_HEADER="${DEV_USER_HEADER:-X-Dev-User-Id}"
TENANT_HEADER="${TENANT_HEADER:-X-Org-Id}"

BEARER_TOKEN="${BEARER_TOKEN:-}"
API_KEY="${API_KEY:-}"
API_KEY_HEADER="${API_KEY_HEADER:-X-API-Key}"

echo "Running load test against ${BASE_URL}${ENDPOINT}"

go run "${ROOT_DIR}/api/cmd/loadtest" \
  -url "${BASE_URL}" \
  -endpoint "${ENDPOINT}" \
  -surface "${SURFACE}" \
  -segment "${SEGMENT}" \
  -item-id "${ITEM_ID}" \
  -k "${K}" \
  -tenant "${TENANT_ID}" \
  -tenant-header "${TENANT_HEADER}" \
  -dev="${DEV_HEADERS}" \
  -dev-tenant-header "${DEV_TENANT_HEADER}" \
  -dev-user-header "${DEV_USER_HEADER}" \
  -bearer "${BEARER_TOKEN}" \
  -api-key "${API_KEY}" \
  -api-key-header "${API_KEY_HEADER}" \
  -user-prefix "${USER_PREFIX}" \
  -user-cardinality "${USER_CARDINALITY}" \
  -n "${REQUESTS}" \
  -c "${CONCURRENCY}"
