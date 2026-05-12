#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROOF_DIR="${PROOF_DIR:-${ROOT_DIR}/tmp/commercial-proof-kit}"
DATA_DIR="${PROOF_KIT_DATA_DIR:-${ROOT_DIR}/examples/data/ecommerce-mini}"
PIPELINES_CFG="${PIPELINES_CFG:-${ROOT_DIR}/examples/demo/recsys-pipelines.ecommerce-mini.minio.json}"
EVAL_DATASET="${EVAL_DATASET:-${ROOT_DIR}/recsys-eval/configs/examples/dataset.ecommerce-mini.jsonl.yaml}"
EVAL_CONFIG="${EVAL_CONFIG:-${ROOT_DIR}/recsys-eval/configs/eval/offline.ecommerce-mini.yaml}"
BASE_URL="${BASE_URL:-http://localhost:8000}"
TENANT_ID="${TENANT_ID:-demo}"
SURFACE="${SURFACE:-home}"
START_DAY="${START_DAY:-2026-01-01}"
END_DAY="${END_DAY:-2026-01-01}"
MINIO_BUCKET="${MINIO_BUCKET:-recsys-artifacts}"
MINIO_USER="${MINIO_ROOT_USER:-minioadmin}"
MINIO_PASS="${MINIO_ROOT_PASSWORD:-minioadmin}"
KEEP_STACK="${KEEP_STACK:-0}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

check_required_path() {
  local path="$1"
  if [ ! -e "${path}" ]; then
    echo "commercial demo missing required path: ${path}" >&2
    exit 1
  fi
}

validate_token() {
  local name="$1"
  local value="$2"
  case "${value}" in
    *[!A-Za-z0-9._-]*|"")
      echo "${name} must contain only letters, numbers, dot, underscore, or dash" >&2
      exit 1
      ;;
  esac
}

cleanup() {
  if [ "${KEEP_STACK}" != "1" ]; then
    RECSYS_API_ENV_FILE="${PROOF_DIR}/api.env" docker compose -f "${ROOT_DIR}/docker-compose.yml" down --remove-orphans >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

require_cmd curl
require_cmd docker
require_cmd go
require_cmd make
require_cmd python3

check_required_path "${DATA_DIR}/pipelines/exposure.jsonl"
check_required_path "${DATA_DIR}/eval/exposures.jsonl"
check_required_path "${DATA_DIR}/eval/outcomes.jsonl"
check_required_path "${PIPELINES_CFG}"
check_required_path "${EVAL_DATASET}"
check_required_path "${EVAL_CONFIG}"
validate_token TENANT_ID "${TENANT_ID}"
validate_token SURFACE "${SURFACE}"
validate_token MINIO_BUCKET "${MINIO_BUCKET}"

rm -rf "${PROOF_DIR}"
mkdir -p "${PROOF_DIR}/eval"

cp "${ROOT_DIR}/api/.env.example" "${PROOF_DIR}/api.env"
cat >> "${PROOF_DIR}/api.env" <<'ENV'

# Commercial proof-kit demo overrides.
RECSYS_ARTIFACT_MODE_ENABLED=true
RECSYS_ARTIFACT_MANIFEST_TEMPLATE=s3://recsys-artifacts/registry/current/{tenant}/{surface}/manifest.json
RECSYS_ARTIFACT_MANIFEST_TTL=1s
RECSYS_ARTIFACT_CACHE_TTL=1s
RECSYS_ALGO_MODE=popularity
RECSYS_ALGO_RULES_ENABLED=false

EXPOSURE_LOG_ENABLED=true
EXPOSURE_LOG_FORMAT=eval_v1
EXPOSURE_LOG_PATH=/app/tmp/commercial-proof-kit.exposures.jsonl

AUTH_VIEWER_ROLE=
AUTH_OPERATOR_ROLE=
AUTH_ADMIN_ROLE=
ENV

export RECSYS_API_ENV_FILE="${PROOF_DIR}/api.env"

echo "Starting local proof-kit stack..."
docker compose -f "${ROOT_DIR}/docker-compose.yml" up -d db minio minio-init api

echo "Waiting for API health..."
for _ in $(seq 1 60); do
  if curl -fsS "${BASE_URL}/healthz" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
curl -fsS "${BASE_URL}/healthz" >/dev/null

echo "Applying migrations..."
docker compose -f "${ROOT_DIR}/docker-compose.yml" exec -T api sh -c 'dir_arg=""; \
  if [ -n "$MIGRATIONS_DIR" ] && [ "$MIGRATIONS_DIR" != "-" ]; then \
    dir_arg="-dir $MIGRATIONS_DIR"; \
  fi; \
  GOTMPDIR=/app/tmp TMPDIR=/app/tmp go run ./cmd/migrate $dir_arg up' >/dev/null

echo "Seeding demo tenant..."
docker exec -i recsys-db psql -U recsys-db -d recsys-db <<SQL >/dev/null
insert into tenants (external_id, name)
values ('${TENANT_ID}', 'Commercial Proof Kit Demo')
on conflict (external_id) do nothing;
SQL

cat > "${PROOF_DIR}/tenant-config.json" <<'JSON'
{
  "weights": { "pop": 1.0, "cooc": 0.0, "emb": 0.0 },
  "flags": { "enable_rules": false },
  "limits": { "max_k": 50, "max_exclude_ids": 200 }
}
JSON

curl -fsS -X PUT "${BASE_URL}/v1/admin/tenants/${TENANT_ID}/config" \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Id: proof-kit-admin' \
  -H "X-Dev-Org-Id: ${TENANT_ID}" \
  -H "X-Org-Id: ${TENANT_ID}" \
  -d @"${PROOF_DIR}/tenant-config.json" >/dev/null

echo "Running pipelines and publishing artifacts to MinIO..."
(cd "${ROOT_DIR}/recsys-pipelines" && GOWORK=off go run ./cmd/recsys-pipelines \
  run --config "${PIPELINES_CFG}" --tenant "${TENANT_ID}" --surface "${SURFACE}" \
  --start "${START_DAY}" --end "${END_DAY}")

MANIFEST_PATH="${PROOF_DIR}/pipelines/registry/current/${TENANT_ID}/${SURFACE}/manifest.json"
if [ ! -s "${MANIFEST_PATH}" ]; then
  echo "manifest not found: ${MANIFEST_PATH}" >&2
  exit 1
fi
cp "${MANIFEST_PATH}" "${PROOF_DIR}/manifest.json"

docker compose -f "${ROOT_DIR}/docker-compose.yml" run --rm --entrypoint sh \
  -e MC_USER="${MINIO_USER}" \
  -e MC_PASS="${MINIO_PASS}" \
  -e MC_BUCKET="${MINIO_BUCKET}" \
  -e MC_TENANT="${TENANT_ID}" \
  -e MC_SURFACE="${SURFACE}" \
  -v "${MANIFEST_PATH}:/tmp/manifest.json:ro" \
  minio-init -c \
  'set -eu; \
   mc alias set local http://minio:9000 "$MC_USER" "$MC_PASS" >/dev/null; \
   mc mb -p "local/$MC_BUCKET" >/dev/null 2>&1 || true; \
   mc cp /tmp/manifest.json "local/$MC_BUCKET/registry/current/$MC_TENANT/$MC_SURFACE/manifest.json" >/dev/null'

sleep 2

echo "Calling /v1/recommend..."
curl -fsS "${BASE_URL}/v1/recommend" \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: proof-kit-serve-1' \
  -H 'X-Dev-User-Id: proof-kit-user' \
  -H "X-Dev-Org-Id: ${TENANT_ID}" \
  -H "X-Org-Id: ${TENANT_ID}" \
  -d "{\"surface\":\"${SURFACE}\",\"k\":5,\"user\":{\"user_id\":\"u_footwear\",\"session_id\":\"s_ecom_live\"},\"options\":{\"include_reasons\":true}}" \
  > "${PROOF_DIR}/recommendation-response.json"

python3 - "${PROOF_DIR}/recommendation-response.json" <<'PY'
import json
import sys

with open(sys.argv[1], encoding="utf-8") as f:
    payload = json.load(f)

items = payload.get("items") or []
if not items:
    raise SystemExit("expected non-empty recommendation items")
PY

docker exec recsys-svc sh -c 'test -s /app/tmp/commercial-proof-kit.exposures.jsonl'
docker cp recsys-svc:/app/tmp/commercial-proof-kit.exposures.jsonl "${PROOF_DIR}/served-exposures.eval.jsonl" >/dev/null

echo "Building recsys-eval and writing evaluation reports..."
(cd "${ROOT_DIR}/recsys-eval" && make build >/dev/null)
(cd "${ROOT_DIR}/recsys-eval" && ./bin/recsys-eval validate --schema exposure.v1 --input "${DATA_DIR}/eval/exposures.jsonl")
(cd "${ROOT_DIR}/recsys-eval" && ./bin/recsys-eval validate --schema outcome.v1 --input "${DATA_DIR}/eval/outcomes.jsonl")
(cd "${ROOT_DIR}/recsys-eval" && ./bin/recsys-eval run \
  --mode offline \
  --dataset "${EVAL_DATASET}" \
  --config "${EVAL_CONFIG}" \
  --output "${PROOF_DIR}/eval/offline-report.json" \
  --output-format json)
(cd "${ROOT_DIR}/recsys-eval" && ./bin/recsys-eval run \
  --mode offline \
  --dataset "${EVAL_DATASET}" \
  --config "${EVAL_CONFIG}" \
  --output "${PROOF_DIR}/eval/offline-report.md" \
  --output-format markdown)

cat > "${PROOF_DIR}/decision-note.md" <<'MD'
# Commercial Proof Kit Decision Note

## Context

- Tenant: `demo`
- Surface: `home`
- Dataset: `examples/data/ecommerce-mini`
- Data handling: synthetic, non-PII fixture

## Proof Artifacts

- Served recommendation response: `recommendation-response.json`
- Served exposure log: `served-exposures.eval.jsonl`
- Published manifest: `manifest.json`
- Offline evaluation report: `eval/offline-report.json` and `eval/offline-report.md`

## Decision

- Decision: ship / hold / rollback
- Reasoning:
  - The local serving API returned non-empty recommendations.
  - Exposure/outcome fixtures validated against the evaluation schemas.
  - The offline report was generated from joinable synthetic data.
- Follow-ups:
  - Replace the synthetic fixture with one real pilot surface.
  - Confirm production auth, logging retention, and rollback ownership.
MD

echo "Commercial proof kit complete."
echo "Outputs:"
find "${PROOF_DIR}" -maxdepth 3 -type f | sort
