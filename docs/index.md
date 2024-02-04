# RecSys suite documentation

Welcome. This documentation is intended to start running the suite, integrate it
into an application, and operate it safely.

You can access this documentation in several ways:

- Visit [`https://recsys.app`](https://recsys.app).
- Visit [`https://github.com/aatuh/recsys`](https://github.com/aatuh/recsys).
- Run `make docs-serve` and open [`http://localhost:8001`](http://localhost:8001).
- Browse the files in the repository under [`/docs`](/docs).

Browse the source repository on GitHub: [`github.com/aatuh/recsys`](https://github.com/aatuh/recsys).

## Need guided help?

- RecSys Copilot for docs/tech/business Q&A: [`chatgpt.com/g/.../recsys-copilot`](https://chatgpt.com/g/g-68c82a5c7704819185d0ff929b6fff11-recsys-copilot).

## What is the suite?

The suite is four modules that form an end-to-end recommendation system loop:

- **recsys-service**: low-latency recommendation API (auth, tenancy, limits, caching, observability, exposure logging).
- **recsys-algo**: deterministic ranking logic (candidate merge, scoring, constraints, rules, diversity).
- **recsys-pipelines**: offline/stream processing that turns events into versioned artifacts the service consumes.
- **recsys-eval**: offline regression + online experiment analysis that decides what to ship.

If you're evaluating this as a product (rather than integrating it right now), start with:

- Stakeholder overview: [`start-here/what-is-recsys.md`](start-here/what-is-recsys.md)

## Where to start

1. Tutorial: [`tutorials/local-end-to-end.md`](tutorials/local-end-to-end.md)
2. Integrate: [`how-to/integrate-recsys-service.md`](how-to/integrate-recsys-service.md)
3. Operate: [`how-to/operate-pipelines.md`](how-to/operate-pipelines.md)
4. Evaluate: [`how-to/run-eval-and-ship.md`](how-to/run-eval-and-ship.md)
5. Deploy: [`how-to/deploy-helm.md`](how-to/deploy-helm.md)

## Reference

- REST API: [`reference/api/openapi.yaml`](reference/api/openapi.yaml)
- Admin API: [`reference/api/admin.md`](reference/api/admin.md)
- Contracts: [`reference/data-contracts/index.md`](reference/data-contracts/index.md)
- Config: [`reference/config/index.md`](reference/config/index.md)
- CLI: [`reference/cli/index.md`](reference/cli/index.md)
- Database: [`reference/database/index.md`](reference/database/index.md)

Note:

- Admin endpoints are documented in [`reference/api/admin.md`](reference/api/admin.md)

## Concepts

- [`explanation/suite-architecture.md`](explanation/suite-architecture.md)
- [`explanation/candidate-vs-ranking.md`](explanation/candidate-vs-ranking.md)
- [`explanation/exposure-logging-and-attribution.md`](explanation/exposure-logging-and-attribution.md)
- [`explanation/surface-namespaces.md`](explanation/surface-namespaces.md)
- [`explanation/data-modes.md`](explanation/data-modes.md)

## Development

- Run tests per module (e.g., `cd recsys-eval && go test ./...`).
- Each module is versioned/released independently; tags are module-prefixed (e.g., `recsys-eval/v0.2.0`).

## Operations

- [`operations/runbooks/service-not-ready.md`](operations/runbooks/service-not-ready.md)
- [`operations/runbooks/empty-recs.md`](operations/runbooks/empty-recs.md)
- [`operations/runbooks/rollback-config-rules.md`](operations/runbooks/rollback-config-rules.md)
- [`operations/performance-and-capacity.md`](operations/performance-and-capacity.md)

## Contributing

- [`project/docs-style.md`](project/docs-style.md)
