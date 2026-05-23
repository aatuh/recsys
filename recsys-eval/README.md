# recsys-eval

`recsys-eval` is the Apache-2.0 evaluation CLI for recommendation systems. It supports offline regression gates,
experiment analysis, OPE, interleaving, schema validation, and decision reports.

## Common commands

```bash
make test
make schema-check
make build
```

Expected result: evaluation tests pass, report schemas validate, and `bin/recsys-eval` is built.

## Documentation

- Integration and evaluation path: [../docs/integration.md](../docs/integration.md)
- Data contracts: [../docs/reference/data-contracts.md](../docs/reference/data-contracts.md)
- Local workflow: [../docs/reference/local-workflow.md](../docs/reference/local-workflow.md)
- Licensing: [../docs/commercial/licensing.md](../docs/commercial/licensing.md)

Release tags use the module prefix, for example `recsys-eval/v0.2.0`.
