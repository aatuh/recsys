---
diataxis: explanation
tags:
  - recsys-eval
---
# Architecture: how the code is organized and how to extend it
This page explains Architecture: how the code is organized and how to extend it and how it fits into the RecSys suite.


## Who this is for

Maintainers and contributors.

## What you will get

- The boundaries (domain vs ports vs adapters)
- Where to add a new metric, datasource, or report writer
- How to avoid creating a god-package

## High-level structure

- cmd/:

  CLI entrypoints

- internal/domain/:

  pure logic: metrics, statistics, joining rules, report models

- internal/ports/:

  interfaces for IO: datasources, report writers, loggers

- internal/adapters/:

  concrete IO: JSONL readers, Postgres readers, writers

- internal/app/:

  usecases that orchestrate domain logic + ports

If you keep domain pure, tests become easy and reliability improves.

## Add a new metric

1) Implement the metric in internal/domain/metrics/...
2) Add it to the registry (internal/domain/metrics/registry.go)
3) Add tests with toy inputs and known outputs
4) Document it in [Metrics: what we measure and why](metrics.md)

## Add a new datasource

1) Implement ports interfaces (ExposureReader, OutcomeReader, etc.)
2) Add adapter under internal/adapters/datasource/`yourtype`/
3) Wire it into the datasource factory or provider registry (depending on repo)

## Add a new report format

1) Implement a writer adapter under internal/adapters/reporting/
2) Ensure the JSON report stays canonical (other formats derive from it)

## The rule of thumb

- Domain code should not import adapters.
- Ports should not import adapters.
- Adapters can import ports and domain.

This keeps the system testable and change-friendly.

## Read next

- Workflows (practical entry points): [recsys-eval docs](index.md)
- Offline gate in CI: [Workflow: Offline gate in CI](workflows/offline-gate-in-ci.md)
- Integration logging plan: [Integration: how to produce the inputs](integration.md)
- Concepts: [Concepts: how to understand recsys-eval](concepts.md)
