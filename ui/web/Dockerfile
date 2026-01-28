FROM node:20-alpine

# Enable pnpm via corepack and pin a modern version
RUN corepack enable && corepack prepare pnpm@9.12.2 --activate

WORKDIR /app
ENV CI=1 PNPM_HOME=/root/.local/share/pnpm PATH="/root/.local/share/pnpm:${PATH}"

# Keep image lean; deps installed at runtime by entrypoint
COPY package.json pnpm-lock.yaml* ./
COPY . .

RUN apk add --no-cache curl

EXPOSE 3000
CMD ["/bin/sh", "-c", "/app/entrypoint.sh"]
