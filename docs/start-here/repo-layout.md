---
diataxis: reference
tags:
  - reference
  - project
  - developer
---
# Repo layout and Go module paths

This repository is hosted at `github.com/aatuh/recsys`, but the Go module import paths currently use the
`github.com/aatuh/recsys-suite/...` prefix. This page explains how to navigate the repo and how to use the modules.

## Who this is for

- Engineers reading the codebase for the first time
- Integrators who want to build/bump module versions independently
- Anyone confused by repo name vs Go module import paths

## What you will get

- Where each module lives in the repo
- The Go import paths to use in each module
- The release/tagging convention used by the suite

## Repo layout (what lives where)

- `api/`: `recsys-service` (the online HTTP API)
- `recsys-algo/`: `recsys-algo` (the deterministic ranking core)
- `recsys-pipelines/`: `recsys-pipelines` (offline pipelines that build artifacts/signals)
- `recsys-eval/`: `recsys-eval` (evaluation tooling and report generation)

Each module is a standalone Go module with its own `go.mod`, tests, and `Makefile`.

## Go module paths (what you `go get`)

- `recsys-service` module: `github.com/aatuh/recsys-suite/api`
- `recsys-algo` module: `github.com/aatuh/recsys-suite/api/recsys-algo`
- `recsys-pipelines` module: `github.com/aatuh/recsys-suite/recsys-pipelines`
- `recsys-eval` module: `github.com/aatuh/recsys-suite/recsys-eval`

## Versioning and tags

Each module is versioned independently. Tags are module-prefixed, for example:

- `recsys-eval/v0.2.0`
- `recsys-pipelines/v0.2.0`
- `recsys-algo/v0.2.0`

## Developing locally

The recommended workflow is:

- run the full suite from the repo root with `make test`
- run builds/tests from within each module directory when you only need one component (e.g., `cd recsys-eval && make test`)
- use Docker Compose for the service when you want the full local stack (`make dev`)

Fresh checkouts do not need manual env-file setup for the common flows. `make dev` creates `api/.env` from
`api/.env.example` when missing, and `make test` creates `api/.env.test` from `api/.env.test.example` when missing.
Existing local env files are never overwritten. The same test env bootstrap runs when you use `cd api && make test`.

The `api/` module uses `replace` directives for local development (for example, to use the local `../recsys-algo`).

## Read next

- Suite architecture (what runs where): [Suite architecture](../explanation/suite-architecture.md)
- Local end-to-end tutorial: [local end-to-end (service → logging → eval)](../tutorials/local-end-to-end.md)
