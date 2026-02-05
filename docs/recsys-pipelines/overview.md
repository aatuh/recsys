# recsys-pipelines

Filesystem-first pipelines that build **versioned recommendation artifacts**
from raw exposure events.

This repository is the **offline factory** of a recommender stack:

- It ingests raw exposure events (JSONL, Postgres, or S3 batch).
- It canonicalizes them into a deterministic, replayable dataset.
- It computes artifacts (v1: popularity, co-occurrence, implicit, content_sim, session_seq).
- It validates outputs and enforces hard resource limits.
- It publishes artifacts to a versioned object store and updates a

  single "current" manifest pointer.

If you are new: start at `docs/start-here.md`.

---

## Quickstart

Requirements:

- Go toolchain (see `go.mod`)

Run the pipeline locally against the tiny sample dataset:

```bash
make test
make build

./bin/recsys-pipelines run \
  --config configs/env/local.json \
  --tenant demo \
  --surface home \
  --start 2026-01-01 \
  --end 2026-01-01
```

Outputs (default `.out/`):

- Canonical events: `.out/canonical/<tenant>/<surface>/exposures/YYYY-MM-DD.jsonl`
- Staged artifacts: `.out/artifacts/<tenant>/<surface>/<segment>/<type>/<window>/...`
- Published blobs: `.out/objectstore/<tenant>/<surface>/<type>/<version>.json`
- Current manifest: `.out/registry/current/<tenant>/<surface>/manifest.json`

Run the smoke test (includes an idempotency check):

```bash
make smoke
```

---

## Documentation

Docs are organized into tutorials, how-to guides, explanations, and reference.
See `docs/index.md` for the entry point.

- Start here: `docs/start-here.md`
- Tutorials: `docs/tutorials/`
- How-to: `docs/how-to/`
- Explanations: `docs/explanation/`
- Reference: `docs/reference/`
- Operations: `docs/operations/`

---

## Binaries

- `recsys-pipelines`: one-shot runner (local/dev, or simple cron)
- `job_ingest`: ingest + canonicalize (job-per-container style)
- `job_popularity`: compute + stage popularity artifact
- `job_cooc`: compute + stage co-occurrence artifact
- `job_implicit`: compute + stage implicit (collaborative) artifact
- `job_content_sim`: compute + stage content similarity artifact
- `job_session_seq`: compute + stage session sequence artifact
- `job_validate`: validate canonical event quality for a window range
- `job_publish`: publish staged artifacts + swap the current manifest
- `job_db_signals`: write popularity + co-vis signals into Postgres
- `job_catalog`: ingest item tags into Postgres

See: `docs/tutorials/job-mode.md`.

---

## Contributing

See `docs/contributing/dev-workflow.md`.

---

## Releases

Tag releases with the module prefix, e.g. `recsys-pipelines/v0.2.0`.
