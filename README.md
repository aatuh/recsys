# RecSys suite

RecSys is an auditable recommendation system suite with deterministic ranking and versioned ship/rollback.

This repository contains a complete, production-oriented recommendation system stack:

- **recsys-service** (`api/`): low-latency HTTP API for serving recommendations
- **recsys-algo** (`recsys-algo/`): deterministic ranking core used by the service
- **recsys-pipelines** (`recsys-pipelines/`): offline pipelines that build versioned artifacts
- **recsys-eval** (`recsys-eval/`): evaluation tooling for regression + experimentation

## Documentation

All the hosted documentation lives under [`/docs`](/docs), rendered with MkDocs using [`mkdocs.yml`](mkdocs.yml).

You can access this documentation in several ways:

- Visit [`https://recsys.app`](https://recsys.app).
- Visit [`https://github.com/aatuh/recsys`](https://github.com/aatuh/recsys) and browse the `/docs` directory.
- Run `make docs-serve` and open [`http://localhost:8001`](http://localhost:8001).

Start here:

- [`docs/index.md`](docs/index.md)
- Tutorial: [`docs/tutorials/local-end-to-end.md`](docs/tutorials/local-end-to-end.md)
- Suite architecture: [`docs/explanation/suite-architecture.md`](docs/explanation/suite-architecture.md)

Module docs in MkDocs:

- recsys-algo: [`docs/recsys-algo/index.md`](docs/recsys-algo/index.md)
- recsys-pipelines: [`docs/recsys-pipelines/docs/index.md`](docs/recsys-pipelines/docs/index.md)
- recsys-eval: [`docs/recsys-eval/docs/index.md`](docs/recsys-eval/docs/index.md)

## Docs versioning

- Tags `recsys-suite/vX.Y.Z` publish `/X.Y.Z/` and update `/latest/`.
- Default branch publishes `/dev/`.
- The site root (`/`) serves the contents of `/latest/` (falls back to `/dev/` before the first tag; no URL change).
- GitHub Pages source is **GitHub Actions** (Settings → Pages).

## Licensing

See the canonical licensing pages in:

- [`docs/licensing/index.md`](docs/licensing/index.md)
- [`docs/licensing/pricing.md`](docs/licensing/pricing.md)
