#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

fail() {
  printf 'env bootstrap regression failed: %s\n' "$1" >&2
  exit 1
}

fixture() {
  local name="$1"
  local dir="${TMP_DIR}/${name}"

  mkdir -p "${dir}/api"
  cp "${ROOT_DIR}/api/Makefile" "${dir}/api/Makefile"
  printf '%s\n' "${dir}"
}

run_make() {
  local dir="$1"
  local target="$2"
  local output

  if ! output="$(make -s --no-print-directory -C "${dir}/api" "${target}" 2>&1)"; then
    fail "make ${target} failed unexpectedly: ${output}"
  fi

  printf '%s' "${output}"
}

assert_same_file() {
  local expected="$1"
  local actual="$2"

  cmp -s "${expected}" "${actual}" || fail "${actual} does not match ${expected}"
}

assert_missing_example_failure() {
  local dir="$1"
  local target="$2"
  local missing_path="$3"
  local failure_output
  local failure_status
  local secret_canary

  secret_canary="secret-value-that-must-not-print"
  set +e
  failure_output="$(RECSYS_SECRET_CANARY="${secret_canary}" make -s --no-print-directory -C "${dir}/api" "${target}" 2>&1)"
  failure_status="$?"
  set -e

  if [ "${failure_status}" -eq 0 ]; then
    fail "${target} succeeded without ${missing_path}"
  fi

  case "${failure_output}" in
    *"${missing_path}"*) ;;
    *) fail "missing example error did not include ${missing_path}" ;;
  esac

  case "${failure_output}" in
    *"${secret_canary}"*|*"DATABASE_URL="*|*"POSTGRES_PASSWORD="*|*"PASSWORD="*|*"SECRET="*)
      fail "missing example error leaked an env value"
      ;;
  esac
}

happy_dir="$(fixture happy)"
printf 'API_ADDR=:8000\n' > "${happy_dir}/api/.env.example"
printf 'API_HOST=http://recsys-svc:8000\n' > "${happy_dir}/api/.env.test.example"

run_make "${happy_dir}" env >/dev/null
run_make "${happy_dir}" test-env >/dev/null
assert_same_file "${happy_dir}/api/.env.example" "${happy_dir}/api/.env"
assert_same_file "${happy_dir}/api/.env.test.example" "${happy_dir}/api/.env.test"

existing_dir="$(fixture existing)"
printf 'LOCAL_ONLY=true\n' > "${existing_dir}/api/.env"
printf 'EXAMPLE_ONLY=true\n' > "${existing_dir}/api/.env.example"
printf 'LOCAL_TEST_ONLY=true\n' > "${existing_dir}/api/.env.test"
printf 'EXAMPLE_TEST_ONLY=true\n' > "${existing_dir}/api/.env.test.example"

run_make "${existing_dir}" env >/dev/null
run_make "${existing_dir}" test-env >/dev/null
grep -qx 'LOCAL_ONLY=true' "${existing_dir}/api/.env" || fail "api/.env was overwritten"
grep -qx 'LOCAL_TEST_ONLY=true' "${existing_dir}/api/.env.test" || fail "api/.env.test was overwritten"

missing_env_dir="$(fixture missing-env)"
missing_test_env_dir="$(fixture missing-test-env)"
assert_missing_example_failure "${missing_env_dir}" env "api/.env.example"
assert_missing_example_failure "${missing_test_env_dir}" test-env "api/.env.test.example"

echo "env bootstrap regression passed"
