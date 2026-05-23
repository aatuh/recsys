# recsys-algo

`recsys-algo` is the deterministic ranking library used by recsys-service. It contains candidate merging, scoring,
personalization, rules, diversity controls, and example programs.

## Common commands

```bash
make test
make build
make plugin-example
```

Expected result: unit tests pass, standard examples build, and the custom algorithm plugin example builds to
`/tmp/recsys-hello-algo.so`.

## Documentation

- Suite architecture: [../docs/architecture.md](../docs/architecture.md)
- Local workflow: [../docs/reference/local-workflow.md](../docs/reference/local-workflow.md)
- API integration context: [../docs/integration.md](../docs/integration.md)

Release tags use the module prefix, for example `recsys-algo/v0.2.0`.
