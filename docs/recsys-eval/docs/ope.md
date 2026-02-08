---
diataxis: explanation
tags:
  - recsys-eval
---
# Off-policy evaluation (OPE): powerful and easy to misuse
This page explains Off-policy evaluation (OPE): powerful and easy to misuse and how it fits into the RecSys suite.


## Who this is for

Advanced users. Read this before using --mode ope in anything serious.

## What you will get

- What OPE tries to estimate
- What propensities are and why they matter
- When OPE results are trustworthy and when they are not

## The promise

OPE tries to answer:

- "What would have happened if we used a different ranking policy?"

using logs collected from an old policy.

This can save you from running an online experiment.

## The catch

OPE depends on assumptions that are easy to violate:

- correct propensity logging
- overlap between old and new policies (support)
- stable user behavior model

If you violate these, OPE can confidently lie.

## Propensities in plain language

A propensity is the probability that a policy shows an item in a position.

If an item never appears under the logging policy, you cannot reliably estimate
how it would perform under a new policy. This is the "no overlap" problem.

## Diagnostics you should take seriously

- near-zero propensities:

  your estimator variance explodes

- missing target propensities:

  you are not evaluating the policy you think you are

- heavy clipping:

  your result is dominated by a few samples

## A practical "when to use" checklist

Use OPE when:

- you log propensities correctly
- your new policy is a mild change from the old
- you mainly want directional signal

Do not use OPE when:

- the new policy changes candidate generation dramatically
- you have missing propensity fields
- you care about strict ship/no-ship

## Recommended practice

- Use offline evaluation first.
- Use OPE as an early filter.
- Use A/B or interleaving for the final decision.

## Read next

- Concepts: [Concepts: how to understand recsys-eval](concepts.md)
- Data contracts: [Data contracts: what inputs look like](data_contracts.md)
- Interpreting results: [Interpreting results: how to go from report to decision](interpreting_results.md)
- Security & privacy: [Security and privacy notes](security_privacy.md)
