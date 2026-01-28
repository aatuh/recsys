# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS base

WORKDIR /app

ENV NEXT_TELEMETRY_DISABLED=1

RUN corepack enable && corepack prepare pnpm@10.19.0 --activate

FROM base AS deps

RUN apk add --no-cache git

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

FROM deps AS builder

COPY . .
RUN pnpm build
RUN mkdir -p public
RUN pnpm prune --prod

FROM node:22-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production \
	NEXT_TELEMETRY_DISABLED=1 \
	PORT=3000 \
	HOSTNAME=0.0.0.0

COPY --chown=node:node --from=builder /app/package.json /app/package.json
COPY --chown=node:node --from=builder /app/public /app/public
COPY --chown=node:node --from=builder /app/.next /app/.next
COPY --chown=node:node --from=builder /app/node_modules /app/node_modules
COPY --chown=node:node --from=builder /app/next.config.ts /app/next.config.ts

EXPOSE 3000

USER node

CMD ["node", "node_modules/next/dist/bin/next", "start", "-p", "3000", "-H", "0.0.0.0"]
