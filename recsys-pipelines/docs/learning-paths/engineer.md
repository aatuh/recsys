
# Learning path: Engineers

## Goals

- Run the pipeline locally
- Understand code structure (ports/adapters/usecases)
- Add a new artifact type safely
- Debug failures

## Read in this order

1) `tutorials/local-quickstart.md`
2) `reference/cli.md`
3) `reference/config.md`
4) `explanation/architecture.md`
5) `how-to/add-artifact-type.md`
6) `contributing/dev-workflow.md`

## Golden rules

- Domain must stay IO-free and deterministic.
- Publishing must be atomic.
- Every step must be safe to retry.
