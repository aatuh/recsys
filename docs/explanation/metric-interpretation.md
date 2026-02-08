---
diataxis: explanation
tags:
  - evaluation
  - metrics
  - interpretation
---
# Interpreting metrics and reports

This page gives a practical mental model for turning an evaluation report into a ship/hold/rollback decision.

!!! info "Canonical reading order"
    This page is an orientation layer. The detailed metric definitions live in `recsys-eval` docs.

## What a report is (and is not)

A RecSys evaluation report is:

- a **decision artifact** (shareable)
- a **reproducible record** (inputs + versions)
- a **compact summary** of multiple metrics and guardrails

It is not:

- a guarantee of online lift
- a substitute for instrumentation hygiene (joinability)

## A 5-minute interpretation flow

1. **Verify the evaluation is valid**

   - Is the population/window what you expected?
   - Are exposures and outcomes joined by stable `request_id`?

   Start here:

   - Evaluation validity: [Evaluation validity](eval-validity.md)
   - Join logic: [Join logic](../reference/data-contracts/join-logic.md)

2. **Check guardrails first**

   - Did any hard guardrail regress beyond tolerance?
   - If yes, decide "hold" even if the primary metric improves.

3. **Read the primary metric in context**

   - Compare relative deltas, not just absolute.
   - Look for segment-specific regressions (new users, cold start surfaces, long-tail items).

4. **Identify tradeoffs and risks**

   - Are you trading diversity for short-term clicks?
   - Are you increasing concentration on a few items?

5. **Write the decision and follow-ups**

   - Ship / hold / rollback
   - 1–5 bullets explaining why
   - The next experiment or mitigation

## Where detailed metric definitions live

- Metric definitions and theory: [Metrics](../recsys-eval/docs/metrics.md)
- Interpreting results (detailed): [Interpreting results](../recsys-eval/docs/interpreting_results.md)
- Interpretation cheat sheet (fast): [Interpretation cheat sheet](../recsys-eval/docs/workflows/interpretation-cheat-sheet.md)
- Decision playbook: [Decision playbook](../recsys-eval/docs/decision-playbook.md)

## Read next

- Run eval and make ship decisions: [Run eval and ship](../how-to/run-eval-and-ship.md)
- Evaluation reasoning and pitfalls: [Evaluation reasoning and pitfalls](evaluation-reasoning.md)
- Evidence (what outputs look like): [Evidence](../for-businesses/evidence.md)
