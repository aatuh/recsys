# Scaling: large datasets and performance

## Who this is for

Anyone running recsys-eval on real production-sized logs.

## What you will get

- When JSONL is enough and when to move to a warehouse
- How stream mode works and what it requires
- Practical tips to avoid OOM and slow runs

## Reality check

If your logs are gigabytes:

- reading everything into memory is not acceptable
- joining exposures and outcomes is the main cost center

Your goals:

- bounded memory
- stable runtime
- reproducible results

## Data source choices

Small datasets:

- JSONL is fine

Large datasets:

- prefer a warehouse-backed adapter (Postgres, etc.)
- let SQL do joins when possible

## Stream mode (JSONL)

Offline mode can support stream mode for large presorted JSONL inputs:

- merge join by request_id
- requires exposures and outcomes sorted by request_id

If inputs are not sorted, stream mode will not behave correctly.
Note: dataset-level distribution metrics (coverage/novelty/diversity) are not
available in stream mode.

## Practical knobs (recommended)

- Run per tenant and per surface, then aggregate.
- Reduce slice keys until you understand performance.
- Start without bootstrap; add it once the basics work.

## Performance debugging checklist

- Is join match rate unexpectedly low?
- Are there duplicate request_id values?
- Are you reading through network storage instead of local disk?
- Are you outputting huge reports because you enabled too much detail?

When in doubt, reduce scope, confirm correctness, then scale up.
