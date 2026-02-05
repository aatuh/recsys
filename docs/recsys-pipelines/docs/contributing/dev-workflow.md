
# Developer workflow

## Local commands

```bash
make fmt
make test
make build
make smoke
```

## Code structure rules

- Keep domain logic deterministic (no IO)
- Keep adapters behind ports
- Add unit tests for domain and usecases

## Adding docs

Docs live under `docs/` and follow Diataxis framework.

- tutorials: `docs/tutorials/`
- how-to: `docs/how-to/`
- explanation: `docs/explanation/`
- reference: `docs/reference/`

## Read next

- Contributing style guide: [`contributing/style.md`](style.md)
- Releasing: [`contributing/releasing.md`](releasing.md)
- Architecture overview: [`explanation/architecture.md`](../explanation/architecture.md)
- Start here: [`start-here.md`](../start-here.md)
