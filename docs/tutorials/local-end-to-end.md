---
tags:
  - tutorial
  - quickstart
  - developer
  - recsys-service
  - recsys-pipelines
  - recsys-eval
---

# Tutorial: local end-to-end (service → logging → eval)

## Who this is for

- Developers who want to prove the full loop locally (serve → log → eval)

## What you will get

- A running `recsys-service` in DB-only mode (popularity baseline)
- An eval-compatible exposure log file
- A sample `recsys-eval` report you can share internally

> **Choose your data mode**
>
> This tutorial uses **DB-only mode** (fastest way to prove the loop locally).
>
> - Choose **DB-only** to validate the full loop quickly: serve → log → eval.
> - Choose **artifact/manifest mode** when you want pipelines to publish versioned artifacts and use the manifest as
>   a ship/rollback lever.
>
> See: [Choose your data mode](../start-here/choose-data-mode.md) (decision guide) and
> [Data modes](../explanation/data-modes.md) (details). For artifact mode end-to-end, follow
> [Production-like run](production-like-run.md).

## Prereqs

- Docker + Docker Compose (v2)
- `make`
- `curl`
- POSIX shell
- `python3` (used to parse the exposure log)
- Go toolchain (to build `recsys-eval`)

Verify you have them:

```bash
docker compose version
make --version
curl --version
python3 --version
go version
```

--8<-- "_snippets/key-terms.list.snippet"
--8<-- "_snippets/key-terms.defs.one-up.snippet"

## Verify (expected outcome)

- `POST /v1/recommend` returns a non-empty list for tenant `demo` and surface `home`
- A local exposure log file exists (eval schema)
- `recsys-eval run` produces a Markdown report

## 1) Start Postgres + recsys-service

From repo root:

```bash
test -f api/.env || cp api/.env.example api/.env
make dev
```

Apply database migrations (idempotent):

```bash
(cd api && make migrate-up)
```

Verify:

```bash
curl -fsS http://localhost:8000/healthz >/dev/null
```

Expected:

- `make dev` exits 0 and starts the local stack.
- The health check exits 0.

## 2) Configure local dev for a runnable tutorial

This tutorial uses dev headers for auth and disables admin RBAC roles so you can call admin endpoints without JWT
claims.

Apply these settings in `api/.env`:

<details markdown="1" open>
<summary>Tutorial env settings (copy/paste)</summary>

```bash
# DB-only mode (no artifact manifest)
RECSYS_ARTIFACT_MODE_ENABLED=false

# Make requests deterministic
RECSYS_ALGO_MODE=popularity

# Enable rules so you can prove control-plane wiring (pin/exclude) works
RECSYS_ALGO_RULES_ENABLED=true

# Enable eval-compatible exposure logs
EXPOSURE_LOG_ENABLED=true
EXPOSURE_LOG_FORMAT=eval_v1
EXPOSURE_LOG_PATH=/app/tmp/exposures.eval.jsonl

# Local dev: disable admin RBAC roles (dev headers don’t carry roles)
AUTH_VIEWER_ROLE=
AUTH_OPERATOR_ROLE=
AUTH_ADMIN_ROLE=
```

</details>

Restart the service:

```bash
docker compose up -d --force-recreate api
```

Verify:

```bash
curl -fsS http://localhost:8000/healthz >/dev/null
```

Expected:

- The health check exits 0 after the restart.

## 3) Bootstrap a demo tenant (Postgres)

Insert a tenant row:

--8<-- "_snippets/demo-tenant-insert.snippet"

Expected:

- The `psql` command exits 0.

## 4) Create minimal tenant config and rules (admin API)

Create a small config document:

```bash
cat > /tmp/demo_config.json <<'JSON'
{
  "weights": { "pop": 1.0, "cooc": 0.0, "emb": 0.0 },
  "flags": { "enable_rules": true },
  "limits": { "max_k": 50, "max_exclude_ids": 200 }
}
JSON
```

Upsert config:

```bash
curl -fsS -X PUT http://localhost:8000/v1/admin/tenants/demo/config \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d @/tmp/demo_config.json
```

Create a small rules document (pin `item_3` to prove control works):

```bash
cat > /tmp/demo_rules.json <<'JSON'
[
  {
    "action": "pin",
    "target_type": "item",
    "item_ids": ["item_3"],
    "surface": "home",
    "priority": 10
  }
]
JSON
```

Upsert rules:

```bash
curl -fsS -X PUT http://localhost:8000/v1/admin/tenants/demo/rules \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d @/tmp/demo_rules.json
```

Expected:

- Both admin `PUT` calls exit 0.

## 5) Seed minimal DB-only signals (tags + popularity)

Seed `item_tags` and `item_popularity_daily` for surface `home`:

```bash
docker exec -i recsys-db psql -U recsys-db -d recsys-db <<'SQL'
with t as (
  select id as tenant_id
    from tenants
   where external_id = 'demo'
)
insert into item_tags (tenant_id, namespace, item_id, tags, price, created_at)
select tenant_id, 'home', 'item_1', array['brand:nike','category:shoes'], 99.90, now() from t
union all
select tenant_id, 'home', 'item_2', array['brand:nike','category:shoes'], 79.00, now() from t
union all
select tenant_id, 'home', 'item_3', array['brand:acme','category:socks'], 12.00, now() from t
on conflict (tenant_id, namespace, item_id)
do update set tags = excluded.tags,
              price = excluded.price,
              created_at = excluded.created_at;

with t as (
  select id as tenant_id
    from tenants
   where external_id = 'demo'
)
insert into item_popularity_daily (tenant_id, namespace, item_id, day, score)
select tenant_id, 'home', 'item_1', current_date, 10 from t
union all
select tenant_id, 'home', 'item_2', current_date, 7 from t
union all
select tenant_id, 'home', 'item_3', current_date, 3 from t
on conflict (tenant_id, namespace, item_id, day)
do update set score = excluded.score;
SQL
```

Expected:

- The `psql` command exits 0.

## 6) Call `/v1/recommend` and verify non-empty output

Send a request with deterministic `request_id`:

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: req-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1","session_id":"s_1"}}'
```

You should see `items` with `item_id` values like `item_1`, `item_2`, `item_3`.

Because you pinned `item_3`, it should appear first in the list.

Example response shape:

```json
{
  "items": [{ "item_id": "item_3", "rank": 1, "score": 0.12 }],
  "meta": {
    "tenant_id": "demo",
    "surface": "home",
    "config_version": "W/\"...\"",
    "rules_version": "W/\"...\"",
    "request_id": "req-1"
  },
  "warnings": []
}
```

If you get an empty list, check:

- you inserted rows into `item_popularity_daily` for `namespace='home'`
- you are calling the API with `surface=home`

Expected:

- The response has a non-empty `items` list.
- `item_3` appears first (pinned rule).

## 7) Extract the exposure log and create a tiny outcome log

Copy the exposure file out of the container:

```bash
docker compose cp api:/app/tmp/exposures.eval.jsonl /tmp/exposures.jsonl
```

Extract the hashed `user_id` from the exposure file (this is what `recsys-service` logs for eval format):

```bash
EXPOSURE_USER_ID="$(
  python3 -c 'import json; print(json.loads(open("/tmp/exposures.jsonl").readline())["user_id"])'
)"
```

Create a minimal outcome log that joins by `request_id` (and matches the exposure `user_id`):

```bash
OUTCOME_TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
cat > /tmp/outcomes.jsonl <<JSONL
{"request_id":"req-1","user_id":"${EXPOSURE_USER_ID}","item_id":"item_3","event_type":"click","ts":"${OUTCOME_TS}"}
JSONL
```

Expected:

- `/tmp/exposures.jsonl` and `/tmp/outcomes.jsonl` both exist and are non-empty.

## 8) Run `recsys-eval` on the logs

Create a dataset config:

```bash
cat > /tmp/dataset.yaml <<'YAML'
exposures:
  type: jsonl
  path: /tmp/exposures.jsonl
outcomes:
  type: jsonl
  path: /tmp/outcomes.jsonl
YAML
```

Create a minimal offline config (slice keys match the service `eval_v1` context keys):

```bash
cat > /tmp/eval.yaml <<'YAML'
mode: offline
offline:
  metrics:
    - name: hitrate
      k: 5
    - name: precision
      k: 5
  slice_keys: ["tenant_id", "surface"]
  gates: []
scale:
  mode: memory
YAML
```

Build + run:

```bash
(cd recsys-eval && make build)

recsys-eval/bin/recsys-eval validate --schema exposure.v1 --input /tmp/exposures.jsonl
recsys-eval/bin/recsys-eval validate --schema outcome.v1 --input /tmp/outcomes.jsonl

recsys-eval/bin/recsys-eval run \
  --mode offline \
  --dataset /tmp/dataset.yaml \
  --config /tmp/eval.yaml \
  --output /tmp/recsys_eval_report.md \
  --output-format markdown
```

Inspect the report:

```bash
sed -n '1,80p' /tmp/recsys_eval_report.md
```

You should see an “Offline Metrics” table with values like:

```text
| hitrate@5 | 1.000000 |
| precision@5 | 0.333333 |
```

Expected:

- Both `recsys-eval ... validate ...` commands exit 0.
- `/tmp/recsys_eval_report.md` exists and is non-empty.

## 9) (Optional) Run pipelines once (produces a manifest)

This step proves `recsys-pipelines` can produce artifacts and a manifest from events.

Ensure MinIO is up:

```bash
curl -fsS http://localhost:9000/minio/health/ready >/dev/null
```

Build + run one day from the tiny pipelines dataset:

```bash
(cd recsys-pipelines && make build)

(cd recsys-pipelines && ./bin/recsys-pipelines run \
  --config configs/env/local.json \
  --tenant demo \
  --surface home \
  --start 2026-01-01 \
  --end 2026-01-01)
```

Verify the local manifest exists:

```bash
cat recsys-pipelines/.out/registry/current/demo/home/manifest.json
```

Verify artifacts exist in MinIO (paths are under the `recsys/` prefix by default):

```bash
docker compose run --rm --entrypoint sh minio-init -c \
  'mc alias set local http://minio:9000 minioadmin minioadmin >/dev/null && \
   mc ls local/recsys-artifacts/recsys/demo/home/ | head'
```

## Appendix: success criteria and troubleshooting

### Success criteria (quick checks)

- Service is healthy: `curl -fsS http://localhost:8000/readyz >/dev/null`
- Tenant exists:

  ```bash
  docker exec -i recsys-db psql -U recsys-db -d recsys-db -c \"select external_id from tenants;\"
  ```

- Config and rules exist:

  ```bash
  curl -fsS http://localhost:8000/v1/admin/tenants/demo/config \\
    -H 'X-Dev-User-Id: dev-user-1' -H 'X-Dev-Org-Id: demo' -H 'X-Org-Id: demo'
  curl -fsS http://localhost:8000/v1/admin/tenants/demo/rules \\
    -H 'X-Dev-User-Id: dev-user-1' -H 'X-Dev-Org-Id: demo' -H 'X-Org-Id: demo'
  ```

- Exposure log exists: `test -s /tmp/exposures.jsonl`
- Eval report exists: `test -s /tmp/recsys_eval_report.md`

### Common failures

- `401/403` from admin or recommend endpoints
  - Check you set `AUTH_*_ROLE=` empty in `api/.env` and recreated the `api` container.
  - Ensure you send both `X-Dev-Org-Id` and `X-Org-Id` headers.
- Empty recommendations
  - Check `item_popularity_daily` has rows for `namespace='home'` and for `day=current_date`.
- Pipelines cannot connect to MinIO
  - Ensure `curl -fsS http://localhost:9000/minio/health/ready` succeeds.

### Runbooks

- Service not ready: [`operations/runbooks/service-not-ready.md`](../operations/runbooks/service-not-ready.md)
- Empty recs: [`operations/runbooks/empty-recs.md`](../operations/runbooks/empty-recs.md)
- Database migration issues: [Database migration issues](../operations/runbooks/db-migration-issues.md)

## Read next

- First surface end-to-end: [`how-to/first-surface-end-to-end.md`](../how-to/first-surface-end-to-end.md)
- Minimum instrumentation spec: [`reference/minimum-instrumentation.md`](../reference/minimum-instrumentation.md)
- Exposure logging & attribution: [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)
