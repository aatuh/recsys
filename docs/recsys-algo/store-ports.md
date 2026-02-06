# Store ports

`recsys-algo` follows a **ports-and-adapters** style:

- `model` defines **interfaces** (ports) for data access.
- `algorithm` consumes those ports to produce ranked outputs.

This separation makes the engine testable and lets you plug in different storage/backends (Postgres, Redis, object
store, in-memory, etc.).

## Minimal required ports

At minimum, implement:

- `model.PopularityStore` — provides popularity candidates
- `model.TagStore` — provides item tags used for filtering/diversity/caps

## Optional ports (enable more signals)

Depending on which signals/features you want, implement one or more of:

- `model.ProfileStore` — user profile for personalization
- `model.CooccurrenceStore` / `model.HistoryStore` — co-visitation
- `model.EmbeddingStore` — embedding similarity
- `model.CollaborativeStore` — ALS/CF similarity
- `model.ContentStore` — content similarity (tag overlap)
- `model.SessionStore` — session sequences
- `model.EventStore` — event-based exclusions

If a capability is missing, the engine should treat the signal as unavailable and continue.

## Runtime feature disabling

To disable a feature at runtime (even if the port exists), return `model.ErrFeatureUnavailable` from a method.

## Read next

- End-to-end extension guide: [`how-to/add-signal-end-to-end.md`](../how-to/add-signal-end-to-end.md)
- Ranking & constraints reference: [`recsys-algo/ranking-reference.md`](ranking-reference.md)
