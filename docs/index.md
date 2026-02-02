# RecSys suite documentation

Welcome. This documentation is intended to start running the suite, integrate it
into an application, and operate it safely.

## What is the suite?

The suite is four modules that form an end-to-end recommendation system loop:

- **recsys-service**: low-latency recommendation API (auth, tenancy, limits,

  caching, observability, exposure logging).

- **recsys-algo**: deterministic ranking logic (candidate merge, scoring,

  constraints, rules, diversity).

- **recsys-pipelines**: offline/stream processing that turns events into

  versioned artifacts the service consumes.

- **recsys-eval**: offline regression + online experiment analysis that decides

  what to ship.

## Where to start

1. Tutorial: [`tutorials/local-end-to-end.md`](tutorials/local-end-to-end.md)
2. Integrate: [`how-to/integrate-recsys-service.md`](how-to/integrate-recsys-service.md)
3. Operate: [`how-to/operate-pipelines.md`](how-to/operate-pipelines.md)
4. Evaluate: [`how-to/run-eval-and-ship.md`](how-to/run-eval-and-ship.md)
5. Deploy: [`how-to/deploy-helm.md`](how-to/deploy-helm.md)

## Reference

- REST API: [`reference/api/openapi.yaml`](reference/api/openapi.yaml)
- Admin API: [`reference/api/admin.md`](reference/api/admin.md)
- Contracts: [`reference/data-contracts/`](reference/data-contracts/)
- Config: [`reference/config/`](reference/config/)
- CLI: [`reference/cli/`](reference/cli/)
- Database: [`reference/database/`](reference/database/)

Note:

- Admin endpoints are documented in [`reference/api/admin.md`](reference/api/admin.md)

## Concepts

- [`explanation/suite-architecture.md`](explanation/suite-architecture.md)
- [`explanation/candidate-vs-ranking.md`](explanation/candidate-vs-ranking.md)
- [`explanation/exposure-logging-and-attribution.md`](explanation/exposure-logging-and-attribution.md)
- [`explanation/surface-namespaces.md`](explanation/surface-namespaces.md)
- [`explanation/data-modes.md`](explanation/data-modes.md)

## Development

- Use the repo-level Go workspace: `go work sync` from `recsys/`.
- Each module is versioned/released independently; tags are module-prefixed

  (e.g., `recsys-eval/v0.2.0`). Run tests per module (e.g., `cd recsys-eval && go test ./...`).

## Operations

- [`operations/runbooks/service-not-ready.md`](operations/runbooks/service-not-ready.md)
- [`operations/runbooks/empty-recs.md`](operations/runbooks/empty-recs.md)
- [`operations/runbooks/rollback-config-rules.md`](operations/runbooks/rollback-config-rules.md)
- [`operations/performance-and-capacity.md`](operations/performance-and-capacity.md)

## Contributing

- [`contributing/docs-style.md`](contributing/docs-style.md)
