# RecSys Suite

RecSys is an auditable recommendation system suite with deterministic ranking, versioned artifact rollout, and offline
evaluation support.

## Components

| Component | Path | Purpose |
| --- | --- | --- |
| recsys-service | `api/` | HTTP API for recommendations, admin config/rules, auth, exposure logging, and OpenAPI artifacts. |
| recsys-algo | `recsys-algo/` | Deterministic ranking core used by the service and examples. |
| recsys-pipelines | `recsys-pipelines/` | Offline jobs that build and publish versioned artifacts. |
| recsys-eval | `recsys-eval/` | Apache-2.0 evaluation tooling for regression gates, experiments, OPE, and reports. |

## Documentation

Canonical documentation lives in [`docs/`](docs/index.md) and is rendered with [`mkdocs.yml`](mkdocs.yml).

Start here:

- [Developer quickstart](docs/developer-quickstart.md)
- [Architecture](docs/architecture.md)
- [Integration and evaluation](docs/integration.md)
- [Operations](docs/operations.md)
- [API reference](docs/reference/api.md)
- [Configuration reference](docs/reference/config.md)
- [Licensing](docs/commercial/licensing.md)
- [Pricing](docs/commercial/pricing.md)

## Local workflow

```bash
make env
make dev
make docs-check
```

Expected result:

- `make env` creates `api/.env` if it is missing.
- `make dev` starts the Compose development stack.
- `make docs-check` validates links, spelling, and strict MkDocs build.

Run the full repository gate when feasible:

```bash
make finalize
```

`make finalize` runs formatting, linting, tests, code generation, Markdown linting, and docs checks. It requires the
same local tooling and Docker prerequisites as the module test targets.

## Licensing

This is a multi-license repository:

- `recsys-eval/**` is Apache-2.0.
- All other paths are AGPL-3.0-only unless a file or closest directory-level notice states otherwise.

See [Licensing](docs/commercial/licensing.md) and [Pricing](docs/commercial/pricing.md) for commercial terms.
