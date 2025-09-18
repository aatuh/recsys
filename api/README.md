# README

An HTTP API for ingesting activity and asking for recommendations.

## API Endpoints

Open **Swagger** at **`/docs`** to inspect schemas and try requests.

## Development

- See the **Makefile** in the repo root for dev/test/migration commands
  (e.g., `make dev`, `make test`).  
- Hot reload in dev; Swagger is generated from annotations.

## Deploying to Railway

- Create new project.
- Paste env variables.
- Set connected branch as `production`.
- Root directory `/api/`.
- Generate custom domain.
