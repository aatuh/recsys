# Recsys Backend (Dev Boilerplate)

Go API + Postgres (Docker Compose) with hot reloading via Air, and Atlas migrations.

## Quickstart

```bash
cp .env.example .env
make dev
# in another shell:
make migrate-apply
# open http://localhost:8000/health  -> {"status":"ok"}
# optional (generate Swagger UI):
make swag
# open http://localhost:8000/swagger/index.html
```

## Commands

- `make dev` — start Postgres and API with hot reload.
- `make migrate-apply` — apply SQL migrations using Atlas.
- `make swag` — generate Swagger docs from annotations.
- `make down` — stop and remove containers and volumes.

## Endpoints (stubs)

- `GET /health` — liveness check.
- `POST /v1/items:upsert`
- `POST /v1/users:upsert`
- `POST /v1/events:batch`
- `POST /v1/recommendations`
- `GET  /v1/items/{item_id}/similar`

Auth, RLS, and full ranking come later.
