# recsys

Production-ready starter for HTTP APIs. It wires a clean architecture,
sensible defaults, and a shared toolkit module (`github.com/aatuh/api-toolkit`)
for routing, middleware, logging, DB access, migrations, validation, and docs.

## Quick start

Find and replace string `recsys` with your service name:

```regex
recsys(?!-core)
```

Find each `.env.example` file and create `.env` files.

```bash
# Start dev stack (API + web + DB) with live reload
make dev

# Apply migrations (inside docker)
cd api && make migrate-up

# Build local binary with stamped version info
cd api && make build-bin

# Health check and docs
make health   # GET http://localhost:8000/health
              # Docs at http://localhost:8000/docs
              # Web at http://localhost:3000
```

## Useful Prompt

I'm working on recsys project. Please refer to AGENTS.md backlog.md files in it. use /api-toolkit and /api-svc as reusable libraries. Always create best industry standard, hexagonal architecture, SOLID principles, clear, testable and developer friendly code. Your task: write task here.

## Directory structure

```plaintext
.
├── Makefile                   # top-level helpers that delegate to ./api
├── docker-compose.yml         # dev stack (api, db, test runner)
├── README.md                  # this file
├── api/                       # the actual API service
│   ├── Makefile               # API tasks (swag, test, lint, health, migrate)
│   ├── Dockerfile             # production image
│   ├── Dockerfile.dev         # dev image with hot-reload
│   ├── go.mod, go.sum         # module (depends on github.com/aatuh/api-toolkit)
│   ├── cmd/
│   │   ├── api/               # HTTP server entrypoint
│   │   │   └── main.go
│   │   └── migrate/           # CLI for DB migrations
│   │       └── main.go
│   ├── internal/              # app internals (services, stores, http)
│   │   ├── services/          # domain services (e.g. foosvc)
│   │   ├── store/             # repositories (SQL via pgx)
│   │   ├── validation/        # request validation helpers
│   │   └── http/
│   │       ├── handlers/      # thin HTTP handlers (mount under routes)
│   │       ├── mapper/        # DTO <-> domain mapping
│   │       └── ...
│   ├── migrations/            # SQL migrations (embedded) + go:embed binder
│   ├── src/
│   │   └── specs/             # API endpoint paths and public types
│   ├── swagger/               # generated OpenAPI docs
│   └── test/                  # integration tests (run in container)
└── web/                       # Next.js app + shared packages
    ├── src/                   # app router pages and UI
    ├── content/               # markdown content pages
    └── packages/
        └── services/foo/      # demo domain, adapters, hooks, config
```

## Core packages (web)

The web app consumes `@recsys/*` from the `api-boilerplate-core` repo
via git dependencies.

## Key components

- Toolkit bootstrap
  - `github.com/aatuh/api-toolkit/bootstrap`: `OpenAndPingDB`, `NewDefaultRouter`,
    `MountSystemEndpoints`, `StartServer`, `NewMigrator`.
- Entrypoints
  - `api/cmd/api`: loads config, opens DB, runs migrations on start
    (optional), wires services/handlers, starts HTTP server.
  - `api/cmd/migrate`: CLI to run `up`, `down`, `status` using the
    same embedded migrations as the server.
- Domain
  - `api/internal/services/foosvc`: example service showing patterns for
    validation, transactions, IDs, and clock usage.
  - `api/internal/store`: data access with `pgx` pools and context.
- HTTP
  - `api/internal/http/handlers`: decode → validate → service → encode.
  - Health at `/health`, metrics at `/metrics`, docs at `/docs`.

## Local development with `api-toolkit`

This repo depends on the released module version of `github.com/aatuh/api-toolkit`. For local cross-repo development, prefer a Go workspace instead of committing `replace` directives.

From a parent directory that contains both repos:

```bash
go work init ./recsys/api
go work use ./api-toolkit
```

From the `recsys/` repo root:

```bash
go work init ./api
go work use ../api-toolkit
```

To verify you’re using the published dependency versions (CI-like), run with `GOWORK=off`.

Docker Compose can also opt into the local toolkit checkout:

```bash
make dev TOOLKIT=1 # defaults to ../api-toolkit

# Optional: override toolkit checkout path
make dev TOOLKIT=1 API_TOOLKIT_PATH=$HOME/src/api-toolkit

# Equivalent docker compose invocation:
API_TOOLKIT_PATH=$HOME/src/api-toolkit docker compose -f docker-compose.yml -f docker-compose.toolkit.yml up -d
```

## Environment variables

- `api/.env`           local dev for API (not committed)
- `api/.env.example`   example of required vars for API
- `api/.env.test`      local test env (not committed)
- `api/.env.test.example` example for tests
- `/.env`              docker compose env (not committed)

Rules:

- Load env at startup; fail fast if required variables are missing.
- Document new envs in the corresponding `.env.example` files.
- Integration tests must use a separate `.env.test`.

## Common commands

Top-level delegates into `./api`:

```bash
make dev          # docker compose up (hot reload)
make down         # stop and clean volumes
make build        # build images
make codegen      # generate swagger and sync artifacts
make test         # run tests (inside container)
make fmt          # gofmt -s -w
make lint         # go vet + golangci-lint
make health       # show logs + curl health endpoint
```

API-specific (from `api/`):

```bash
make codegen              # regen swagger from cmd/api/main.go
make migrate-up           # apply migrations
make migrate-down         # rollback (dangerous; off in server)
make migrate-status       # show applied/pending migrations
make build-bin            # go build with version metadata (bin/api)
```

## Build metadata

To embed version info in the binary, run `cd api && make build-bin`. The
target stamps the git describe, commit SHA, and UTC build time via `-ldflags`
before producing `bin/api`. You can override these values:

```bash
cd api
VERSION=1.2.3 COMMIT=$(git rev-parse HEAD) DATE=$(date -u +%FT%TZ) make build-bin
```

The version endpoint in `cmd/api` picks up the injected values automatically.

## Migrations

- SQL lives in `api/migrations/*.up.sql` and `*.down.sql` with timestamped
  names.
- The API can run `up` on start if `MIGRATE_ON_START=true`.
- Sources:
  - Embedded (default): bundled via `go:embed` (`migrations_embed.go`).
  - Directory: set `MIGRATIONS_DIR=/path/to/sql` to override.

## Development flow

1) Model your domain in `api/internal/services` and `api/internal/store`.
2) Add HTTP handlers in `api/internal/http/handlers` and mount under
   `api/src/specs/endpoints` paths.
3) Add/modify migrations in `api/migrations` and run `make migrate-up`.
4) Regenerate docs with `make codegen`.
5) Run `make fmt`, `make lint`, `make test`, `make health`.

## Customizing the boilerplate

- Replace `foosvc` with your service name and follow the same wiring in
  `cmd/api/main.go` (repositories → services → handlers → routes).
- Keep handlers thin and push logic into services.
- Always accept `context.Context` for blocking or external operations.
- Use the provided logger, validator, ID generator, and clock via
  dependency injection for testability.

## Endpoints (default)

- Health: `GET /health`
- Metrics: `GET /metrics` (Prometheus)
- Docs: `GET /docs`
- Version: `GET /version`
