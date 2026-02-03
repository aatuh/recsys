# RecSys suite

This repository contains a complete, production-oriented recommendation system stack:

- **recsys-service** (`api/`): low-latency HTTP API for serving recommendations
- **recsys-algo** (`recsys-algo/`): deterministic ranking core used by the service
- **recsys-pipelines** (`recsys-pipelines/`): offline pipelines that build versioned artifacts
- **recsys-eval** (`recsys-eval/`): evaluation tooling for regression + experimentation

## Documentation

Visit [`recsys.app`](https://recsys.app) to see hosted documentation.

All the hosted documentation lives under [`/docs`](docs/), rendered with MkDocs using [`mkdocs.yml`](mkdocs.yml).

Start here:

- [`docs/index.md`](docs/index.md)
- Tutorial: [`docs/tutorials/local-end-to-end.md`](docs/tutorials/local-end-to-end.md)
- Suite architecture: [`docs/explanation/suite-architecture.md`](docs/explanation/suite-architecture.md)

Module docs in MkDocs:

- recsys-algo: [`docs/recsys-algo/index.md`](docs/recsys-algo/index.md)
- recsys-pipelines: [`docs/recsys-pipelines/docs/index.md`](docs/recsys-pipelines/docs/index.md)
- recsys-eval: [`docs/recsys-eval/docs/index.md`](docs/recsys-eval/docs/index.md)

## Licensing

See the canonical licensing pages in:

- [`docs/licensing/index.md`](docs/licensing/index.md)
- [`docs/licensing/pricing.md`](docs/licensing/pricing.md)
