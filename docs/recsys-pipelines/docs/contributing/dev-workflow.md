---
diataxis: explanation
tags:
  - recsys-pipelines
---
# Developer workflow
This page explains Developer workflow and how it fits into the RecSys suite.


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

- Contributing style guide: [Style guide](style.md)
- Releasing: [Releasing](releasing.md)
- Architecture overview: [Architecture](../explanation/architecture.md)
- Start here: [Start here](../start-here.md)
