---
title: "Evaluate before changing recommendation ranking"
description: "How to use exposure and outcome joins, guardrails, and offline reports before shipping recommendation ranking changes."
language: "en"
pubDate: "2026-05-23"
translationKey: "evaluate-before-ranking"
tags: ["evaluation", "ranking", "experiments"]
---

Ranking changes are tempting because they are easy to describe: boost this, diversify that, personalize more. They are
harder to trust because the visible result depends on data quality, request context, constraints, rules, and user
behavior after the response.

The evaluation path should start before the ranking change ships.

## Start with joins

Recommendation evaluation depends on joining what was shown to what happened later. That means the request ID must
survive the full path from recommendation response to exposure log and outcome event.

If join integrity is weak, the report is weak. A positive KPI movement can be an instrumentation artifact.

## Guardrails come before optimism

A useful report separates primary KPI movement from operational guardrails. The ranking change may improve clicks and
still be unsafe if it increases latency, empty recommendations, errors, or warning rates.

Guardrails make the decision easier:

- ship when the KPI clears the agreed bar and guardrails hold
- hold when results are inconclusive or joins are weak
- roll back when KPI or guardrails regress materially

## Keep the decision reproducible

The report should point back to the dataset, schemas, config, rules, algorithm version, and artifact state used in the
evaluation. That makes the decision repeatable, and it gives operators a starting point if the rollout behaves
differently in production.

RecSys is designed around this loop: serve, log, evaluate, decide, and roll back if needed.
