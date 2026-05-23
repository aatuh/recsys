# RecSys Documentation

RecSys is a self-hosted recommendation system suite for teams that need deterministic serving, versioned artifact
rollout, offline evaluation, and auditable operational decisions.

This restarted documentation set is intentionally small. It documents the current repository surface first and links to
source files for details that are safer to maintain in code.

## Choose a path

| Reader | Start here | Outcome |
| --- | --- | --- |
| Developer integrating the service | [Developer quickstart](developer-quickstart.md) | Run the local stack and make one recommendation request. |
| Developer proving first success | [Local end-to-end](local-end-to-end.md) | Run the smoke-tested tenant bootstrap, rules, recommendation, and exposure-log path. |
| Product or data team validating quality | [Integration and evaluation](integration.md) and [Evaluation decisions](evaluation-decisions.md) | Understand request IDs, exposure/outcome logging, and ship/hold/rollback gates. |
| Operator preparing a rollout | [Operations](operations.md) and [Artifacts and pipelines](artifacts-and-pipelines.md) | Know health checks, rollback levers, artifact freshness, and first triage steps. |
| Reviewer checking licensing or procurement | [Licensing](commercial/licensing.md), [Pricing](commercial/pricing.md), and [Procurement](commercial/procurement.md) | See the license model, commercial plan definitions, and review packet. |

## What is in this repository

| Path | Responsibility |
| --- | --- |
| `api/` | HTTP recommendation service, admin routes, OpenAPI generation, migrations, auth, exposure logging, and license status. |
| `recsys-algo/` | Deterministic ranking library with blending, personalization, rules, diversity controls, and plugin examples. |
| `recsys-pipelines/` | Batch jobs that build and publish versioned recommendation artifacts. |
| `recsys-eval/` | Apache-2.0 evaluation CLI for offline gates, experiment analysis, OPE, interleaving, and report schemas. |
| `examples/` | Small datasets and fixtures used by tutorials, tests, and demos. |

## Canonical docs contract

- `docs/` is the only source directory for the MkDocs site.
- `.site/` is generated output and is not source documentation.
- `.docs/` and `.trash/docs/` are historical workspace material. Do not restore them wholesale.
- Pricing, licensing schema, and contact details may be recovered from `.trash/docs` when they are verified against
  current repository files.

## Local documentation commands

```bash
make docs-build
make docs-check
make docs-serve
```

Expected result: `make docs-build` produces `.site/documentation/technical/`, `make docs-check` passes internal links,
external links, codespell, and strict MkDocs build, and `make docs-serve` serves the MkDocs site at
`http://localhost:8001`.

## Core pages

- [Architecture](architecture.md)
- [Artifacts and pipelines](artifacts-and-pipelines.md)
- [Local end-to-end](local-end-to-end.md)
- [Integration and evaluation](integration.md)
- [Evaluation decisions](evaluation-decisions.md)
- [Operations](operations.md)
- [API reference](reference/api.md)
- [Configuration reference](reference/config.md)
- [Data contracts](reference/data-contracts.md)
- [Glossary](reference/glossary.md)
- [Local workflow](reference/local-workflow.md)
- [Security](security.md)
- [Support](support.md)
- [Contributing](contributing.md)
