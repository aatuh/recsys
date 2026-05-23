# Contributing

Contribution workflow and documentation maintenance rules are documented in [docs/contributing.md](docs/contributing.md).

Before handing work off, run the closest relevant quality gate:

```bash
make docs-check
make finalize
```

If a broad gate is blocked by local infrastructure, report the blocker and the narrower checks that passed.
