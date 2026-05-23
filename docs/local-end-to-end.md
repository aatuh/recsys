# Local End-to-End

This page documents the maintained first-success path from `scripts/tutorial_smoke_test.sh`. Use it when you need more
than a health check: a tenant, config, rules, seeded signals, a non-empty recommendation, an exposure log, and schema
validation.

## Prerequisites

- Docker with Compose v2.
- Go, because the script builds `recsys-eval`.
- `curl`, `python3`, and `bash`.

## Run the maintained smoke path

From the repository root:

```bash
bash scripts/tutorial_smoke_test.sh
```

Expected result:

```text
Tutorial smoke test OK
```

The script restores the previous `api/.env` on exit and tears down the Compose stack.

## What the script proves

| Step | Evidence of success |
| --- | --- |
| Prepare local config | `api/.env` is copied from `api/.env.example` and overridden for DB-only mode. |
| Start service | `db` and `api` start through `docker compose`. |
| Health check | `http://localhost:8000/healthz` returns success. |
| Migrations | `api/migrations/` are applied idempotently. |
| Tenant bootstrap | Tenant `demo` exists in Postgres. |
| Config write | `PUT /v1/admin/tenants/demo/config` succeeds. |
| Rules write | `PUT /v1/admin/tenants/demo/rules` pins `item_3` on `home`. |
| Signal seed | `item_tags` and `item_popularity_daily` contain three demo items. |
| Recommendation | `POST /v1/recommend` returns non-empty `items`. |
| Rule assertion | The first returned item is `item_3`. |
| Exposure log | `/app/tmp/exposures.eval.jsonl` exists inside the service container. |
| Schema validation | `recsys-eval validate --schema exposure.v1` accepts the copied exposure log. |

## Manual inspection commands

Run these while the script is active or adapt them for a local stack that you started yourself:

```bash
curl -fsS http://localhost:8000/healthz
docker compose logs --tail=100 api
docker exec recsys-db psql -U recsys-db -d recsys-db -c "select external_id from tenants;"
```

To inspect the recommendation output after adapting the script manually:

```bash
cat /tmp/recommend.json
cat /tmp/exposures.jsonl
```

Expected result: the recommendation response contains `items`, and the exposure log contains JSONL records accepted by
the `recsys-eval` exposure schema.

## Failure recovery

- If Docker services fail to start, run `docker compose logs --tail=100 api db`.
- If migrations fail, run `cd api && make migrate-status`.
- If the recommendation is empty, check tenant config, rules, seeded item tags, and seeded popularity rows.
- If exposure validation fails, keep the copied `/tmp/exposures.jsonl` and run `cd recsys-eval && ./bin/recsys-eval
  validate --schema exposure.v1 --input /tmp/exposures.jsonl`.
