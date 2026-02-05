
# How-to: Roll back to a previous artifact version

This repo intentionally separates:

- versioned blobs (immutable)
- a small manifest pointer (mutable)

Rollback is therefore a pointer change.

## Local filesystem example

1) Find previous versions:

```bash
ls -1 .out/registry/records/demo/home/popularity | head
ls -1 .out/registry/records/demo/home/cooc | head
```

1) Pick a version record and get its `URI`.

1) Edit the manifest file:

`.out/registry/current/demo/home/manifest.json`

Change `current.popularity` and/or `current.cooc` to point to the older URIs.

## Production guidance

In production, implement a dedicated rollback command in your operator tooling
that:

- validates the target blob exists
- writes an audit record
- swaps the pointer atomically

See `explanation/artifacts-and-versioning.md`.

## Read next

- Roll back artifacts safely: [`how-to/rollback-safely.md`](rollback-safely.md)
- Artifacts and versioning: [`explanation/artifacts-and-versioning.md`](../explanation/artifacts-and-versioning.md)
- Output layout (registry layout): [`reference/output-layout.md`](../reference/output-layout.md)
- Stale artifacts runbook: [`operations/runbooks/stale-artifacts.md`](../operations/runbooks/stale-artifacts.md)
