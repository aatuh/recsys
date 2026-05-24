# Developer Quickstart

This quickstart gives a new developer one local service run and one recommendation request. It uses the Docker Compose
stack and development header auth from `api/.env.example`.

## Prerequisites

- Docker with Compose v2.
- Go installed for module-level commands.
- Python 3 and Node.js 22.12 or newer if you run documentation or site checks.

## Start the stack

From the repository root:

```bash
make env
make dev
```

Expected result:

- `api/.env` exists, copied from `api/.env.example` if it was missing.
- Compose starts `db` and `api`.
- The API listens on `http://localhost:8000`.

Check health:

```bash
curl -f http://localhost:8000/healthz
curl -f http://localhost:8000/readyz
```

`/healthz` checks liveness. `/readyz` checks whether runtime dependencies are ready.

## Apply migrations when needed

The default local env has `MIGRATE_ON_START=true`. If you need to apply migrations manually:

```bash
cd api
make migrate-up
```

Expected result: migrations in `api/migrations/` are applied to the Compose Postgres database.

## Make a recommendation request

Local development auth accepts dev identity headers. Use a tenant header and a pseudonymous user identifier:

```bash
curl -sS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Org-Id: demo' \
  -H 'X-Dev-User-Id: local-dev' \
  -H 'X-Dev-Org-Id: demo' \
  -d '{
    "surface": "home",
    "k": 5,
    "user": {
      "anonymous_id": "anon-local-1",
      "session_id": "session-local-1"
    },
    "context": {
      "device": "web",
      "country": "FI"
    }
  }'
```

Expected result: a JSON response with `items`, `meta`, and optional `warnings`. Empty recommendations are still a
valid local result if no catalog or artifact data has been loaded.

## Run the smoke-tested success path

For a non-empty recommendation, pinned-rule assertion, exposure log, and schema validation, run the maintained
end-to-end smoke path:

```bash
bash scripts/tutorial_smoke_test.sh
```

Expected result: the script prints `Tutorial smoke test OK`. It bootstraps a `demo` tenant, writes tenant config and
rules, seeds DB-only item signals, verifies `item_3` is pinned first, creates an eval-compatible exposure log, and
validates that exposure log with `recsys-eval`.

Read the step-by-step version in [Local end-to-end](local-end-to-end.md).

## Validate a request without serving it

```bash
curl -sS http://localhost:8000/v1/recommend/validate \
  -H 'Content-Type: application/json' \
  -H 'X-Org-Id: demo' \
  -H 'X-Dev-User-Id: local-dev' \
  -H 'X-Dev-Org-Id: demo' \
  -d '{"surface":"home","k":5}'
```

Expected result: normalized request fields and any validation warnings.

## Stop the stack

```bash
make down
```

Expected result: Compose services and volumes for the local stack are removed.

## Next steps

- Use [Integration and evaluation](integration.md) to connect web, mobile, or desktop clients.
- Use [Local end-to-end](local-end-to-end.md) to reproduce the smoke-tested first-success path.
- Use [API reference](reference/api.md) for endpoint and payload details.
- Use [Local workflow](reference/local-workflow.md) for repo quality gates.
