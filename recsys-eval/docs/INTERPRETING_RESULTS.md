# Interpreting results: how to go from report to decision

## Who this is for
Anyone making ship/hold decisions (engineers, PMs, analysts).

## What you will get
- How to read a report
- How to decide "ship / hold / rollback" without fooling yourself
- What to do when results are unclear

## Step 0: Trust the data before trusting the metrics

Check:
- data_quality: missing fields, duplicates, anomalies
- join integrity: match rates, unexpected drops
- warnings: especially for OPE

If these look bad, stop. Fix logging.

## Step 1: Start with the summary

The report includes a summary for quick scanning:
- mode
- main deltas (baseline vs candidate or control vs candidate)
- whether gates passed

If the summary says "inconclusive", treat it as a real outcome.

## Step 2: Check guardrails

Even if the primary metric improves, do not ship if:
- empty rate regressed
- latency regressed outside budget
- error rate regressed
- a critical segment cliff appears

Guardrails exist because "winning slowly" is losing.

## Step 3: Look at segments as a radar, not a scoreboard

Segments answer:
- Who did this help?
- Who did this hurt?
- Is the impact consistent?

Segments can also create false positives when you slice too much.
Use segments as diagnostics unless you have power to claim segment wins.

## Step 4: Interpreting uncertainty

If you use confidence intervals or p-values:
- wide intervals mean you do not know yet
- small p-values can still happen by chance if you test too many things

"Inconclusive" is not failure. It is a request for more data or a better
experiment design.

## Step 5: A simple decision policy you can adopt

Suggested policy:
- SHIP:
  primary metric improves and guardrails hold and no major segment regressions
- HOLD:
  results are inconclusive or diagnostics warn about data quality
- ROLLBACK:
  primary regresses or guardrails regress or a major segment cliff appears

This maps well to a decision artifact (api/schemas/decision.v1.json).

## What to do when it is unclear

Choose one:
- run longer / collect more samples
- reduce variance (CUPED / better covariates)
- narrow the change (smaller delta)
- use interleaving for ranker comparison
- do offline gating first, then A/B

## Common pitfalls

- Confusing "statistically significant" with "practically important".
- Shipping a win that is isolated to a single surface and breaks another.
- Ignoring SRM warnings in experiments.

Treat the report as a navigation tool, not a trophy.
