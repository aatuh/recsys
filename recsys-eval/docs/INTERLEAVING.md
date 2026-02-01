# Interleaving: fast ranker comparison on the same traffic

## Who this is for
Engineers comparing two rankers or weight sets.

## What you will get
- What interleaving measures
- When it is the right tool
- Common mistakes

## What it is

Interleaving mixes two ranked lists (A and B) into one displayed list.
Then it attributes user actions (often clicks) back to A or B.

This can be more sensitive than a full A/B when you only care about ranking.

## What it is not

Interleaving is not a full product KPI decision engine.
It does not account for all downstream effects.
Use it to choose between rankers, then validate with A/B.

## Inputs

- ranker_a results (per request_id)
- ranker_b results (per request_id)
- outcomes (clicks)

Dataset wiring example:
configs/examples/dataset.interleaving.jsonl.yaml

## Output

- A wins / B wins counts
- win rate and tie rate
- a significance estimate

## Common mistakes

- comparing rankers trained on different candidate sets without noting it
- treating interleaving wins as business KPI wins
