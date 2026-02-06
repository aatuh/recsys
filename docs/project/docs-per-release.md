---
tags:
  - project
  - docs
---

# Docs per release policy

## Who this is for

- Maintainers preparing a release
- Reviewers validating that a change is “customer-ready”

## What you will get

- A checklist that turns doc updates into a release habit (not an afterthought)
- Clear “when X changes, update Y” guidance for the suite docs site

## Policy

- Doc changes ship with product changes.
- The canonical API spec lives in [`reference/api/openapi.yaml`](../reference/api/openapi.yaml).
- Tutorials must remain runnable (see the tutorial smoke test workflow).

## Checklists (by change type)

### HTTP API changes

- Update `docs/reference/api/openapi.yaml`.
- Regenerate derived API artifacts: `cd api && make codegen`.
- Update examples and troubleshooting pages if behavior changed:
  - [`reference/api/examples.md`](../reference/api/examples.md)
  - [`reference/api/errors.md`](../reference/api/errors.md)
- Add an entry to “What’s new” for customer-visible changes:
  - [`whats-new/index.md`](../whats-new/index.md)

### Config changes

- Update the config reference:
  - [`reference/config/index.md`](../reference/config/index.md)
  - module-specific pages under `reference/config/`
- If a setting changes defaults or semantics, add a note to “What’s new”.

### Data contract changes (exposures/outcomes/manifests)

- Update:
  - [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
  - schema files under `reference/data-contracts/` (and examples)
- Update join/explanation docs if attribution logic changes:
  - [`reference/data-contracts/join-logic.md`](../reference/data-contracts/join-logic.md)
  - [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)

### Operational behavior changes

- Update runbooks for new failure modes or changes in remediation:
  - `docs/operations/runbooks/*`
- Update capacity guidance if perf characteristics changed:
  - [`operations/performance-and-capacity.md`](../operations/performance-and-capacity.md)

## Verification (required)

Run the suite-level quality gates:

```bash
make finalize
```

If you changed tutorials or serving behavior, also run the tutorial smoke test locally (or wait for CI):

```bash
bash scripts/tutorial_smoke_test.sh
```

## Read next

- Docs versioning: [`project/docs-versioning.md`](docs-versioning.md)
