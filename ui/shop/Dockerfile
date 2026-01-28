FROM node:20-alpine

# Enable pnpm via corepack and pin version
RUN corepack enable && corepack prepare pnpm@9.12.2 --activate

WORKDIR /app
ENV CI=1 PNPM_HOME=/root/.local/share/pnpm PATH="/root/.local/share/pnpm:${PATH}"

COPY package.json pnpm-lock.yaml* ./
COPY . .

RUN apk add --no-cache curl openssl1.1-compat || apk add --no-cache curl openssl

EXPOSE 3002
CMD ["/bin/sh", "-c", "/app/entrypoint.sh"]

