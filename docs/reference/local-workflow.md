# Local Workflow Reference

## Who this is for

Maintainers and contributors running repository quality gates or debugging local development commands.

## What you will get

- The repo-level command map.
- Module-level quality gates.
- Documentation checks and expected high-level outcomes.
- Notes about checks that need Docker or external tools.

## Repository commands

```bash
make help
make docs-check
make test
make finalize
```

Expected result:

- `make help` lists available targets.
- `make docs-check` validates docs links, external links, spelling, and strict MkDocs build.
- `make test` runs the proof-kit smoke test and module tests.
- `make finalize` runs formatting, linting, tests, codegen, Markdown lint, and docs checks.

## Module commands

| Module | Common commands |
| --- | --- |
| `api/` | `make env`, `make test-env`, `make dev`, `make test`, `make codegen`, `make finalize` |
| `recsys-algo/` | `make test`, `make build`, `make plugin-example`, `make finalize` |
| `recsys-pipelines/` | `make test`, `make build`, `make smoke`, `make finalize` |
| `recsys-eval/` | `make test`, `make schema-check`, `make build`, `make finalize` |

## Documentation checks

```bash
make mdlint
python3 scripts/docs_linkcheck.py
python3 scripts/docs_external_linkcheck.py
make docs-build
```

Expected result: Markdown style passes, internal links resolve, external links are reachable or intentionally skipped,
and MkDocs builds into `.site/` with strict mode.

## Generated files

Run this when API spec sources change:

```bash
make codegen
```

Expected result: generated OpenAPI artifacts under `api/docs/` match `docs/reference/api/openapi.yaml`.

## Notes

- Docker is required for the repo-level API test workflow.
- Some module `finalize` targets install tools such as `golangci-lint`, `gosec`, and `govulncheck`.
- `.site/` is generated output. Do not edit it as source documentation.
