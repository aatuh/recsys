---
tags:
  - overview
  - quickstart
  - business
  - developer
---

# FAQ (for stakeholders)

## Who this is for

- PMs, engineering leaders, and procurement stakeholders evaluating RecSys
- Anyone who needs “plain English” answers before diving into the technical docs

## What this answers

- What RecSys does (and does not) replace
- What you need to run a pilot
- How you control results and measure lift
- How rollbacks and safety work operationally

## Questions

<details markdown="1">
<summary>Will this replace our search?</summary>

No. Search is **intent-driven** (“I want X”), while recommendations are typically **discovery-driven** (“you might like
Y”). In most products, they work together:

- Search solves “find what I asked for”.
- Recommendations solve “help me decide / help me discover”.

</details>

<details markdown="1">
<summary>How do we control what shows up?</summary>

The suite provides a control plane and policy hooks:

- **Rules**: pin / boost / block by surface and segment.
- **Constraints**: required/forbidden tags and per-tag caps.
- **Allow-lists / exclude lists**: constrain candidate sets per request when needed.

For “why did we show this?”, use explainability options during development (`options.include_reasons` / `options.explain`).
</details>

<details markdown="1">
<summary>How do we measure lift and make shipping decisions?</summary>

You measure recommendations with logs:

- log exposures (what you showed)
- log outcomes (click/conversion)
- join by `request_id`

Then choose an evaluation mode:

- **Offline evaluation**: fast regression gate before shipping.
- **A/B experiments**: business KPI lift and guardrails for shipping decisions.
- **Interleaving / OPE**: advanced options when experiments are slow or hard.

</details>

<details markdown="1">
<summary>What do we need for a pilot?</summary>

Minimum viable pilot (DB-only mode):

- a small **catalog** (items + tags)
- a minimal **popularity signal** (daily rows)
- exposure + outcome logging for evaluation

This is intentionally lightweight so teams can validate the end-to-end loop before adding pipelines and artifact mode.
</details>

<details markdown="1">
<summary>How do rollbacks work?</summary>

Two rollback levers exist:

- **Config/rules rollback**: versioned control-plane documents stored in Postgres.
- **Manifest rollback** (artifact mode): “ship” and “rollback” are manifest pointer updates; artifacts are immutable.

</details>

<details markdown="1">
<summary>What about privacy and compliance?</summary>

RecSys is designed to work with **pseudonymous identifiers**. The service can hash identifiers for exposure logging (do
not log raw PII), and you can configure retention policies for logs.
</details>

## Read next

- Stakeholder overview: [`start-here/what-is-recsys.md`](what-is-recsys.md)
- Pilot plan (4–6 weeks): [`start-here/pilot-plan.md`](pilot-plan.md)
- Security, privacy, compliance: [`start-here/security-privacy-compliance.md`](security-privacy-compliance.md)
- Experimentation model: [`explanation/experimentation-model.md`](../explanation/experimentation-model.md)
