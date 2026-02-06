---
tags:
  - tutorial
  - ops
  - artifacts
  - recsys-service
  - recsys-pipelines
---

# Tutorial: production-like run (pipelines → object store → ship/rollback)

## Who this is for

- Developers/operators who want to practice the artifact/manifest deployment mechanics locally
- Anyone who wants a concrete “ship/rollback by pointer swap” workflow before production

## What you will get

- A local MinIO + manifest setup that `recsys-service` can read
- A reproducible “good ship” and “bad ship” sequence (and a rollback)
- A way to confirm the service is actually reading manifests (not silently falling back)

Run `recsys-pipelines` to publish artifacts to local MinIO, configure `recsys-service` to read an artifact manifest,
then practice **ship + rollback** by updating the manifest pointer.

This tutorial uses the small built-in pipelines dataset and focuses on the deployment mechanics (not model quality).

## Prereqs

- Docker + Docker Compose (v2)
- `make`
- `curl`
- POSIX shell
- Go toolchain (to build `recsys-pipelines`)
- `jq` (optional; only used to print item counts)

Verify you have them:

```bash
docker compose version
make --version
curl --version
go version
```

## Verify (expected outcome)

- `recsys-service` loads the manifest from object storage (log line: `artifact manifest loaded`)
- `POST /v1/recommend` returns non-empty results after a “good” ship
- `POST /v1/recommend` returns an empty list after a “bad” ship
- rolling back the manifest restores non-empty results

## 1) Start Postgres, MinIO, and recsys-service

From repo root:

```bash
make dev
```

Verify:

```bash
curl -fsS http://localhost:8000/healthz >/dev/null
curl -fsS http://localhost:9000/minio/health/ready >/dev/null
```

Expected:

- Both health checks exit 0.

## 2) Enable artifact/manifest mode in `recsys-service`

Edit `api/.env` (create it from `api/.env.example` if missing) and set:

<details markdown="1" open>
<summary>Tutorial env settings (copy/paste)</summary>

```bash
RECSYS_ARTIFACT_MODE_ENABLED=true
RECSYS_ARTIFACT_MANIFEST_TEMPLATE=s3://recsys-artifacts/registry/current/{tenant}/{surface}/manifest.json

# Tutorial convenience: reload quickly when we swap the manifest
RECSYS_ARTIFACT_MANIFEST_TTL=1s
RECSYS_ARTIFACT_CACHE_TTL=1s

# Deterministic behavior for the demo
RECSYS_ALGO_MODE=popularity
```

</details>

Apply env changes (Compose loads `env_file` values at container creation time):

```bash
docker compose up -d --force-recreate api
```

Verify:

```bash
curl -fsS http://localhost:8000/healthz >/dev/null
```

Expected:

- The service restarts and the health check exits 0.

## 3) Run pipelines to publish a “good” artifact set

Build the CLI:

```bash
(cd recsys-pipelines && make build)
```

Run one day (this produces non-empty popularity/co-vis artifacts from the tiny dataset):

```bash
(cd recsys-pipelines && ./bin/recsys-pipelines run \
  --config configs/env/local.json \
  --tenant demo \
  --surface home \
  --start 2026-01-01 \
  --end 2026-01-01)
```

This writes a local manifest at:

- `recsys-pipelines/.out/registry/current/demo/home/manifest.json`

And uploads referenced blobs to the `recsys-artifacts` bucket in MinIO.

Expected:

- `recsys-pipelines/.out/registry/current/demo/home/manifest.json` exists and is non-empty.

## 4) Publish the manifest pointer to MinIO (so the service can load it)

Today, `recsys-pipelines` writes the manifest locally. In a production setup, you would publish the manifest pointer
as part of your pipeline/CI run.

For local dev, copy the manifest file into MinIO using the `minio-init` image (it includes `mc`):

```bash
docker compose run --rm --entrypoint sh \
  -v "$PWD/recsys-pipelines/.out/registry/current/demo/home/manifest.json:/tmp/manifest.json:ro" \
  minio-init -c \
  'mc alias set local http://minio:9000 minioadmin minioadmin >/dev/null && \
   mc cp /tmp/manifest.json local/recsys-artifacts/registry/current/demo/home/manifest.json'
```

Expected:

- The `mc cp ...` command exits 0.

## 5) Call `/v1/recommend` and verify non-empty output

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: prodlike-1' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1","session_id":"s_1"}}'
```

Optional quick check:

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: prodlike-1b' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1","session_id":"s_1"}}' | jq '.items | length'
```

You should see a non-empty `items` array with `item_id` values.

Expected:

- `POST /v1/recommend` returns a non-empty `items` list.

## 6) Ship a “bad” manifest (empty window) and observe the impact

Back up the current (good) manifest locally:

```bash
cp recsys-pipelines/.out/registry/current/demo/home/manifest.json /tmp/manifest-good.json
```

Generate a manifest for a day with **no events** in the tiny dataset (this produces empty artifacts):

```bash
(cd recsys-pipelines && ./bin/recsys-pipelines run \
  --config configs/env/local.json \
  --tenant demo \
  --surface home \
  --start 2026-01-02 \
  --end 2026-01-02)
```

Publish the new (bad) manifest:

```bash
docker compose run --rm --entrypoint sh \
  -v "$PWD/recsys-pipelines/.out/registry/current/demo/home/manifest.json:/tmp/manifest.json:ro" \
  minio-init -c \
  'mc alias set local http://minio:9000 minioadmin minioadmin >/dev/null && \
   mc cp /tmp/manifest.json local/recsys-artifacts/registry/current/demo/home/manifest.json'
```

Wait for the manifest TTL to expire (we set it to 1s):

```bash
sleep 2
```

Call recommend again; you should now get an empty list:

```bash
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: prodlike-2' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1","session_id":"s_1"}}' | jq '.items | length'
```

Expected:

- The item count is `0` after shipping the bad manifest.

## 7) Roll back the manifest and verify recovery

Publish the backed-up manifest back to the “current” pointer:

```bash
docker compose run --rm --entrypoint sh \
  -v "/tmp/manifest-good.json:/tmp/manifest.json:ro" \
  minio-init -c \
  'mc alias set local http://minio:9000 minioadmin minioadmin >/dev/null && \
   mc cp /tmp/manifest.json local/recsys-artifacts/registry/current/demo/home/manifest.json'
```

Wait for TTL and retry:

```bash
sleep 2
curl -fsS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: prodlike-3' \
  -H 'X-Dev-User-Id: dev-user-1' \
  -H 'X-Dev-Org-Id: demo' \
  -H 'X-Org-Id: demo' \
  -d '{"surface":"home","k":5,"user":{"user_id":"u_1","session_id":"s_1"}}' | jq '.items | length'
```

You should be back to a non-zero item count.

Expected:

- The item count is non-zero after rollback.

## 8) Confirm the service is actually reading manifests

The service logs a line when it loads a manifest:

```bash
docker compose logs --tail 200 api | grep -i "artifact manifest loaded"
```

If you do not see it, confirm the environment inside the container:

```bash
docker compose exec -T api sh -c 'printenv | grep -E "RECSYS_ARTIFACT_MODE_ENABLED|RECSYS_ARTIFACT_MANIFEST_TEMPLATE"'
```

Expected:

- The logs include `artifact manifest loaded` after you publish/swapped the manifest.

## Troubleshooting (runbooks)

- Service not ready: [`operations/runbooks/service-not-ready.md`](../operations/runbooks/service-not-ready.md)
- Empty recs: [`operations/runbooks/empty-recs.md`](../operations/runbooks/empty-recs.md)
- Stale manifest: [`operations/runbooks/stale-manifest.md`](../operations/runbooks/stale-manifest.md)

## Read next

- Operate pipelines: [`how-to/operate-pipelines.md`](../how-to/operate-pipelines.md)
- Run evaluation and make ship decisions: [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)
- Deploy with Helm: [`how-to/deploy-helm.md`](../how-to/deploy-helm.md)
