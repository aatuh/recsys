# Contributing

This page documents the contribution workflow for code and documentation changes in this repository.

## Local workflow

1. Read the root `README.md` and the relevant module docs.
2. Make the smallest reviewable change.
3. Update docs when behavior, commands, config, API, or operational expectations change.
4. Run the closest quality gate while developing.
5. Run the repository-standard final gate before handing work off when feasible.

## Quality gates

```bash
make fmt
make lint
make test
make docs-check
make finalize
```

Expected result: formatting, linting, tests, docs validation, codegen, and strict docs build all pass. If a broad gate
is blocked by missing local infrastructure, run the closest module-level equivalent and document the blocker.

## Documentation rules

- `docs/` is the canonical MkDocs source.
- Keep pages task-oriented and lean.
- Prefer links to code, schemas, OpenAPI, and Makefiles over duplicating long reference data.
- Do not edit `.site/` as source.
- Do not restore `.docs/` or `.trash/docs/` wholesale.
- Use `.trash/docs` only for pricing, licensing schema, and contact details when rebuilding commercial pages.

## Commit style

Use Conventional Commits, for example:

```text
docs: restart canonical documentation set
fix: harden docs link checker
```

## Code of conduct and governance

Use normal professional conduct: be specific, respectful, evidence-based, and focused on reproducible issues. Project
governance is maintainer-led unless a future governance document states a different model.
