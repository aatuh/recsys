# Documentation Map by Role & Task

Use this map to find the right docs for your role and what you’re trying to do.

> **Who should read this?** New teammates or partners who want a guided reading list instead of exploring the docs tree by hand.

---

## If you are a…

### Product Manager / Business Stakeholder

- Start with:
  - `docs/recsys_in_plain_language.md` – what RecSys is in plain language.
  - `docs/business_overview.md` – value, risks, and guardrails.
- Then:
  - `docs/overview.md` – how RecSys fits into your lifecycle and who owns what.
  - `docs/concepts_and_metrics.md` – skim the “Core concepts” and “Metrics primer”.

### Backend / Integration Engineer

- Start with:
  - `README.md` – “New here?” path.
  - `docs/zero_to_first_recommendation.md` – narrative Acme walkthrough.
  - `docs/quickstart_http.md` – HTTP quickstart (Part 1 first).
- Then:
  - `docs/object_model_concepts.md` – items/users/events/org/namespaces.
  - `docs/api_reference.md` – endpoint details.
  - `docs/faq_and_troubleshooting.md` – common problems and fixes.

### Dev/Ops / SRE

- Start with the integration path above (to understand the basics).
- Then:
  - `docs/configuration.md` – configuration pipeline and mental model.
  - `docs/env_reference.md` – env knobs and overrides (advanced).
  - `docs/tuning_playbook.md` – tuning workflow (advanced).
  - `docs/simulations_and_guardrails.md` – simulations and guardrails (advanced).

---

## If you want to…

### Integrate against the HTTP API

- `docs/quickstart_http.md` – Part 1 and Part 2.
- `docs/client_examples.md` – Python/Node examples.
- `docs/api_reference.md` – per-endpoint reference.
- `docs/faq_and_troubleshooting.md` – when something doesn’t work.

### Understand the object model and schema

- `docs/object_model_concepts.md` – conceptual view (Week 1).
- `docs/object_model_schema_mapping.md` – how this maps to tables (Week 2+).
- `docs/database_schema.md` – full schema guide.

### Tune ranking & guardrails

- `docs/configuration.md` – conceptual pipeline.
- `docs/tuning_playbook.md` – tuning run workflow.
- `docs/simulations_and_guardrails.md` – simulations and guardrails.
- `docs/env_reference.md` – knobs and overrides.

### Debug errors or odd results

- `docs/faq_and_troubleshooting.md` – symptom-based guide.
- `docs/api_errors_and_limits.md` – status codes and limits.
- `docs/rules_runbook.md` – rules/overrides behavior and telemetry.
- `docs/analysis_scripts_reference.md` – scripts for deeper investigation.

Refer back to this map whenever you’re unsure which doc to read next.

