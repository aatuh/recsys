# Analytics Docs Index

This guide maps each analytics plan in `/analytics/` to the product or engineering question it answers. Use it when you need to know *which* document explains a given workflow (e.g., “How do we evaluate blend weights?”).

Each bullet below follows the pattern **Document — Question it answers — Highlights**:

- **`analytics/blend_eval.md`** — How do we replay recent user interactions to score different blend configurations offline? – Covers the CLI harness, flags, and config file example.
- **`analytics/blend_ab_plan.md`** — What’s the plan for an online blend-weight experiment? – Outlines goals, metrics, assignment strategy, rollout timeline, and risks.
- **`analytics/diversity_validation.md`** — How do we prove diversity guardrails (MMR/caps) keep variety high? – Provides SQL snippets, Grafana panels, and remediation steps.
- **`analytics/personalization_dashboard.md`** — How do we monitor personalized share, overlap, and related signals in production? – Describes dashboard panels, alert thresholds, and the supporting data pipeline.
- **`analytics/retriever_dashboard.md`** — How do we track retriever coverage, latency, and candidate health? – Lists dashboard components, PromQL queries, and alerting tips.
- **`analytics/cold_start_dashboard.md`** — How do we measure cold-start performance (starter profiles, new-user MRR)? – Explains the dashboard layout, metrics, and guardrail hooks.
- **`analytics/embedding_pipeline.md`** — How do we generate and store product embeddings? – Details source data requirements, batch schedule, storage schema, and retriever integration.
- **`analytics/catalog_metadata_inventory.md`** — Which catalog fields exist today, and what gaps remain? – Summarizes current columns, missing data, and proposed additions.
- **`analytics/catalog_backfill_plan.md`** — How do we backfill catalog metadata and keep it fresh? – Describes snapshot/enrichment steps, daily refresh jobs, and monitoring.
- **`analytics/bandit_exploration_plan.md`** — How do we launch the bandit exploration framework? – Covers policy design, rollout sequence, and success metrics.
- **`analytics/bandit_reward_pipeline.md`** — How are bandit rewards ingested and processed? – Shows reward schema, ETL steps, and how `/v1/bandit/reward` feeds downstream jobs.
- **`analytics/blend_eval_configs.example.yaml`** — What does a blend evaluation config file look like? – Provides a sample YAML structure for the offline harness.

## How to use this index

- Start with the question you need to answer (e.g., “We need to increase diversity without hurting CTR”) and open the corresponding doc.
- Each analytics doc is intentionally focused—don’t expect a full tutorial; pair it with the topical runbooks (`docs/tuning_playbook.md`, `docs/simulations_and_guardrails.md`, etc.).
- Many analytics docs assume you already seeded data and have access to Prometheus/Grafana; see `GETTING_STARTED.md` and `docs/quickstart_http.md` for prerequisites.
