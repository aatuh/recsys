# Start here

This site documents the **RecSys suite**: a production-ready set of modules for building, shipping, and operating a
recommendation system.

## Recommended path

1. **Run the suite locally end-to-end**
   - Tutorial: [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)

2. **Understand the architecture and data flow**
   - Diagram: [`start-here/diagrams/suite-context.md`](diagrams/suite-context.md)
   - Explanation: [`explanation/suite-architecture.md`](../explanation/suite-architecture.md)

3. **Integrate, operate, and validate**
   - Integrate the API: [`how-to/integrate-recsys-service.md`](../how-to/integrate-recsys-service.md)
   - Operate pipelines: [`how-to/operate-pipelines.md`](../how-to/operate-pipelines.md)
   - Run evaluation and ship decisions: [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)

## Role-based entry points

### Lead developer / platform engineer

- Suite architecture and contracts:
  - [`explanation/suite-architecture.md`](../explanation/suite-architecture.md)
  - [`reference/data-contracts/`](../reference/data-contracts/)
  - [`reference/config/`](../reference/config/)
  - [`reference/api/openapi.yaml`](../reference/api/openapi.yaml)

### Recommendation engineer / ML engineer

- Ranking core:
  - [`recsys-algo/`](../recsys-algo/)
- Pipelines outputs:
  - [`recsys-pipelines/docs/reference/output-layout.md`](../recsys-pipelines/docs/reference/output-layout.md)

### Product / business stakeholder

- What to expect from evaluation and decisions:
  - [`recsys-eval/docs/interpreting_results.md`](../recsys-eval/docs/interpreting_results.md)
  - [`recsys-eval/docs/metrics.md`](../recsys-eval/docs/metrics.md)

### SRE / on-call

- Operational runbooks:
  - [`operations/runbooks/`](../operations/runbooks/)
  - [`recsys-pipelines/docs/operations/runbooks/`](../recsys-pipelines/docs/operations/runbooks/)
