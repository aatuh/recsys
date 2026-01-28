#!/bin/sh
set -eu

URL="${OPENAPI_URL:-http://recsys-swagger:8080/swagger.json}"

# Install deps once (mounted volume persists node_modules)
# If the volume created an empty node_modules directory, ensure we still install
if [ ! -f node_modules/.bin/vite ]; then
  echo "Installing deps with pnpm..."
  pnpm install --frozen-lockfile=false
fi

# Wait for Swagger service to expose swagger.json
echo "Waiting for Swagger schema at: $URL"
for i in $(seq 1 60); do
  if curl -fsSL "$URL" >/dev/null 2>&1; then break; fi
  sleep 1
done

echo "Starting Vite dev server..."
exec pnpm dev --host 0.0.0.0 --port 3000
