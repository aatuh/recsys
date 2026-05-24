# Ecommerce mini dataset

This directory contains a deterministic, synthetic ecommerce dataset for local
commercial proof-kit demos.

The data is hand-authored, non-PII, and safe to commit. User IDs, session IDs,
request IDs, and item IDs are fictional. Product metadata is illustrative and
does not describe real customers or transactions.

## Files

- `catalog.csv`: small product catalog for human-readable context.
- `pipelines/exposure.jsonl`: flat exposure events consumed by
  `recsys-pipelines`.
- `eval/exposures.jsonl`: `recsys-eval` exposure fixtures.
- `eval/outcomes.jsonl`: `recsys-eval` outcome fixtures.
- `eval/assignments.jsonl`: optional experiment assignment fixtures.

## Scenario

The fixture models one `demo` tenant and one `home` recommendation surface. It
is intentionally small enough to inspect by eye while still producing:

- non-empty popularity/co-vis artifacts,
- implicit, content-similarity, and session-sequence artifacts,
- a served recommendation response,
- joinable exposure/outcome records,
- a shareable offline evaluation report.
