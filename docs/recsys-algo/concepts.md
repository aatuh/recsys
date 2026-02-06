# Concepts

`recsys-algo` is built for **determinism, explainability, and safe operational behavior**.

## Signals and blending

The engine can blend multiple signals:

- **Popularity** (top-K)
- **Co-visitation** (users/items seen together)
- **Similarity** (embeddings / collaborative / content / session)

A typical configuration exposes blending weights (often referred to as `BlendAlpha`, `BlendBeta`, `BlendGamma`) to
control each signal's contribution.

See also: **[Ranking & constraints reference](ranking-reference.md)** (implemented signals, knobs, determinism notes).

## Cold start and sparsity

For how to handle new users, new items, and new surfaces (including what works in DB-only mode), see:
[Cold start strategies](../explanation/cold-start-strategies.md).

## Rules and constraints

After scoring, the engine can apply:

- **Merchandising rules**: pin / boost / block
- **Diversity and capping**: MMR-style diversification, brand/category caps
- **Hard limits**: K bounds, exclusions, safety checks

## Explainability

For debugging, audits, and safer rollouts, responses can include:

- **Reasons** (high-level explain blocks)
- **Trace data** (low-level diagnostics suitable for audit logs)

Use explain/trace only when you need it — it can increase payload size and computation.
