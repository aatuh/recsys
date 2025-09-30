# Repository Guidelines

## Project Structure & Module Organization
Core services are orchestrated through `docker-compose.yml` and the root `Makefile`. The Go API lives in `api/`, with HTTP handlers under `api/`, business logic and shared utilities in `internal/`, migrations in `migrations/`, and integration fixtures in `test/`. The Vite + React UI sits in `web/`; generated API clients are stored in `web/src/lib/api-client` and should not be edited manually. Database seeds are defined in `db/initdb/`, and the edge proxy Dockerfile resides in `proxy/`.

## Build, Test, and Development Commands
Use `make dev` to boot API, Postgres, proxy, and UI together; `make down` tears them down and resets volumes. Rebuild containers with `make build`. Refresh Swagger and regenerate the web API client with `make codegen` before touching UI that relies on new API fields. Execute Go tests via `make test`, or scope with `make test PKG=./internal/...` or `make test TEST_PATTERN=Recommend`. Inside `web/`, `pnpm dev` previews the UI, `pnpm build` produces bundles, and `pnpm lint` plus `pnpm typecheck` should succeed before commits.

## Coding Style & Naming Conventions
Keep Go code formatted with `gofmt`; packages use lower_snake_case, and exports use PascalCase. Co-locate new handlers with routes and prefer structured `zap` logging. TypeScript follows `web/eslint.config.js`; components adopt PascalCase, hooks and utilities use camelCase, and generated clients remain untouched. Default to ASCII unless existing files require otherwise.

## Testing Guidelines
Go tests belong in `*_test.go` files alongside the target package. Use `make test` for the disposable containerized run, and lean on `PKG` or `TEST_PATTERN` for faster iteration. UI work must pass `pnpm lint` and `pnpm typecheck`; add `.test.tsx` files next to the component under test when introducing new coverage.

## Commit & Pull Request Guidelines
Write commits in sentence case with imperative verbs ending in a period, e.g., `Refine bandit defaults.`. Group related code, schema, and generated artifacts within the same change. PRs should explain intent, list validation (tests, screenshots), and link issues. Flag breaking migrations or config updates early and request reviewers for the impacted surface (`api`, `web`, infra).

## Environment & Configuration Tips
Copy `.env.example` files in `api/` and `web/` to configure local runs. `docker-compose` loads these automatically. During API development, verify contract updates against `http://localhost:8081/swagger.json` before regenerating clients.
