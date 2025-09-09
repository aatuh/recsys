#!/bin/sh
set -eu

URL="${OPENAPI_URL:-http://recsys-api:8000/swagger/swagger.json}"

# Install deps once (mounted volume persists node_modules)
if [ ! -d node_modules ]; then
  echo "Installing deps with pnpm..."
  pnpm install --frozen-lockfile=false
fi

# Wait for API to expose swagger.json
echo "Waiting for API schema at: $URL"
for i in $(seq 1 60); do
  if curl -fsSL "$URL" >/dev/null 2>&1; then break; fi
  sleep 1
done

echo "Starting Vite dev server..."
exec pnpm dev --host 0.0.0.0 --port 3000
