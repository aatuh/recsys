# Troubleshooting: symptom -> cause -> fix

## Who this is for

Anyone stuck. Use this as a quick lookup.

## Report is empty or missing sections

Cause:

- wrong mode
- output path permission issue

Fix:

- verify --mode and config
- write to a writable path

## "unknown schema" in validate

Cause:

- wrong schema name

Fix:

- use exposure.v1, outcome.v1, assignment.v1

## Metrics are all zero

Cause:

- outcomes not joined to exposures
- event types not matching expectations

Fix:

- check request_id alignment
- inspect a few joined records

## Everything looks like a win

Cause:

- you compared the same dataset against itself
- you sliced too much and found random wins

Fix:

- run AA-check or use a known baseline
- reduce slices and focus on primary metrics

## Interleaving says A wins but A/B says B wins

Cause:

- interleaving measures relative ranker preference on the same traffic
- A/B includes broader effects and guardrails

Fix:

- use interleaving to choose between rankers
- use A/B to decide shipping
