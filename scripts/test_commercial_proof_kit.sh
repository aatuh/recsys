#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
PROOF_DIR="${PROOF_DIR:-${ROOT_DIR}/tmp/commercial-proof-kit}"
DATA_DIR="${PROOF_KIT_DATA_DIR:-${ROOT_DIR}/examples/data/ecommerce-mini}"
PIPELINES_CFG="${PIPELINES_CFG:-${ROOT_DIR}/examples/demo/recsys-pipelines.ecommerce-mini.fs.json}"
EVAL_DATASET="${EVAL_DATASET:-${ROOT_DIR}/recsys-eval/configs/examples/dataset.ecommerce-mini.jsonl.yaml}"
EVAL_CONFIG="${EVAL_CONFIG:-${ROOT_DIR}/recsys-eval/configs/eval/offline.ecommerce-mini.yaml}"

cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

fail() {
  printf 'commercial proof kit smoke failed: %s\n' "$1" >&2
  exit 1
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    fail "missing required command: $1"
  fi
}

check_required_path() {
  local path="$1"
  if [ ! -e "${path}" ]; then
    printf 'commercial proof kit missing required path: %s\n' "${path}" >&2
    return 1
  fi
}

check_required_inputs() {
  check_required_path "${DATA_DIR}/pipelines/exposure.jsonl"
  check_required_path "${DATA_DIR}/eval/exposures.jsonl"
  check_required_path "${DATA_DIR}/eval/outcomes.jsonl"
  check_required_path "${DATA_DIR}/eval/assignments.jsonl"
  check_required_path "${PIPELINES_CFG}"
  check_required_path "${EVAL_DATASET}"
  check_required_path "${EVAL_CONFIG}"
}

assert_missing_path_failure_is_safe() {
  local output
  local status
  local secret_canary="secret-value-that-must-not-print"

  set +e
  output="$(
    PROOF_KIT_CHECK_ONLY=1 \
    PROOF_KIT_DATA_DIR="${TMP_DIR}/missing-fixture" \
    RECSYS_SECRET_CANARY="${secret_canary}" \
    bash "$0" 2>&1
  )"
  status="$?"
  set -e

  if [ "${status}" -eq 0 ]; then
    fail "missing fixture check unexpectedly succeeded"
  fi
  case "${output}" in
    *"${TMP_DIR}/missing-fixture/pipelines/exposure.jsonl"*) ;;
    *) fail "missing fixture error did not include the missing path" ;;
  esac
  case "${output}" in
    *"${secret_canary}"*|*"DATABASE_URL="*|*"POSTGRES_PASSWORD="*|*"SECRET="*|*"PASSWORD="*)
      fail "missing fixture error leaked an env value"
      ;;
  esac

  set +e
  output="$(
    PROOF_KIT_CHECK_ONLY=1 \
    PIPELINES_CFG="${TMP_DIR}/missing-config.json" \
    RECSYS_SECRET_CANARY="${secret_canary}" \
    bash "$0" 2>&1
  )"
  status="$?"
  set -e

  if [ "${status}" -eq 0 ]; then
    fail "missing config check unexpectedly succeeded"
  fi
  case "${output}" in
    *"${TMP_DIR}/missing-config.json"*) ;;
    *) fail "missing config error did not include the missing path" ;;
  esac
  case "${output}" in
    *"${secret_canary}"*|*"DATABASE_URL="*|*"POSTGRES_PASSWORD="*|*"SECRET="*|*"PASSWORD="*)
      fail "missing config error leaked an env value"
      ;;
  esac
}

assert_file_nonempty() {
  local path="$1"
  if [ ! -s "${path}" ]; then
    fail "expected non-empty file: ${path}"
  fi
}

if [ "${PROOF_KIT_CHECK_ONLY:-0}" = "1" ]; then
  check_required_inputs
  exit 0
fi

require_cmd go
require_cmd make

check_required_inputs
assert_missing_path_failure_is_safe

rm -rf "${PROOF_DIR}"
mkdir -p "${PROOF_DIR}/eval"

echo "Building recsys-eval..."
(cd "${ROOT_DIR}/recsys-eval" && make build >/dev/null)

echo "Validating ecommerce-mini eval fixtures..."
(cd "${ROOT_DIR}/recsys-eval" && ./bin/recsys-eval validate --schema exposure.v1 --input "${DATA_DIR}/eval/exposures.jsonl")
(cd "${ROOT_DIR}/recsys-eval" && ./bin/recsys-eval validate --schema outcome.v1 --input "${DATA_DIR}/eval/outcomes.jsonl")
(cd "${ROOT_DIR}/recsys-eval" && ./bin/recsys-eval validate --schema assignment.v1 --input "${DATA_DIR}/eval/assignments.jsonl")

echo "Running offline evaluation report..."
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

echo "Running pipelines against ecommerce-mini fixture..."
(cd "${ROOT_DIR}/recsys-pipelines" && GOWORK=off go run ./cmd/recsys-pipelines \
  run --config "${PIPELINES_CFG}" --tenant demo --surface home --segment default --start 2026-01-01 --end 2026-01-01)

MANIFEST_PATH="${PROOF_DIR}/pipelines/registry/current/demo/home/manifest.json"
assert_file_nonempty "${PROOF_DIR}/eval/offline-report.json"
assert_file_nonempty "${PROOF_DIR}/eval/offline-report.md"
assert_file_nonempty "${MANIFEST_PATH}"
grep -q 'popularity' "${MANIFEST_PATH}" || fail "manifest does not reference popularity artifact"
find "${PROOF_DIR}/pipelines/objectstore" -type f | grep -q . || fail "expected published object-store artifacts"

echo "commercial proof kit smoke passed"
