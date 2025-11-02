#!/bin/sh
set -eu

# Ensure pnpm is available
corepack enable >/dev/null 2>&1 || true
corepack prepare pnpm@9.12.2 --activate >/dev/null 2>&1 || true

# Install deps if node_modules is missing or empty (first-run volume)
if [ ! -d node_modules ] || [ -z "$(ls -A node_modules 2>/dev/null || true)" ]; then
  echo "Installing dependencies with pnpm..."
  pnpm install
fi

# Generate Prisma client if schema exists
if [ -f prisma/schema.prisma ]; then
  echo "Generating Prisma client..."
  pnpm prisma:generate || true
  echo "Applying Prisma migrations..."
  if [ "${NODE_ENV:-development}" = "production" ]; then
    pnpm exec prisma migrate deploy || true
  else
    pnpm exec prisma db push --accept-data-loss || true
  fi
fi

# Run dev by default if NODE_ENV not production
if [ "${NODE_ENV:-development}" = "production" ]; then
  pnpm build && pnpm start
else
  pnpm dev
fi

