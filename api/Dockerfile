FROM golang:1.24-alpine AS build
WORKDIR /app
RUN apk add --no-cache build-base git ca-certificates tzdata

# Copy go mod first for caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy the rest
COPY . .

# Build the API and the migrate helper
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# ---- runtime stage ----
FROM alpine:3.20
WORKDIR /srv
RUN apk add --no-cache ca-certificates tzdata curl
COPY --from=build /out/api /srv/api
COPY --from=build /out/migrate /srv/migrate

# Include SQL migrations directory for MIGRATE_ON_START flow
COPY migrations /srv/migrations

# Include swagger directory for API documentation
COPY swagger /srv/swagger

# Healthcheck (simple)
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD wget -qO- http://127.0.0.1:${PORT:-8000}/docs >/dev/null 2>&1 || exit 1

# On start, ensure API_PORT mirrors Railway's PORT, then launch
ENTRYPOINT ["/bin/sh","-lc","API_PORT=${PORT:-8000} exec /srv/api"]
