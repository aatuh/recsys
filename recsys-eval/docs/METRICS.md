# Metrics: what we measure and why

## Who this is for
Analysts, DS, and engineers who need to interpret the output correctly.

## What you will get
- A practical description of the main metric families
- When each metric is useful
- What each metric can hide

You do not need to love metrics. You just need to not be fooled by them.

## Offline ranking metrics (relevance proxies)

Offline metrics compare the ranked list (exposure) against some notion of
"ground truth" (outcomes). Common examples:

- HitRate@K:
  Did at least one relevant item appear in the top K?

- Precision@K:
  Of the top K items, how many were relevant?

- Recall@K:
  Of all relevant items, how many did we include in top K?

- MAP@K:
  Rewards putting relevant items early, averaged across requests.

- NDCG@K:
  A discounted gain metric: earlier is better; supports graded relevance.

These are great for fast regression gating. They are not the same as business
KPIs. They can disagree with online results.

## Experiment metrics (business-facing)

Online metrics are computed from experiments (control vs candidate). Examples:
- CTR: clicks / exposures
- conversion rate: purchases / exposures
- revenue per exposure: sum(value) / exposures
- downstream engagement proxies (if you log them)

These are closer to what the business cares about. They are noisier.

## Guardrails

Guardrails exist to prevent you from shipping a "win" that breaks the system.

Common guardrails:
- empty recommendation rate: response has zero items
- latency: p95/p99 changes
- error rate: HTTP failures or upstream store failures
- join integrity: if joins break, your metrics are fiction

A typical decision policy:
- ship only if primary improves AND guardrails hold AND no segment cliffs

## Distribution and quality metrics

These answer: "Did we change what we show, even if CTR is stable?"

Examples:
- item coverage: how much of the catalog appears
- long-tail share: are we showing only popular items
- category shift: are we drifting away from desired category mix
- diversity: are top K items too similar

These are especially useful when you care about discovery and fairness.

### Distribution metrics implemented here

These are **proxy** metrics derived from exposures and outcomes (no catalog
metadata required):

- **Coverage@K**: unique items shown in top K across all requests divided by
  unique items seen anywhere in recommendations or outcomes. It answers
  "how much of the observed catalog is exposed in the top slots?"

- **Novelty@K**: average `-log2(popularity)` for items shown in top K, where
  popularity is the global exposure frequency. Higher means "less popular on
  average". This is a proxy for long‑tail exposure.

- **Diversity@K**: normalized entropy of the item distribution in top K
  recommendations across requests. Values near 1.0 mean a wide spread;
  values near 0 mean concentration on a few items.

If you have real catalog metadata (categories, embeddings), you should
compute richer diversity/novelty metrics upstream and feed them as outcomes.

## Common mistakes

- "CTR improved so we are done":
  CTR can increase by getting click-bait-y or repeating popular items.

- "Offline NDCG improved so it must ship":
  Offline evaluation can be biased or too simplified.

- "We looked at 50 segments and found 3 big wins":
  This can be pure chance. Treat segments as diagnostics unless powered.

## Practical recommendations

Start with:
- 1-2 primary metrics
- 2-4 guardrails
- segment slicing limited to the top business cuts (tenant/surface/device)

Add more only after you can run the basics reliably.
