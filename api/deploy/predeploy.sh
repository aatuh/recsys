#!/usr/bin/env sh
set -eu

# Require DATABASE_URL
: "${DATABASE_URL:?DATABASE_URL must be set}"

mask_url() {
  # Mask password for logging: postgres://user:***@host/db?...
  printf %s "$1" | sed -E 's#(postgres(ql)?://[^:]+:)[^@]+#\1***#'
}

# Normalize scheme to postgres://
NORM_URL="$(printf %s "$DATABASE_URL" | sed -E 's/^postgresql:/postgres:/')"

# Split URL into base and query string
BASE="${NORM_URL%%\?*}"
QS=""
case "$NORM_URL" in
  *\?*) QS="${NORM_URL#*\?}";;
esac

# Keep only safe/known params for Atlas
# Add here if you need more: application_name, search_path, timezone, etc.
ALLOWED='^(sslmode|application_name|search_path|timezone)='
SAFE_QS=""
if [ -n "$QS" ]; then
  # split on &
  IFS='&'
  for kv in $QS; do
    if printf %s "$kv" | grep -Eq "$ALLOWED"; then
      if [ -n "$SAFE_QS" ]; then SAFE_QS="${SAFE_QS}&${kv}"; else SAFE_QS="${kv}"; fi
    fi
  done
  unset IFS
fi

# Always ensure sslmode=require for Neon unless you already provide one
case "$SAFE_QS" in
  *sslmode=*) : ;;
  "" ) SAFE_QS="sslmode=require" ;;
  * ) SAFE_QS="${SAFE_QS}&sslmode=require" ;;
esac

CLEAN_URL="${BASE}?${SAFE_QS}"

echo "[predeploy] DATABASE_URL (masked): $(mask_url "$CLEAN_URL")"

# Install latest Atlas if not present
if ! command -v atlas >/dev/null 2>&1; then
  echo "[predeploy] installing atlas (latest)"
  curl -sSfL https://atlasgo.sh | sh -s -- -b /usr/local/bin
fi
atlas version || true

# Apply migrations
atlas migrate apply \
  --url "$CLEAN_URL" \
  --dir file://migrations \
  --revisions-schema public
