# Onboarding Checklist

Use this guide as a suggested ramp-up path for new teammates. Timelines are indicative—feel free to move faster or slower based on your background. Focus on the persona closest to your role; you do **not** have to complete every item.

---

## Persona A – Backend / Integration Engineer

### Phase 1 – Day 1–2: Understand the surface area

- Read `README.md` to learn what RecSys provides and where the docs live.
- Skim `docs/business_overview.md` for the product/value narrative.
- Read `docs/concepts_and_metrics.md` to learn the core terminology (namespace, guardrails, MMR, etc.).
- Explore `docs/quickstart_http.md` and send the sample ingest + `/v1/recommendations` requests against the hosted environment (no code changes needed).

### Phase 2 – Day 3–4: Wire data and configs

- Follow `GETTING_STARTED.md` to run the local stack, seed data, and hit `/v1/recommendations` on localhost.
- Read `docs/api_reference.md` fully; take notes on key headers (`X-Org-ID`, auth) and error cases that matter for your client(s).
- Review `docs/database_schema.md` to see how catalog, user, and event tables are shaped.
- Browse `docs/env_reference.md` and `docs/configuration.md` to understand how defaults, profiles, and overrides interact.

### Phase 3 – Day 5+: Operate confidently

- Run one pass of `docs/tuning_playbook.md` (reset → seed → tune → guardrails) on a sandbox namespace to understand the workflow.
- Skim `docs/simulations_and_guardrails.md` to learn how CI guardrails work and where evidence is stored.
- Review the relevant sections of `docs/rules_runbook.md` so you can support merchandising or override questions.
- Pair with a teammate to review the integration points you own (SDKs, backend services, admin tooling) and log future improvements in backlog.

---

## Persona B – RecSys / ML Engineer

### Phase 1 – Day 1–2: Context & data

- Complete the Backend/Integration Phase 1 steps (README, Business Overview, Concepts & Metrics).
- Read `docs/overview.md` to understand personas, lifecycle, and architecture at a glance.
- Skim recent tuning or guardrail evidence under `analysis/results/` to see examples of acceptable artifacts.

### Phase 2 – Day 3–5: Tuning & evaluation toolchain

- Work through `GETTING_STARTED.md` to run the stack locally if you haven’t already.
- Deep dive `docs/tuning_playbook.md`; run the harness once with your own parameter sweep and commit the evidence to a throwaway branch.
- Follow `docs/simulations_and_guardrails.md` to build a small fixture and execute `run_simulation.py`, then read the generated report.
- Review `docs/env_reference.md` for knobs you are likely to change and note any unknowns for follow-up.

### Phase 3 – Day 6+: Advanced operations

- Read `docs/rules_runbook.md` and experiment with `/v1/admin/manual_overrides` in a sandbox namespace to observe telemetry changes.
- Examine Prometheus/Grafana dashboards (or their JSON definitions) so you know where guardrail and coverage metrics surface.
- Meet with the product/merch team to understand upcoming experiments and document how ML/tuning efforts support them.

---

## Persona C – Product Manager / Merchandising Lead

### Phase 1 – Day 1–2: Big picture

- Read `README.md` and `docs/business_overview.md` to learn what RecSys does for each surface.
- Study `docs/concepts_and_metrics.md`—focus on guardrail metrics (NDCG, MRR, coverage, long-tail share) so reports feel familiar.
- Skim `docs/overview.md` to understand the lifecycle and who does what.

### Phase 2 – Day 3–4: Rules, audits, and API fluency

- Walk through `docs/quickstart_http.md` so you can recognize the key fields involved in ingests and recommendation calls.
- Read the introductory sections of `docs/rules_runbook.md` to understand precedence, telemetry, and day-to-day monitoring.
- Review `docs/business_overview.md#evidence--auditability` plus `docs/simulations_and_guardrails.md` (high level) to know what evidence to expect in reviews.

### Phase 3 – Day 5+: Collaborate and plan

- Sit with an engineer to watch them run a guardrail or scenario suite so you know how to request or interpret results.
- Note any terminology or workflows that still feel unclear and add questions to the backlog or shared FAQ.
- Align with go-to-market / support teams on how rule changes, guardrail failures, or large catalog updates are communicated.

---

If you discover gaps while following this checklist, jot them down and either update the relevant doc or file an issue—continuous documentation feedback keeps onboarding smooth for the next teammate.
