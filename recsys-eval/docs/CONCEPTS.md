# Concepts: how to understand recsys-eval

## Who this is for
Anyone. This is the "map" of the system.

## What you will get
- The core mental model in 5 minutes
- The four workflows and when to use each
- A small glossary so words like "exposure" stop being mysterious

## The mental model in one picture

You log:
- what you showed (exposures)
- what users did (outcomes)
- who was in A vs B (assignments, for experiments)

recsys-eval reads those logs and produces a report and optional decision.

```text
exposures (ranked list shown)
        + outcomes (clicks, purchases, etc.)
        + assignments (control vs candidate)
  -----------------------------------------> recsys-eval
  -----------------------------------------> report.json (+ optional decision.json)
```

## Glossary

- request_id:
  A single recommendation moment. One screen, one call, one "ranked list shown".
  In recsys-eval, request_id is the main join key.

- exposure:
  What you showed to the user for a request_id: the ranked list of items plus
  context (tenant, surface, etc.).

- outcome:
  What the user did after the exposure: click, conversion, revenue, etc.

- assignment:
  Which experiment variant a request/user belongs to (control or candidate).

- segment:
  A slice such as tenant + surface + device. Segments are where hidden problems
  show up. Global averages lie.

- guardrail:
  A metric that must not regress even if a primary metric improves. Typical
  guardrails: latency, errors, empty recommendation rate.

- propensity (OPE only):
  A probability that a policy would show an item in a position. If you do not
  have correct propensities, OPE can confidently produce nonsense.

## The four workflows (pick the right tool)

### 1) Offline evaluation
Question:
- "If we rank differently, does it better match what users later did?"

Inputs:
- exposures + outcomes

Outputs:
- ranking metrics (NDCG@K, Recall@K, MAP@K, etc.)
- segment breakdowns
- optional confidence intervals

When to use:
- before shipping changes
- regression gate in CI

Common pitfalls:
- your join from exposures to outcomes is broken
- your "ground truth" is too sparse or biased

### 2) Experiment analysis (A/B)
Question:
- "In production, did variant B outperform A, and did we stay within guardrails?"

Inputs:
- exposures + outcomes + assignments

Outputs:
- KPI deltas (CTR, conversion, etc.)
- confidence intervals or p-values (depending on config)
- guardrail checks
- optional decision artifact (ship/hold/rollback)

When to use:
- shipping decisions

Common pitfalls:
- SRM (sample ratio mismatch): buckets are not balanced
- too many segments: false positives

### 3) Off-policy evaluation (OPE)
Question:
- "Can we estimate impact from logs without running an experiment?"

Inputs:
- exposures + outcomes + propensities

Outputs:
- IPS/SNIPS/DR estimates and diagnostics
- warnings about variance and missing propensities

When to use:
- directional iteration when A/B is expensive

Common pitfalls:
- missing overlap: the new policy behaves outside the support of the logged one
- near-zero propensities: variance explodes

### 4) Interleaving
Question:
- "Between ranker A and B, which one wins more often on the same traffic?"

Inputs:
- ranker A results + ranker B results + outcomes (often clicks)

Outputs:
- win rates, tie rate, p-value

When to use:
- comparing two rankers or weight sets quickly
- when A/B would be too slow or noisy

Common pitfalls:
- you treat interleaving as a full business KPI replacement (it is not)

## Where this fits in the bigger system

Typical stack:
- recsys-service: serves recs and logs exposures and outcomes
- recsys-pipelines: builds artifacts (popularity, co-occurrence, embeddings)
- recsys-algo: ranks and applies rules
- recsys-eval: measures and decides

recsys-eval is the "truth serum": it turns change claims into evidence.
