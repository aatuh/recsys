
# Learning path: Engineers

## Goals

- Run the pipeline locally
- Understand code structure (ports/adapters/usecases)
- Add a new artifact type safely
- Debug failures

## Read in this order

1) `tutorials/local-quickstart.md`
1) `reference/cli.md`
1) `reference/config.md`
1) `explanation/architecture.md`
1) `how-to/add-artifact-type.md`
1) `contributing/dev-workflow.md`

## Golden rules

- Domain must stay IO-free and deterministic.
- Publishing must be atomic.
- Every step must be safe to retry.

## Read next

- Start here: [`start-here.md`](../start-here.md)
- Local quickstart: [`tutorials/local-quickstart.md`](../tutorials/local-quickstart.md)
- Add artifact type: [`how-to/add-artifact-type.md`](../how-to/add-artifact-type.md)
- Developer workflow: [`contributing/dev-workflow.md`](../contributing/dev-workflow.md)
