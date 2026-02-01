#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

rm -rf .out
mkdir -p .out

./bin/recsys-pipelines run   --config configs/env/local.json   --tenant demo   --surface home   --start 2026-01-01   --end 2026-01-01

hash_tree() {
  # Hash file contents with stable ordering.
  find .out -type f ! -name '.hashes1' ! -name '.hashes2' -print0 \
    | sort -z \
    | xargs -0 sha256sum
}

hash_tree > .out/.hashes1

# Run again and ensure idempotency for the tiny dataset.
./bin/recsys-pipelines run \
  --config configs/env/local.json \
  --tenant demo \
  --surface home \
  --start 2026-01-01 \
  --end 2026-01-01

hash_tree > .out/.hashes2

if ! diff -u .out/.hashes1 .out/.hashes2 >/dev/null; then
  echo "Smoke failed: output is not idempotent" >&2
  diff -u .out/.hashes1 .out/.hashes2 || true
  exit 1
fi

echo "Smoke OK: .out/ created"
