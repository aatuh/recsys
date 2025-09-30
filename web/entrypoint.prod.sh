#!/bin/sh
set -eu

# Production entrypoint for web
# Handles CSP, caching, and production optimizations

echo "Starting production web server..."

# Set production environment variables
export NODE_ENV=production
export VITE_ALLOWED_HOSTS="${VITE_ALLOWED_HOSTS:-localhost,0.0.0.0,127.0.0.1}"

# Wait for API service to be ready (optional, for health checks)
if [ -n "${API_HEALTH_CHECK_URL:-}" ]; then
  echo "Waiting for API service at: $API_HEALTH_CHECK_URL"
  for i in $(seq 1 30); do
    if curl -fsSL "$API_HEALTH_CHECK_URL" >/dev/null 2>&1; then 
      echo "API service is ready"
      break
    fi
    echo "Waiting for API service... ($i/30)"
    sleep 2
  done
fi

# Start the production server
echo "Starting Vite preview server with production optimizations..."
exec pnpm preview --host 0.0.0.0 --port 3000
