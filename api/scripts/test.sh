#!/usr/bin/env bash
set -euo pipefail

if [[ "${SKIP_API_WAIT:-false}" != "true" && -n "${API_HOST:-}" ]]; then
  health_url="${API_HOST%/}/health"
  timeout_secs="${API_WAIT_TIMEOUT_SECONDS:-90}"
  interval_secs="${API_WAIT_INTERVAL_SECONDS:-2}"
  start_ts="$(date +%s)"

  echo "Pre-waiting for API server at ${health_url} (timeout ${timeout_secs}s)"
  while true; do
    if curl -fsS --max-time 2 "${health_url}" >/dev/null 2>&1; then
      echo "API server is ready."
      export SKIP_API_WAIT=true
      break
    fi

    now_ts="$(date +%s)"
    elapsed="$((now_ts - start_ts))"
    if (( elapsed >= timeout_secs )); then
      echo "failed waiting for API: timed out waiting for API after ${timeout_secs}s" >&2
      exit 1
    fi

    sleep "${interval_secs}"
  done
fi

exec go run github.com/aatuh/api-toolkit/contrib/v2/cmd/tester
