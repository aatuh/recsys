# Repository Guidelines

## Project Structure & Module Organization

- `api/` holds the HTTP service (`cmd/api` entrypoint, `internal/*` domain layers, `migrations/` SQL, and generated `swagger/` specs).
- `api/migrations/` defines the seeds, sessions, requests, and results tables plus the global audit chain state; keep new migrations in timestamped `*.up.sql`/`*.down.sql` pairs.
- `swagger/public/` exposes the synced OpenAPI bundle for the nginx sidecar, while root assets (`Makefile`, `docker compose.yml`, `todo.md`) define automation and roadmap.

- Architecture invariants
  - Use `api-toolkit` adapters/interfaces for router, CORS, logging, DB, validation etc.
  - Use `api-boilerplate-core` for reusable web components.
  - Emit errors as RFC-7807 problem+json via `api-toolkit/httpx`; successes via `response_writer`.

- `api/test` contains API integration tests.

- Everything related to "foo" is used for boilerplate and example code. It is used to demonstrate code style and patterns. Do not use "foo" related code or files for actual feature development, only as reference.

## Environment Variables

- `api/.env` contains API environment variables for local development, not Git-committed. `api/.env.example` is the Git-committed example file for it.
- `api/.env.test` contains API environment variables for local development integration tests, not Git-committed. `api/.env.test.example` is the Git-committed example file for it.
- `/.env` contains Docker Compose environment variables for local development, not Git-committed.
- `api/.env.example` is the Git-committed example template file for all env files.
- `/.env.prod` contains Docker Compose environment variables for production, not Git-committes. It is copied from `api/.env.example` file.

- Load env at startup; fail fast on missing vars.
- Define new environment variables in corresponding `.env` and `.env.example` file.
- Keep example env and the actual files in sync.
- Example env files must not contain secrets but can contain sensible defaults for that particular environment.
- Integration tests must use a separate `.env.test` file and own config structure with only needed variables.
- Docker Compose should use the `/.env` file.
- Always load environment variables on program launch phase.
- In code environment variables must exist i.e. program must panic/fail if environment variable doesn't exist.

## Running Code

Service runs in hot reloaded Docker Compose container and is probably already started by the human operator but verify this on suitable error situations and start service with proper `make` command.

## Build, Test, and Development Commands

Most useful commands are the following:

- `make cycle` stop and then start the full stack via Docker Compose with live reload. If you change environment variables always run make cycle after that.
- `make build` build Docker Compose images without cache.
- `make codegen` regenerate OpenAPI (`api/swagger`) and sync artifacts into `swagger/public/`.
- `make finalize` runs everything needed to verify code is good quality and works.
- `make health` service healthiness check.

- In any commands prefer non-interactive flags; run long-lived commands in the background.
- Primary targets are the `make` commands.

`/Makefile` contains all the commands.

## Architecture, Coding Style & Naming Conventions

- Use the `github.com/aatuh/api-toolkit` package as basis/framework of the project.
  - If there are new features that can be re-used in other projects they belong to the `api-toolkit` package. In that case suggest making this addition.
- Do not edit auto-generated files; regenerate instead (e.g., `make codegen`).
- Follow idiomatic language naming patterns: e.g Golang exported identifiers use `CamelCase`, package-private code stays lowerCamel, and tests mirror package names.
- Keep handlers and services thin.
- Use SOLID principles and hexagonal architecture to achieve maximum modularity and flexibility.
- Write secure, robust, scalable, extendable, performant, clear, developer-friendly code.
- Write idiomatic, industry best practice code and patterns.

## Testing Guidelines

- Both unit tests and integration tests must cover reasonable and sensible amount of the system's functionality.
- Unit tests reside alongside code (`*_test.go`). Extend them when adding capabilities or utilities.
- Integration tests reside in `api/test`.
- Tests should use random data for testing.
- Use `make test` for all tests.
- Use `make health` to verify service is functional.
- Integration tests always run against the actual running service or database.

## Feature Completion

- A feature is ready when the code works as intended, both unit and integration test coverage is sufficient, tests succeed, `make codegen`, `make fmt`, `make lint`, `make health` succeed.
- If using a backlog or todo file, mark complete items with a checkmark emoji or with markdown box `[x]` when applicable.

## Commit Guidelines

- Only commit when requested.
- Commit only when feature criteria are met; use Conventional Commits; keep diffs minimal.
- Prefer Conventional Commit-style subjects (e.g., `feat(api): add bytes handler`) in active voice.
