# README

An HTTP API for ingesting activity and asking for recommendations.

## API Endpoints

Open **Swagger** at **`/docs`** to inspect schemas and try requests.

### Rule engine (admin)

- `GET /v1/admin/rules` — list merchandising rules with optional filters.
- `POST /v1/admin/rules` — create a rule (`BLOCK`, `PIN`, `BOOST`).
- `PUT /v1/admin/rules/{rule_id}` — update an existing rule.
- `POST /v1/admin/rules/dry-run` — preview matched rules and effects without
  mutating state.

### Explain RCA

- `POST /v1/explain/llm` — generate a markdown RCA report using curated facts
  and the configured LLM; response includes both the human-readable summary and
  the facts pack used (requires `LLM_EXPLAIN_ENABLED=true` and `LLM_API_KEY`).

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
